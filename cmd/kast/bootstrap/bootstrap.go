package bootstrap

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"path"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/manifestival/manifestival"
	bmov1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	capm3 "github.com/metal3-io/cluster-api-provider-metal3/api/v1beta1"
	"github.com/metalkast/metalkast/cmd/kast/log"
	"github.com/metalkast/metalkast/pkg/cluster"
	"github.com/metalkast/metalkast/pkg/kustomize"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	bootstrapv1beta1 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta1"
	kubeadmv1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"
	clusterv2 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	metal3contractVersion = "v1beta1"
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

	kubeadmControlPlane := &kubeadmv1beta2.KubeadmControlPlane{}
	setup.bootstrapClusterManifests, err = manifests.Filter(
		manifestival.Not(manifestival.In(bmhManifest)),
		manifestival.Not(manifestival.ByAnnotation(bootstrapClusterApplyAnnotation, "false")),
		manifestival.Not(manifestival.ByKind(reflect.TypeOf(clusterv2.MachineDeployment{}).Name())),
	).Transform(func(u *unstructured.Unstructured) error {
		gvk := u.GroupVersionKind()
		if gvk.Kind == "KubeadmControlPlane" && gvk.Group == "controlplane.cluster.x-k8s.io" &&
			gvk.Version == "v1beta2" {
			unstructured.SetNestedField(u.Object, int64(1), "spec", "replicas")
			setup.clusterNamespace = u.GetNamespace()
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, kubeadmControlPlane)
			if err != nil {
				return fmt.Errorf("failed to convert unstructured to KubeadmControlPlane: %w", err)
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
		manifestival.ByGVK(kubeadmControlPlane.Spec.MachineTemplate.Spec.InfrastructureRef.GroupKind().WithVersion(metal3contractVersion)),
		manifestival.ByName(kubeadmControlPlane.Spec.MachineTemplate.Spec.InfrastructureRef.Name),
		func(u *unstructured.Unstructured) bool {
			return u.GetNamespace() == kubeadmControlPlane.GetNamespace()
		},
	))
	if l := len(kubeadmControlPlaneMachineTemplateManifests.Resources()); l != 1 {
		return nil, fmt.Errorf("want exactly one machine template for KubeadmControlPlane, but found %d", l)
	}

	kubeadmControlPlaneMachineTemplate := &capm3.Metal3MachineTemplate{}
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
			kubeadmv1beta2.GroupVersion.Group,
			clusterv2.GroupVersion.Group,
			bootstrapv1beta1.GroupVersion.Group,
			bmov1alpha1.GroupVersion.Group,
			capm3.GroupVersion.Group,
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

	log.Log.Info("Creating target cluster")
	var lastEventTimestamp time.Time
	var targetClusterInitialNodeIP string
	err = wait.PollUntilContextTimeout(context.TODO(), time.Second*1, time.Hour, true, func(ctx context.Context) (bool, error) {
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

		bareMetalHostsLists := &bmov1alpha1.BareMetalHostList{}
		if err := bootstrapCluster.List(ctx, bareMetalHostsLists); err != nil {
			log.Log.V(1).Error(err, "failed to list bareMetalHostsLists")
			return false, nil
		}
		if len(bareMetalHostsLists.Items) > 1 {
			return false, fmt.Errorf("expected only single BareMetalHost")
		} else if len(bareMetalHostsLists.Items) != 1 {
			return false, nil
		}

		if bareMetalHostsLists.Items[0].Status.HardwareDetails != nil {
			for _, nic := range bareMetalHostsLists.Items[0].Status.HardwareDetails.NIC {
				ip := net.ParseIP(nic.IP)
				if ip == nil || ip.To4() == nil {
					continue
				}
				conn, err := net.DialTimeout("tcp", net.JoinHostPort(nic.IP, "6443"), time.Second*2)
				if err != nil {
					continue
				}
				conn.Close()
				targetClusterInitialNodeIP = nic.IP
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("timed out waiting for target cluster to be created: %w", err)
	}

	targetClusterKubeconfig, err := b.getTargetClusterKubeconfig(bootstrapCluster)
	if err != nil {
		return fmt.Errorf("failed to get target cluster kubeconfig: %w", err)
	}

	temporaryTargetClusterKubeconfig, err := kubeconfigWithReplacedHost(targetClusterKubeconfig, targetClusterInitialNodeIP)
	if err != nil {
		return fmt.Errorf("failed to create temporary target cluster kubeconfig: %w", err)
	}

	targetCluster, err := cluster.NewCluster(
		temporaryTargetClusterKubeconfig,
		path.Join(path.Dir(bootstrapNode.kubeCfgDest), "target.kubeconfig"),
		log.Log.V(1).WithName("target cluster"),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize target cluster client: %w", err)
	}

	log.Log.Info("Applying initial manifests to target cluster")
	if err := targetCluster.Applier.ApplyManifest(b.targetClusterPreMoveManifests); err != nil {
		return fmt.Errorf("failed to apply pre-move manifests to target cluster: %w", err)
	}

	log.Log.Info("Waiting for target cluster to finish provisioning")
	if err := wait.PollUntilContextTimeout(context.TODO(), time.Second*1, time.Minute*5, true, func(ctx context.Context) (bool, error) {
		clusterList := &clusterv2.ClusterList{}
		if err := bootstrapCluster.List(ctx, clusterList); err != nil {
			log.Log.V(1).Error(err, "failed to list clusters")
			return false, nil
		}

		clusterCount := len(clusterList.Items)
		if clusterCount == 0 {
			return false, nil
		}

		if clusterCount > 1 {
			return false, fmt.Errorf("expected only one cluster, but found %d", clusterCount)
		}

		cpi := clusterList.Items[0].Status.Initialization.ControlPlaneInitialized
		if cpi == nil {
			return false, nil
		}

		return *cpi, nil
	}); err != nil {
		return fmt.Errorf("timed out waiting for target cluster to be provisioned: %w", err)
	}

	targetCluster, err = cluster.NewCluster(
		targetClusterKubeconfig,
		path.Join(path.Dir(bootstrapNode.kubeCfgDest), "target.kubeconfig"),
		log.Log.V(1).WithName("target cluster"),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize target cluster client: %w", err)
	}

	log.Log.Info("Waiting for CAPI webhook to be ready in target cluster")
	if err := wait.PollUntilContextTimeout(context.TODO(), b.pollInterval, time.Minute*10, true, func(ctx context.Context) (bool, error) {
		endpoints := &corev1.Endpoints{}
		if err := targetCluster.Get(ctx, types.NamespacedName{
			Name:      "capi-webhook-service",
			Namespace: "capi-system",
		}, endpoints); err != nil {
			log.Log.V(1).Error(err, "failed to get CAPI webhook service endpoints")
			return false, nil
		}
		for _, subset := range endpoints.Subsets {
			if len(subset.Addresses) > 0 {
				return true, nil
			}
		}
		return false, nil
	}); err != nil {
		return fmt.Errorf("timed out waiting for CAPI webhook to be ready in target cluster: %w", err)
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
		manifestival.ByGVK(kubeadmv1beta2.GroupVersion.WithKind("KubeadmControlPlane")),
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

func kubeconfigWithReplacedHost(kubeconfigContentInput []byte, newHost string) ([]byte, error) {
	kubeconfig, err := clientcmd.Load(kubeconfigContentInput)
	if err != nil {
		return nil, err
	}
	clusters := maps.Keys(kubeconfig.Clusters)
	if len(clusters) != 1 {
		return nil, fmt.Errorf("expected single cluster in kubeconfig, got %v", len(clusters))
	}
	kubeconfig.Clusters[clusters[0]].Server = fmt.Sprintf("https://%s:6443", newHost)

	kubeconfigContentResult, err := clientcmd.Write(*kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to render kubeconfig: %w", err)
	}
	return kubeconfigContentResult, nil
}
