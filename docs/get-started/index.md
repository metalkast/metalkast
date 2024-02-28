<script setup>
import { data } from './index.data.ts'
</script>

# Get started

This guide will help you deploying a basic Kubernetes cluster on your baremetal machines.

## Install kast binary

You can install `kast` binary from source with Go toolchain.

```shell
go install github.com/metalkast/metalkast/cmd/kast@latest
```

## Install ipmitool

Install `ipmitool` based on the operating system you're running.

**MacOS**

```shell
brew install ipmitool
```

**Ubuntu**

```shell
apt-get install -y ipmitool
```

## Prepare manifests

To use kast, you'll first need to prepare ClusterAPI and system manifests.

### Cluster manifests

Create a kustomization layer in `{{ data.clusterManifestsPath }}` directory and use the example below as reference for your configuration.

To use metalkast prebuilt Kubernetes cluster images, you can include preconfigured `k8s-cluster-version` ConfigMap like in the example below.
You can find the list of all published releases on [Image Releases](/image-releases) page.


::: code-group

<<< @/get-started/manifests/cluster/deployments/example/kustomization.yaml
<<< @/../lab/manifests/cluster/deployments/dev/k8s-config.yaml

:::

#### Configure hosts

Create secret(s) with [`metalkast.io/redfish-urls`](/annotations#metalkast-io-redfish-urls) annotation
set to Redfish URLs of the hosts you want to join the cluster and Redfish login credentials.
Make sure to include the secrets in cluster's manifests.
The example is encrypted with [sops](/sops).

::: code-group

<<< @/../lab/manifests/cluster/deployments/dev/nodes-secrets.yaml{7-10,12-13}

:::

## Configure system manifests

Create a kustomization layer in `{{ data.systemManifestsPath }}` directory and use the example below are reference for your configuration.

::: code-group

<<< @/get-started/manifests/system/deployments/example/kustomization.yaml
<<< @/../lab/manifests/system/deployments/dev/kube-apiserver-config.yaml
<<< @/../lab/manifests/system/deployments/dev/ingress-config.yaml

:::

## Generate hosts manifests

Generate `BareMetalHosts` manifests and make sure to include them in cluster's manifests.

```shell-vue { name=generate }
kast generate \
  {{ data.clusterManifestsPath }}/nodes-secrets.yaml \
  {{ data.clusterManifestsPath }}/nodes.yaml
```

## Bootstrap the cluster

Finally, run the bootstrap. This can take up to an hour depending on your hardware.

```shell-vue { name=bootstrap }
kast bootstrap \
  {{ data.systemManifestsPath }} \
  {{ data.clusterManifestsPath }}
```
