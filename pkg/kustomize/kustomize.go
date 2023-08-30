package kustomize

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.mozilla.org/sops/v3/cmd/sops/formats"
	"go.mozilla.org/sops/v3/decrypt"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/build"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func Build(path string) ([]byte, error) {
	options := build.HonorKustomizeFlags(krusty.MakeDefaultOptions(), (&cobra.Command{}).Flags())
	k := krusty.MakeKustomizer(
		options,
	)

	if path == "" {
		path = "."
	}

	m, err := k.Run(sopsDecryptingFs{
		filesys.MakeFsOnDisk(),
	}, path)
	if err != nil {
		return nil, fmt.Errorf("failed to build kustomize layer (%s): %w", path, err)
	}

	yamlManifests, err := m.AsYaml()
	if err != nil {
		return nil, fmt.Errorf("failed converting kustomize build (%s) to yaml: %w", path, err)
	}

	return yamlManifests, nil
}

var _ filesys.FileSystem = sopsDecryptingFs{}

type sopsDecryptingFs struct {
	filesys.FileSystem
}

// ReadFile implements filesys.FileSystem.
func (fs sopsDecryptingFs) ReadFile(path string) ([]byte, error) {
	data, err := fs.FileSystem.ReadFile(path)
	if err != nil {
		return nil, err
	}

	rNode, err := yaml.Parse(string(data))
	if err != nil {
		return data, nil
	}

	rNodeMap, err := rNode.Map()
	if err != nil {
		// Yaml might successfully parse some files as string or similar
		// if it cannot be converted to map (i.e. is not nested), it's not sops encrypted YAML
		return data, nil
	}

	if _, isSopsEncrypted := rNodeMap["sops"]; isSopsEncrypted {
		decrypted, err := decrypt.DataWithFormat(data, formats.Yaml)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt file (%v): %w", path, err)
		}
		return decrypted, nil
	}

	return data, nil
}
