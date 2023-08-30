package cluster

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/go-logr/logr"
	"github.com/manifestival/manifestival/fake"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	resmaptest_test "sigs.k8s.io/kustomize/api/testutils/resmaptest"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var depProvider = provider.NewDefaultDepProvider()
var rf = depProvider.GetResourceFactory()

func TestApply(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    resmap.ResMap
		wantErr bool
	}{{
		name:    "single",
		wantErr: false,
		input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: single-configmap
  namespace: default
`,
		want: resmaptest_test.NewRmBuilder(t, rf).
			Add(map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "single-configmap",
					"namespace": "default",
				}}).ResMap(),
	}, {
		name: "multiple-objects",
		input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: pod-config
  namespace: default
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: default
spec:
  containers:
  - name: nginx
    image: nginx:1.14.2
`,
		want: resmaptest_test.NewRmBuilder(t, rf).
			Add(map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "pod-config",
					"namespace": "default",
				}}).
			Add(map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name":      "nginx",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name":  "nginx",
							"image": "nginx:1.14.2",
						},
					},
				}}).ResMap(),
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := fake.Client{
				Stubs: fake.Stubs{
					Create: func(u *unstructured.Unstructured) error {
						rNode, err := yaml.FromMap(u.Object)
						if err != nil {
							return fmt.Errorf("failed to convert unstructured to RNode: %w", err)
						}
						resId := resid.FromRNode(rNode)
						if tc.want == nil {
							t.Errorf("unwanted object: %s", resId)
						} else if err := tc.want.Remove(resId); err != nil {
							t.Errorf("resource %s not found: %s", resId, err)
						}
						return nil
					},
				},
			}

			applier := NewApplier(client, logr.New(ctrllog.NullLogSink{}))
			err := applier.Apply(tc.input)
			if tc.wantErr {
				assert.NotNil(t, err, "wanted error but got nil")
			} else {
				assert.Nil(t, err, "failed to apply all objects: %w", err)
			}
			assert.Equal(t, 0, tc.want.Size(), "leftover objects wanted")
		})
	}
}

func TestKustomizeBuild(t *testing.T) {
	testCases := []struct {
		name      string
		buildPath string
		want      string
		wantErr   bool
		files     fstest.MapFS
	}{
		{
			name:      "no kustomization",
			buildPath: "",
			wantErr:   true,
			files: fstest.MapFS{
				"other/kustomization.yaml": {Data: []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`)},
			},
		}, {
			name:      "alternative extension",
			buildPath: "",
			want: strings.TrimSpace(`
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  name: config
			`),
			files: fstest.MapFS{
				"kustomization.yml": {Data: []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: config
  options:
    disableNameSuffixHash: true
  literals:
  - foo=bar
`)},
			},
		}, {
			name:      "multi layer build",
			buildPath: "overlay",
			want: strings.TrimSpace(`
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  name: config
			`),
			files: fstest.MapFS{
				"base/kustomization.yml": {Data: []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: config
  options:
    disableNameSuffixHash: true
  literals:
  - foo=bar
`)},
				"overlay/Kustomization": {Data: []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- ../base
`)},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testDir := t.TempDir()
			fs.WalkDir(tc.files, ".", func(path string, d fs.DirEntry, err error) error {
				dest := filepath.Join(testDir, path)
				if d.IsDir() {
					os.MkdirAll(dest, 0755)
				} else {
					fileContents, err := fs.ReadFile(tc.files, path)
					assert.NoError(t, err)
					os.WriteFile(dest, fileContents, 0644)
				}
				return nil
			})

			wantResMap, err := resmap.NewFactory(rf).NewResMapFromBytes([]byte(tc.want))
			assert.NoError(t, err)

			client := fake.Client{
				Stubs: fake.Stubs{
					Create: func(u *unstructured.Unstructured) error {
						rNode, err := yaml.FromMap(u.Object)
						if err != nil {
							return fmt.Errorf("failed to convert unstructured to RNode: %w", err)
						}
						resId := resid.FromRNode(rNode)
						if wantResMap == nil {
							t.Errorf("unwanted object: %s", resId)
						} else if err := wantResMap.Remove(resId); err != nil {
							t.Errorf("resource %s not found: %s", resId, err)
						}
						return nil
					},
				},
			}

			applier := NewApplier(client, logr.New(ctrllog.NullLogSink{}))
			err = applier.ApplyKustomize(filepath.Join(testDir, tc.buildPath))
			if tc.wantErr {
				assert.NotNil(t, err, "wanted error but got nil")
			} else {
				assert.Nil(t, err, "failed to apply all objects: %w", err)
			}
			assert.Equal(t, 0, wantResMap.Size(), "leftover objects wanted")
		})
	}
}
