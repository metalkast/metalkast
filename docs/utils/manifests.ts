interface ManifestOptions {
  manifestsRef?: string;
}

function remoteParams(options: ManifestOptions) {
  return options.manifestsRef ? `?ref=${options.manifestsRef}` : ""
}

interface ClusterManifestOptions extends ManifestOptions {
  k8sVersion: string;
  extraCompoents?: string[];
}

export function clusterManifest(options: ClusterManifestOptions) {
  return `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ${options.k8sVersion} # [!code highlight]
  - k8s-config.yaml
  - nodes-secrets.yaml
  - nodes.yaml

components:
  - https://github.com/metalkast/metalkast//manifests/cluster/base${remoteParams(options)}
  - https://github.com/metalkast/metalkast//manifests/cluster/components/disable-certificate-verification${remoteParams(options)}
${(options.extraCompoents ?? []).map(c => `  - ${c}`).join("\n")}
`.trim()
}

export function systemManifest(options: ManifestOptions) {
  return `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - kube-apiserver-config.yaml
  - ingress-config.yaml

components:
  - https://github.com/metalkast/metalkast//manifests/system/base${remoteParams(options)}
  - https://github.com/metalkast/metalkast//manifests/system/base/ironic/components/insecure${remoteParams(options)}
  - https://github.com/metalkast/metalkast//manifests/system/base/nginx-ingress/components/issuers/self-signed${remoteParams(options)}
`.trim()
}
