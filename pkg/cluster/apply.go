package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/manifestival/manifestival"
	"github.com/metalkast/metalkast/pkg/kustomize"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type Applier struct {
	client dynamic.Interface
	config *rest.Config
	mapper meta.RESTMapper
	logger logr.Logger
}

func NewApplier(client dynamic.Interface, config *rest.Config, mapper meta.RESTMapper, logger logr.Logger) *Applier {
	return &Applier{
		client: client,
		config: config,
		mapper: mapper,
		logger: logger,
	}
}

func (a *Applier) ApplyManifest(manifest manifestival.Manifest) error {
	resources := manifest.Resources()
	manifestCount := len(resources)
	successCount := 0
	applyList := resources
	if err := wait.PollUntilContextTimeout(context.Background(), time.Second*1, time.Minute*20, true, func(ctx context.Context) (bool, error) {
		retryList := []unstructured.Unstructured{}
		for _, r := range applyList {
			resourceName := strings.TrimPrefix(types.NamespacedName{Name: r.GetName(), Namespace: r.GetNamespace()}.String(), "/")
			a.logger.Info(fmt.Sprintf(
				"[%d/%d] Applying manifest %s",
				successCount+1,
				manifestCount,
				resourceName,
			), "refreshable", true)

			if err := a.applyResource(ctx, &r); err != nil {
				a.logger.Error(err, fmt.Sprintf("failed to apply manifest %s, will retry later...", resourceName))
				retryList = append(retryList, r)
			} else {
				successCount++
			}
		}
		applyList = retryList
		return len(applyList) == 0, nil
	}); err != nil {
		return fmt.Errorf("failed to apply all manifests: %w", err)
	}
	a.logger.Info(fmt.Sprintf("Applied all %d manifests", manifestCount), "refreshable", true)
	return nil
}

func (a *Applier) applyResource(ctx context.Context, obj *unstructured.Unstructured) error {
	if obj.GetKind() == "ClusterRole" && obj.GroupVersionKind().Group == "rbac.authorization.k8s.io" {
		if _, found, _ := unstructured.NestedFieldNoCopy(obj.Object, "aggregationRule"); found {
			unstructured.RemoveNestedField(obj.Object, "rules")
		}
	}

	gvk := obj.GroupVersionKind()
	restMapping, err := a.mapper.RESTMapping(gvk.GroupKind())
	// If mapping fails, it might be because a new CRD was just installed
	// Refresh discovery and try again
	if err != nil {
		a.logger.Info(fmt.Sprintf("REST mapping not found for %s, refreshing discovery and retrying", gvk))
		if err := a.refreshDiscovery(); err != nil {
			a.logger.Error(err, "failed to refresh discovery")
			// Continue anyway - might work with stale cache
		}
		restMapping, err = a.mapper.RESTMapping(gvk.GroupKind())
	}

	if err != nil {
		return fmt.Errorf("failed to map %s to resource: %w", gvk, err)
	}

	namespace := obj.GetNamespace()
	if namespace == "" && restMapping.Scope.Name() != meta.RESTScopeNameRoot {
		namespace = "default"
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal object: %w", err)
	}

	var resourceInterface dynamic.ResourceInterface
	if namespace != "" && restMapping.Scope.Name() != meta.RESTScopeNameRoot {
		resourceInterface = a.client.Resource(restMapping.Resource).Namespace(namespace)
	} else {
		resourceInterface = a.client.Resource(restMapping.Resource)
	}

	_, err = resourceInterface.Patch(ctx, obj.GetName(), types.ApplyPatchType, data,
		metav1.PatchOptions{FieldManager: "metalkast"})
	if err != nil {
		return fmt.Errorf("failed to apply resource: %w", err)
	}

	return nil
}

func (a *Applier) refreshDiscovery() error {
	// Create a fresh discovery client to get updated API resources
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(a.config)
	if err != nil {
		return fmt.Errorf("failed to create discovery client: %w", err)
	}

	// Wrap with memory cache
	cachedDiscovery := memory.NewMemCacheClient(discoveryClient)

	apiGroupResources, err := restmapper.GetAPIGroupResources(cachedDiscovery)
	if err != nil {
		return fmt.Errorf("failed to refresh API group resources: %w", err)
	}

	a.mapper = restmapper.NewDiscoveryRESTMapper(apiGroupResources)
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
