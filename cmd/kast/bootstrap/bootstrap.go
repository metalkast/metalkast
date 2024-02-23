package bootstrap

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/manifestival/manifestival"
	bmov1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	metal3v1beta1 "github.com/metal3-io/cluster-api-provider-metal3/api/v1beta1"
	"github.com/metalkast/metalkast/cmd/kast/log"
	"github.com/metalkast/metalkast/pkg/cluster"
	"github.com/metalkast/metalkast/pkg/kustomize"
	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1beta1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Bootstrap struct {
	bootstrapClusterConfig        bootstrapClusterConfig
	manifests                     manifestival.Manifest
	targetClusterPreMoveManifests manifestival.Manifest
	pollInterval                  time.Duration
}

func FromManifests(manifestsPaths []string) (*Bootstrap, error) {
	manifests := manifestival.Manifest{}
	for _, p := range manifestsPaths {
		manifestsYaml, err := kustomize.Build(p)
		if err != nil {
			return nil, err
		}
		m, err := manifestival.ManifestFrom(manifestival.Reader(bytes.NewReader(manifestsYaml)))
		if err != nil {
			return nil, fmt.Errorf("failed to convert kustomize layer (%s) to in-memory manifests: %w", p, err)
		}
		manifests = manifests.Append(m)
	}

	bcc, err := bootstrapClusterConfigFromManifests(manifests)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bootstrap node BMC config from manifests: %w", err)
	}

	return &Bootstrap{
		manifests:                     manifests,
		bootstrapClusterConfig:        *bcc,
		targetClusterPreMoveManifests: targetClusterPreMoveManifests(manifests),
		pollInterval:                  time.Second,
	}, nil
}

type bootstrapClusterConfig struct {
	bootstrapNodeOptions      BootstrapNodeOptions
	bootstrapClusterManifests manifestival.Manifest
	clusterNamespace          string
}

