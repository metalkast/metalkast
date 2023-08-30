package cluster

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-retryablehttp"
	mfc "github.com/manifestival/controller-runtime-client"
	kastlogr "github.com/metalkast/metalkast/pkg/logr"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clusterctlclient "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Cluster struct {
	kubeCfgPath string
	*Applier
	client.Client
	logger logr.Logger
}

func NewCluster(kubeCfgData []byte, kubeCfgDest string, logger logr.Logger) (*Cluster, error) {
	if err := os.WriteFile(kubeCfgDest, kubeCfgData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write cluster kubeconfig (to destination path %v): %w", kubeCfgDest, err)
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeCfgDest)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cluster config: %w", err)
	}

	kubeClient, err := rest.HTTPClientFor(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster client: %w", err)
	}

	retryClient := retryablehttp.NewClient()
	retryClient.Logger = kastlogr.NewFromLogger(logger.V(1).WithName("retry-client"))
	retryClient.HTTPClient = kubeClient
	retryClient.RetryMax =
		// reach retry max
		int(math.Ceil(math.Sqrt(float64(retryClient.RetryWaitMax/retryClient.RetryWaitMin)))) +
			// wait additional 10 minutes
			int((10*time.Minute)/retryClient.RetryWaitMax)

	kubeControllerClient, err := client.New(config, client.Options{
		HTTPClient: retryClient.StandardClient(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster client: %w", err)
	}

	mc := mfc.NewClient(kubeControllerClient)
	return &Cluster{
		kubeCfgPath: kubeCfgDest,
		Applier:     NewApplier(mc, logger),
		Client:      kubeControllerClient,
		logger:      logger,
	}, nil
}

func (c Cluster) ApplyPaths(paths ...string) error {
	for _, p := range paths {
		if err := c.Applier.ApplyKustomize(p); err != nil {
			return fmt.Errorf("failed to deploy %s: %s", p, err)
		}
	}

	return nil
}

func (c *Cluster) Move(target *Cluster, namespace string) error {
	clusterctlClient, err := clusterctlclient.New("")
	if err != nil {
		return fmt.Errorf("failed to init Cluster API client: %w", err)
	}
	err = clusterctlClient.Move(clusterctlclient.MoveOptions{
		FromKubeconfig: clusterctlclient.Kubeconfig{Path: c.kubeCfgPath},
		ToKubeconfig:   clusterctlclient.Kubeconfig{Path: target.kubeCfgPath},
		Namespace:      namespace,
	})
	if err != nil {
		return fmt.Errorf("failed to move Cluster API objects: %w", err)
	}
	return nil
}

func (c *Cluster) KubeCfgPath() string {
	return c.kubeCfgPath
}
