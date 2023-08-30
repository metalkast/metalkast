package cluster

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/manifestival/manifestival"
	"github.com/metalkast/metalkast/pkg/kustomize"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

type Applier struct {
	client manifestival.Client
	logger logr.Logger
}

func NewApplier(client manifestival.Client, logger logr.Logger) *Applier {
	return &Applier{
		client: client,
		logger: logger,
	}
}

func (a *Applier) ApplyManifest(manifest manifestival.Manifest) error {
	manifestCount := len(manifest.Resources())
	for i, r := range manifest.Resources() {
		m, err := manifestival.ManifestFrom(
			manifestival.Slice([]unstructured.Unstructured{r}),
			manifestival.UseClient(a.client),
		)
		if err != nil {
			return err
		}
		a.logger.Info(fmt.Sprintf(
			"[%d/%d] Applying manifest %s",
			i+1,
			manifestCount,
			strings.TrimPrefix(types.NamespacedName{Name: r.GetName(), Namespace: r.GetNamespace()}.String(), "/"),
		), "refreshable", true)
		if err := m.Apply(); err != nil {
			a.logger.Error(err, fmt.Sprintf("failed to apply manifest %s", strings.TrimPrefix(types.NamespacedName{Name: r.GetName(), Namespace: r.GetNamespace()}.String(), "/")))
			return err
		}
	}
	a.logger.Info(fmt.Sprintf("Applied all %d manifests", manifestCount), "refreshable", true)
	return nil
}

func (a *Applier) Apply(manifests string) error {
	m, err := manifestival.ManifestFrom(manifestival.Reader(strings.NewReader(manifests)))
	if err != nil {
		return fmt.Errorf("failed to instantiate manifests: %w", err)
	}
	return a.ApplyManifest(m)
}

func (a *Applier) ApplyKustomize(path string) error {
	manifests, err := kustomize.Build(path)
	if err != nil {
		return err
	}

	if err = a.Apply(string(manifests)); err != nil {
		return fmt.Errorf("failed to apply kustomize layer (%s): %w", path, err)
	}

	return nil
}
