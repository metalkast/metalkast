package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/manifestival/manifestival"
	"github.com/metalkast/metalkast/cmd/kast/log"
	"github.com/spf13/cobra"
	"github.com/stmcginnis/gofish"
	"go.mozilla.org/sops/v3/decrypt"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// generateCmd represents the generate command
var (
	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generates BareMetalHosts manifests from set of Secret credentials",
		Long: `Generates BareMetalHosts manifests from set of Secret credentials.

User should provide a source .yaml file containing set of Secrets, each containing
credentials for a given set of nodes configured in metalkast.io/redfish-urls annotation.

Example:

apiVersion: v1
kind: Secret
metadata:
  name: k8s-nodes
  annotations:
    metalkast.io/redfish-urls: |-
      https://192.168.122.101
	  https://192.168.122.102
	  https://192.168.122.103
stringData:
  username: admin
  password: password
type: Opaque

Usage:

kast generate SRC DEST
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliLogger, err := log.NewLogger(log.LoggerOptions{})
			if err != nil {
				return fmt.Errorf("failed to init logger: %w", err)
			}
			defer (cliLogger.GetSink()).(*log.TeaLogSink).Close()
			log.SetLogger(cliLogger)

			return generateBareMetalHosts(args[0], args[1], generateOptions{})
		},
	}
)

func init() {
	rootCmd.AddCommand(generateCmd)
}

const (
	redfishUrlsAnnotation = "metalkast.io/redfish-urls"

	bareMetalHostSecretUsernameField = "username"
	bareMetalHostSecretPasswordField = "password"
)

type generateOptions struct {
	HTTPClient *http.Client
}

func generateBareMetalHosts(inputPath, outputPath string, options generateOptions) error {
	manifests, err := manifestival.ManifestFrom(manifestival.Path(inputPath))
	if len(manifests.Filter(func(u *unstructured.Unstructured) bool {
		_, isSopsEncrypted := u.Object["sops"]
		return isSopsEncrypted
	}).Resources()) > 0 {
		decrypted, err := decrypt.File(inputPath, "yaml")
		if err != nil {
			return fmt.Errorf("failed to decrypt input: %w", err)
		}
		manifests, err = manifestival.ManifestFrom(manifestival.Reader(bytes.NewReader(decrypted)))
		if err != nil {
			return err
		}
	}
	if err != nil {
		return fmt.Errorf("failed to decrypt input: %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to read source manifests (%v): %w", inputPath, err)
	}

	secrets := manifests.Filter(manifestival.All(
		manifestival.ByGVK(corev1.SchemeGroupVersion.WithKind("Secret")),
		func(u *unstructured.Unstructured) bool {
			_, found := u.GetAnnotations()[redfishUrlsAnnotation]
			return found
		},
	))

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	resultsChannel := make(chan *yaml.RNode)
	var bmhCount int
	for _, s := range secrets.Resources() {
		secret := &corev1.Secret{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(s.Object, secret)
		if err != nil {
			return err
		}
		for i, redfishUrl := range strings.Split(secret.GetAnnotations()[redfishUrlsAnnotation], "\n") {
			if strings.TrimSpace(redfishUrl) == "" {
				continue
			}
			bmhCount++
			outputIndex := bmhCount
			redfishUrl := strings.TrimSuffix(redfishUrl, "/")
			suffix := fmt.Sprintf("-%d", i+1)
			g.Go(func() error {
				bmhRNode, err := generateBareMetalHost(*secret, redfishUrl, suffix, outputIndex, options)
				if err != nil {
					return err
				}
				resultsChannel <- bmhRNode
				return nil
			})
		}
	}

	var results []*yaml.RNode
outer:
	for {
		select {
		case res := <-resultsChannel:
			results = append(results, res)
			if len(results) == bmhCount {
				err := g.Wait()
				if err != nil {
					return err
				}
			}
		case <-ctx.Done():
			if len(results) != bmhCount {
				return fmt.Errorf("some nodes timed out")
			}
			break outer
		}
	}

	f, err := os.OpenFile(outputPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	writer := kio.ByteWriter{
		Writer: f,
		Sort:   true,
	}
	return writer.Write(results)
}

func generateBareMetalHost(secret corev1.Secret, redfishUrl, suffix string, outputIndex int, options generateOptions) (*yaml.RNode, error) {
	username, ok := secret.StringData[bareMetalHostSecretUsernameField]
	if !ok {
		username = string(secret.Data[bareMetalHostSecretUsernameField])
	}
	password, ok := secret.StringData[bareMetalHostSecretPasswordField]
	if !ok {
		password = string(secret.Data[bareMetalHostSecretPasswordField])
	}
	config := gofish.ClientConfig{
		Endpoint: redfishUrl,
		// TODO: (GAL-311) Parametrize
		Insecure:  true,
		Username:  username,
		Password:  password,
		BasicAuth: true,
	}
	if options.HTTPClient != nil {
		config.HTTPClient = options.HTTPClient
	}
	client, err := gofish.Connect(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create redfish client: %w", err)
	}
	defer client.Logout()
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil
	retryClient.HTTPClient = client.HTTPClient
	client.HTTPClient = retryClient.StandardClient()

	systems, err := client.Service.Systems()
	if err != nil {
		return nil, fmt.Errorf("failed to list systems: %w", err)
	}
	if len(systems) == 0 {
		return nil, fmt.Errorf("no systems found")
	}

	provider := "redfish"
	if systems[0].Manufacturer == "Dell Inc." {
		provider = "idrac"
	}
	// TODO: add support for HPE iLO 5

	ethernetInterfaces, err := systems[0].EthernetInterfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to list ethernet interfaces: %w", err)
	}
	if len(ethernetInterfaces) == 0 {
		return nil, fmt.Errorf("no ethernet interfaces found")
	}

	redfishUrlParsed, err := url.Parse(redfishUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redfish url (%v): %w", redfishUrl, err)
	}
	bmh := map[string]interface{}{
		"apiVersion": "metal3.io/v1alpha1",
		"kind":       "BareMetalHost",
		"metadata": map[string]interface{}{
			"name": secret.Name + suffix,
			"annotations": map[string]interface{}{
				kioutil.IndexAnnotation: fmt.Sprint(outputIndex),
			},
		},
		"spec": map[string]interface{}{
			"bmc": map[string]interface{}{
				"address":         fmt.Sprintf("%s-virtualmedia+%s", provider, redfishUrlParsed.JoinPath(systems[0].ODataID)),
				"credentialsName": secret.Name,
			},
			"bootMACAddress": ethernetInterfaces[0].PermanentMACAddress,
			"online":         true,
			"rootDeviceHints": map[string]interface{}{
				"minSizeGigabytes": 10,
			},
		},
	}

	bmhRNode, err := yaml.FromMap(bmh)
	if err != nil {
		return nil, err
	}

	if secret.Namespace != "" {
		bmhRNode.SetNamespace(secret.Namespace)
	}

	return bmhRNode, nil
}
