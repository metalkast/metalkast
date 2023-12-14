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
apt-get install install -y ipmitool
```

## Prepare cluster manifests

To use kast, you'll first need to prepare your cluster's manifests.
You can use the [manifests used for development](https://github.com/metalkast/metalkast/tree/main/lab/manifests) as a starting point.

To use metalkast prebuilt Kubernetes cluster images, you can include preconfigured `k8s-cluster-version` ConfigMap.

<<< @/../lab/manifests/cluster/deployments/lab-remote-images/kustomization.yaml{5}

## Configure hosts

Create secret(s) with [`metalkast.io/redfish-urls`](/annotations#metalkast-io-redfish-urls) annotation
set to Redfish URLs of the hosts you want to join the cluster and Redfish login credentials.
Make sure to include the secrets in cluster's manifests.

**Example:**

<<< @/../lab/manifests/cluster/deployments/lab-dev/nodes/secrets.yaml{6-9,11-12}

You can encrypt these secrets with [SOPS][sops]:

```shell
sops \
  --age <age_public_key> \
  --encrypted-regex '^(data|stringData)$' \
  manifests/cluster/deployments/lab-dev/nodes/secrets.yaml
```

To use a different editor (e.g. VSCode):

```shell
export EDITOR='code --wait'
```

## Generate hosts manifests

Generate `BareMetalHosts` manifests and make sure to include them in cluster's manifests.

```shell { name=generate }
kast generate \
  manifests/cluster/deployments/lab-dev/nodes/secrets.yaml \
  manifests/cluster/deployments/lab-dev/nodes/nodes.yaml
```

## Bootstrap the cluster

Finally, run the bootstrap. This can take up to an hour depending on your hardware.

```shell { name=bootstrap }
kast bootstrap \
  manifests/system/deployments/lab-dev \
  manifests/cluster/deployments/lab-dev
```

[sops]: https://github.com/getsops/sops