func bootstrapClusterConfigFromManifests(manifests manifestival.Manifest) (*bootstrapClusterConfig, error) {
	var setup bootstrapClusterConfig

	bmhManifests := manifests.Filter(manifestival.ByGVK(bmov1alpha1.GroupVersion.WithKind("BareMetalHost")))
	bmhResources := bmhManifests.Resources()
	if len(bmhResources) == 0 {
		return nil, fmt.Errorf("failed to find any BareMetalHosts")
	}

	bmhManifest, err := manifestival.ManifestFrom(manifestival.Slice(append(bmhManifests.Resources()[:1], bmhManifests.Resources()[2:]...)))
	if err != nil {
		panic(fmt.Errorf("failed to create manifests subset: %w", err))
	}

	kubeadmControlPlane := &kubeadmv1beta1.KubeadmControlPlane{}
	setup.bootstrapClusterManifests, err = manifests.Filter(
		manifestival.Not(manifestival.In(bmhManifest)),
		manifestival.Not(manifestival.ByAnnotation(bootstrapClusterApplyAnnotation, "false")),
	).Transform(func(u *unstructured.Unstructured) error {
		if u.GroupVersionKind() == kubeadmv1beta1.GroupVersion.WithKind("KubeadmControlPlane") {
			unstructured.SetNestedField(u.Object, int64(1), "spec", "replicas")
			setup.clusterNamespace = u.GetNamespace()
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, kubeadmControlPlane)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to transform some manifests: %w", err)
	}

	bmh := &bmov1alpha1.BareMetalHost{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(bmhResources[0].Object, bmh)
	if err != nil {
		return nil, fmt.Errorf("failed to parse BareMetalHost (%s): %w", bmhResources[0].GetName(), err)
	}

	redfishUrlParser := regexp.MustCompile(`http(s)?:\/\/[^\/]*`)
	setup.bootstrapNodeOptions.RedfishUrl = redfishUrlParser.FindString(bmh.Spec.BMC.Address)
	if setup.bootstrapNodeOptions.RedfishUrl == "" {
		return nil, fmt.Errorf("failed to find a BareMetalHost with a valid redfish address")
	}

	secretManifest := manifests.Filter(manifestival.All(
		manifestival.ByGVK(schema.FromAPIVersionAndKind("v1", "Secret")),
		manifestival.ByName(bmh.Spec.BMC.CredentialsName),
		func(u *unstructured.Unstructured) bool {
			return u.GetNamespace() == bmh.Namespace
		},
	)).Resources()
	if len(secretManifest) != 1 {
		return nil, fmt.Errorf("failed to find credentials for BareMetalHost (%s): %w", bmh.GetName(), err)
	}

	secret := corev1.Secret{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(secretManifest[0].Object, &secret)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Secret (%s): %w", secretManifest[0].GetName(), err)
	}

	var ok bool
	setup.bootstrapNodeOptions.RedfishUsername, ok = secret.StringData["username"]
	if !ok {
		setup.bootstrapNodeOptions.RedfishUsername = string(secret.Data["username"])
	}
	setup.bootstrapNodeOptions.RedfishPassword, ok = secret.StringData["password"]
	if !ok {
		setup.bootstrapNodeOptions.RedfishPassword = string(secret.Data["password"])
	}

	kubeadmControlPlaneMachineTemplateManifests := manifests.Filter(manifestival.All(
		manifestival.ByGVK(kubeadmControlPlane.Spec.MachineTemplate.InfrastructureRef.GroupVersionKind()),
		manifestival.ByName(kubeadmControlPlane.Spec.MachineTemplate.InfrastructureRef.Name),
		func(u *unstructured.Unstructured) bool {
			return u.GetNamespace() == kubeadmControlPlane.GetNamespace()
		},
	))
	if l := len(kubeadmControlPlaneMachineTemplateManifests.Resources()); l != 1 {
		return nil, fmt.Errorf("want exactly one machine template for KubeadmControlPlane, but found %d", l)
	}

	kubeadmControlPlaneMachineTemplate := &metal3v1beta1.Metal3MachineTemplate{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(
		kubeadmControlPlaneMachineTemplateManifests.Resources()[0].Object,
		kubeadmControlPlaneMachineTemplate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cluster's Metal3MachineTemplate: %w", err)
	}

	const nodeImgSuffix = ".img"
	if !strings.HasSuffix(kubeadmControlPlaneMachineTemplate.Spec.Template.Spec.Image.URL, nodeImgSuffix) {
		return nil, fmt.Errorf("cluster's node image URL needs to have .img extension")
	}
	setup.bootstrapNodeOptions.LiveIsoUrl = strings.TrimSuffix(
		kubeadmControlPlaneMachineTemplate.Spec.Template.Spec.Image.URL, nodeImgSuffix,
	) + "-netboot-live.iso"

	return &setup, nil
}

func targetClusterPreMoveManifests(manifests manifestival.Manifest) manifestival.Manifest {
	return manifests.Filter(func(u *unstructured.Unstructured) bool {
		excludedGroups := []string{
			kubeadmv1beta1.GroupVersion.Group,
			clusterapiv1beta1.GroupVersion.Group,
			bootstrapv1beta1.GroupVersion.Group,
			bmov1alpha1.GroupVersion.Group,
			metal3v1beta1.GroupVersion.Group,
		}
		return !slices.Contains(excludedGroups, u.GroupVersionKind().Group)
	})
}

type BootstrapOptions struct {
	BootstrapNodeOptions
}

func (b *Bootstrap) Run(options BootstrapOptions) error {
	if kubecfgDestPath := options.BootstrapNodeOptions.KubeCfgDestPath; kubecfgDestPath != "" {
		b.bootstrapClusterConfig.bootstrapNodeOptions.KubeCfgDestPath = kubecfgDestPath
	}
	bootstrapNode, err := NewBootstrapNode(b.bootstrapClusterConfig.bootstrapNodeOptions)
	if err != nil {
		return fmt.Errorf("failed to init bootstrap node: %w", err)
	}

	log.Log.Info("Provisioning bootstrap cluster")
	bootstrapCluster, err := bootstrapNode.BootstrapCluster()
	if err != nil {
		return fmt.Errorf("failed to provision target cluster: %w", err)
	}

	log.Log.Info("Applying manifests to bootstrap cluster")
	if err := bootstrapCluster.ApplyManifest(b.bootstrapClusterConfig.bootstrapClusterManifests); err != nil {
		return fmt.Errorf("failed to apply manifests: %w", err)
	}

	targetClusterKubeconfig, err := b.getTargetClusterKubeconfig(bootstrapCluster)
	if err != nil {
		return fmt.Errorf("failed to get target cluster kubeconfig: %w", err)
	}

	targetCluster, err := cluster.NewCluster(
		targetClusterKubeconfig,
		path.Join(path.Dir(bootstrapNode.kubeCfgDest), "target.kubeconfig"),
		log.Log.V(1).WithName("target cluster"),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize target cluster client: %w", err)
	}

	log.Log.Info("Creating target cluster")
	timeout := time.Hour
	var lastEventTimestamp time.Time
	err = wait.PollUntilContextTimeout(context.TODO(), time.Second*1, timeout, true, func(ctx context.Context) (bool, error) {
		events := &corev1.EventList{}
		if err := bootstrapCluster.List(ctx, events, &client.ListOptions{Namespace: b.bootstrapClusterConfig.clusterNamespace}); err != nil {
			log.Log.V(1).Error(err, "failed to list events")
			return false, nil
		}
		slices.SortStableFunc(events.Items, func(a, b corev1.Event) int {
			return a.LastTimestamp.Compare(b.LastTimestamp.Time)
		})
		for _, e := range events.Items {
			if e.LastTimestamp.After(lastEventTimestamp) {
				log.Log.V(1).Info(e.Message, "name", e.InvolvedObject.Name, "kind", e.InvolvedObject.Kind)
			}
		}
		if eventCount := len(events.Items); eventCount > 0 {
			lastEventTimestamp = events.Items[eventCount-1].LastTimestamp.Time
		}

		machines := &clusterapiv1beta1.MachineList{}
		if err := bootstrapCluster.List(ctx, machines); err != nil {
			log.Log.V(1).Error(err, "failed to list machines")
			return false, nil
		}
		if len(machines.Items) > 1 {
			return false, fmt.Errorf("expected only single BareMetalHost")
		} else if len(machines.Items) != 1 {
			return false, nil
		}
		return machines.Items[0].Status.Phase == string(clusterapiv1beta1.MachinePhaseRunning), nil
	})
	if err != nil {
		return fmt.Errorf("timed out after %v waiting for target cluster to be created: %w", timeout, err)
	}

	log.Log.Info("Applying initial manifests to target cluster")
	if err := targetCluster.Applier.ApplyManifest(b.targetClusterPreMoveManifests); err != nil {
		return fmt.Errorf("failed to apply pre-move manifests to target cluster: %w", err)
	}

	log.Log.Info("Moving the cluster")
	wait.PollUntilContextCancel(context.TODO(), b.pollInterval, true, func(ctx context.Context) (done bool, err error) {
		err = bootstrapCluster.Move(targetCluster, b.bootstrapClusterConfig.clusterNamespace)
		if err != nil {
			log.Log.V(1).Error(err, "Failed to move the cluster")
			log.Log.V(1).Info("Retrying to move the cluster")
		}
		return err == nil, nil
	})

	log.Log.Info("Applying all manifests to the target cluster")
	if err := targetCluster.Applier.ApplyManifest(b.manifests); err != nil {
		return fmt.Errorf("failed to apply all manifests after cluster pivoting: %w", err)
	}

	log.Log.Info(fmt.Sprintf(
		`Your Kubernetes target cluster has initialized successfully!

To start using your cluster, you need to run the following:

  export KUBECONFIG=%s

You should now commit all the source files to the git repository.`,
		targetCluster.KubeCfgPath()))
	return nil
}

func (b *Bootstrap) getTargetClusterKubeconfig(bootstrapCluster *cluster.Cluster) ([]byte, error) {
	kubeadmControlPlaneResources := b.manifests.Filter(
		manifestival.ByGVK(kubeadmv1beta1.GroupVersion.WithKind("KubeadmControlPlane")),
	).Resources()
	if l := len(kubeadmControlPlaneResources); l != 1 {
		return nil, fmt.Errorf("want exactly one KubeadmControlPlane in manifests but found: %d", l)
	}

	kubeadmControlPlane := kubeadmControlPlaneResources[0]

	kubeconfigSecret := corev1.Secret{}
	err := wait.PollUntilContextTimeout(context.TODO(), b.pollInterval, time.Minute*10, true, func(ctx context.Context) (done bool, err error) {
		return bootstrapCluster.Client.Get(context.TODO(), types.NamespacedName{
			Name:      fmt.Sprintf("%s-kubeconfig", kubeadmControlPlane.GetName()),
			Namespace: kubeadmControlPlane.GetNamespace(),
		}, &kubeconfigSecret) == nil, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get target cluster kubeconfig from secret: %w", err)
	}

	return kubeconfigSecret.Data["value"], nil
}
