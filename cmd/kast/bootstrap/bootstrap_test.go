package bootstrap

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/manifestival/manifestival"
	"github.com/metalkast/metalkast/pkg/cluster"
	"github.com/metalkast/metalkast/pkg/testutil"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestBootstrapClusterConfigFromManifests(t *testing.T) {
	testCases := []struct {
		name      string
		manifests manifestival.Manifest
		want      bootstrapClusterConfig
	}{
		{
			name: "valid",
			want: bootstrapClusterConfig{
				bootstrapNodeOptions: BootstrapNodeOptions{
					RedfishUrl:      "https://192.168.122.101",
					RedfishUsername: "admin",
					RedfishPassword: "password",
					LiveIsoUrl:      "http://192.168.122.1/k8s-cluster-node-1.27.2-ubuntu-22.04-amd64-netboot-live.iso",
				},
				clusterNamespace: "clusters-test-ns",
				bootstrapClusterManifests: testutil.TestManifests(t, manifestival.Reader(strings.NewReader(`
apiVersion: v1
stringData:
  password: password
  username: admin
kind: Secret
metadata:
  name: redfish-creds-k8s-node-1
  namespace: capi-clusters
type: Opaque
---
apiVersion: v1
data:
  password: cGFzc3dvcmQ=
  username: YWRtaW4=
kind: Secret
metadata:
  name: redfish-creds-k8s-node-2
  namespace: capi-clusters
type: Opaque
---
apiVersion: v1
data:
  password: cGFzc3dvcmQ=
  username: YWRtaW4=
kind: Secret
metadata:
  name: redfish-creds-k8s-node-3
  namespace: capi-clusters
type: Opaque
---
apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: k8s-node-2
  namespace: capi-clusters
spec:
  automatedCleaningMode: disabled
  bmc:
    address: redfish-virtualmedia+https://192.168.122.102/redfish/v1/Systems/0189dfec-bf89-43d6-8e82-abaaba21770a
    credentialsName: redfish-creds-k8s-node-2
    disableCertificateVerification: true
  bootMACAddress: 52:54:00:6c:3c:02
  externallyProvisioned: false
  online: true
  rootDeviceHints:
    deviceName: /dev/vda
---
apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: KubeadmControlPlane
metadata:
  name: root
  namespace: clusters-test-ns
spec:
  replicas: 1
  machineTemplate:
    infrastructureRef:
      apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
      kind: Metal3MachineTemplate
      name: root-controlplane
  rolloutStrategy:
    rollingUpdate:
      maxSurge: 1
    type: RollingUpdate
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kept-with-annotation-explicit
  annotations:
    metalkast.io/bootstrap-cluster-apply: "true"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kept-without-annotation-implicit
---
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: Metal3MachineTemplate
metadata:
  name: root-controlplane
  namespace: clusters-test-ns
spec:
  template:
    spec:
      image:
        checksum: http://192.168.122.1/k8s-cluster-node-1.27.2-ubuntu-22.04-amd64.img.sha256sum
        url: http://192.168.122.1/k8s-cluster-node-1.27.2-ubuntu-22.04-amd64.img
`))),
			},
			manifests: testutil.TestManifests(t, manifestival.Reader(strings.NewReader(`
apiVersion: v1
stringData:
  password: password
  username: admin
kind: Secret
metadata:
  name: redfish-creds-k8s-node-1
  namespace: capi-clusters
type: Opaque
---
apiVersion: v1
data:
  password: cGFzc3dvcmQ=
  username: YWRtaW4=
kind: Secret
metadata:
  name: redfish-creds-k8s-node-2
  namespace: capi-clusters
type: Opaque
---
apiVersion: v1
data:
  password: cGFzc3dvcmQ=
  username: YWRtaW4=
kind: Secret
metadata:
  name: redfish-creds-k8s-node-3
  namespace: capi-clusters
type: Opaque
---
apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: k8s-node-1
  namespace: capi-clusters
spec:
  automatedCleaningMode: disabled
  bmc:
    address: redfish-virtualmedia+https://192.168.122.101/redfish/v1/Systems/68d3eeea-5c54-4e24-b43f-d0ff0367db96
    credentialsName: redfish-creds-k8s-node-1
    disableCertificateVerification: true
  bootMACAddress: 52:54:00:6c:3c:01
  externallyProvisioned: false
  online: true
  rootDeviceHints:
    deviceName: /dev/vda
---
apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: k8s-node-2
  namespace: capi-clusters
spec:
  automatedCleaningMode: disabled
  bmc:
    address: redfish-virtualmedia+https://192.168.122.102/redfish/v1/Systems/0189dfec-bf89-43d6-8e82-abaaba21770a
    credentialsName: redfish-creds-k8s-node-2
    disableCertificateVerification: true
  bootMACAddress: 52:54:00:6c:3c:02
  externallyProvisioned: false
  online: true
  rootDeviceHints:
    deviceName: /dev/vda
---
apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: k8s-node-3
  namespace: capi-clusters
spec:
  automatedCleaningMode: disabled
  bmc:
    address: redfish-virtualmedia+https://192.168.122.103/redfish/v1/Systems/23b1b88f-5658-442d-a9e1-242027544e6c
    credentialsName: redfish-creds-k8s-node-3
    disableCertificateVerification: true
  bootMACAddress: 52:54:00:6c:3c:03
  externallyProvisioned: false
  online: true
  rootDeviceHints:
    deviceName: /dev/vda
---
apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: KubeadmControlPlane
metadata:
  name: root
  namespace: clusters-test-ns
spec:
  replicas: 3
  machineTemplate:
    infrastructureRef:
      apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
      kind: Metal3MachineTemplate
      name: root-controlplane
  rolloutStrategy:
    rollingUpdate:
      maxSurge: 0
    type: RollingUpdate
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: bootstrap-cluster-skip
  annotations:
    metalkast.io/bootstrap-cluster-apply: "false"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kept-with-annotation-explicit
  annotations:
    metalkast.io/bootstrap-cluster-apply: "true"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kept-without-annotation-implicit
---
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: Metal3MachineTemplate
metadata:
  name: root-controlplane
  namespace: clusters-test-ns
spec:
  template:
    spec:
      image:
        checksum: http://192.168.122.1/k8s-cluster-node-1.27.2-ubuntu-22.04-amd64.img.sha256sum
        url: http://192.168.122.1/k8s-cluster-node-1.27.2-ubuntu-22.04-amd64.img
---
apiVersion: cluster.x-k8s.io/v1beta1
kind: MachineDeployment
metadata:
  name: root
`,
			))),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bootstrapNodeOptions, err := bootstrapClusterConfigFromManifests(tc.manifests)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, *bootstrapNodeOptions)
		})
	}
}

func TestTargetClusterPreMoveManifests(t *testing.T) {
	input := testutil.TestManifests(t, manifestival.Reader(strings.NewReader(`
apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: k8s-node-3
---
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: root
---
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: Metal3Cluster
metadata:
  name: root
---
apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: KubeadmControlPlane
metadata:
  name: root
---
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: Metal3MachineTemplate
metadata:
  name: root-controlplane
---
apiVersion: cluster.x-k8s.io/v1beta1
kind: MachineDeployment
metadata:
  name: root
---
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: Metal3MachineTemplate
metadata:
  name: root-workers
---
apiVersion: bootstrap.cluster.x-k8s.io/v1beta1
kind: KubeadmConfigTemplate
metadata:
  name: root-workers
---
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: Metal3DataTemplate
metadata:
  name: root-controlplane-template
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: bootstrap-cluster-skip
  annotations:
    metalkast.io/bootstrap-cluster-apply: "false"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kept-with-annotation-explicit
  annotations:
    metalkast.io/bootstrap-cluster-apply: "true"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kept-without-annotation-implicit
`)))
	want := testutil.TestManifests(t, manifestival.Reader(strings.NewReader(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: bootstrap-cluster-skip
  annotations:
    metalkast.io/bootstrap-cluster-apply: "false"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kept-with-annotation-explicit
  annotations:
    metalkast.io/bootstrap-cluster-apply: "true"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kept-without-annotation-implicit
`)))
	got := targetClusterPreMoveManifests(input)
	assert.Equal(t, want, got)
}

func TestGetTargetClusterKubeconfig(t *testing.T) {
	bootstrapNode := Bootstrap{
		pollInterval: time.Nanosecond,
		manifests: testutil.TestManifests(t, manifestival.Reader(strings.NewReader(`
apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: KubeadmControlPlane
metadata:
  name: test-cluster
  namespace: clusters-test
spec:
  replicas: 1
`))),
	}
	wantKubeconfig := "<kubeconfig>"

	wantFailureCount := 2
	gotFailureCount := 0
	bootstrapCluster := cluster.Cluster{
		Client: fake.NewClientBuilder().
			WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					if key.Name == "test-cluster-kubeconfig" && key.Namespace == "clusters-test" {
						if gotFailureCount < wantFailureCount {
							gotFailureCount++
							return fmt.Errorf("not found: %s", key.String())
						}
					}
					return client.Get(ctx, key, obj, opts...)
				},
			}).
			WithObjects(&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster-kubeconfig",
					Namespace: "clusters-test",
				},
				Data: map[string][]byte{
					"value": []byte(wantKubeconfig),
				},
			}).Build(),
	}
	kubeconfig, err := bootstrapNode.getTargetClusterKubeconfig(&bootstrapCluster)
	assert.NoError(t, err)
	assert.Equal(t, wantKubeconfig, string(kubeconfig))
	assert.Equal(t, wantFailureCount, gotFailureCount)
}
