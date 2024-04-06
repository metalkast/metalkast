interface ManifestOptions {
  manifestsRef?: string;
}

function remoteParams(options: ManifestOptions) {
  return options.manifestsRef ? `?ref=${options.manifestsRef}` : ""
}

interface ClusterManifestOptions extends ManifestOptions {
  k8sVersion: string;
  controlPlaneHostname: string;
  controlPlaneIP: string;
  extraCompoents?: string[];
}

export function clusterManifest(options: ClusterManifestOptions) {
  return `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ${options.k8sVersion} # [!code highlight]
  - nodes-secrets.yaml
  - nodes.yaml

configMapGenerator:
  - name: metalkast.io/cluster-config
    options:
      annotations:
        config.kubernetes.io/local-config: "true"
    literals:
      - control_plane_hostname=${options.controlPlaneHostname} # [!code highlight]
      - control_plane_ip=${options.controlPlaneIP} # [!code highlight]

components:
  - https://github.com/metalkast/metalkast//manifests/cluster/base${remoteParams(options)}
  - https://github.com/metalkast/metalkast//manifests/cluster/components/disable-certificate-verification${remoteParams(options)}
${(options.extraCompoents ?? []).map(c => `  - ${c}`).join("\n")}
`.trim()
}

interface SystemManifestOptions extends ManifestOptions {
  ingressIP: string;
  ingressDomain: string;
}

export function systemManifest(options: SystemManifestOptions) {
  return `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

configMapGenerator:
  - name: metalkast.io/system-config
    options:
      annotations:
        config.kubernetes.io/local-config: "true"
    literals:
      - ingress_ip=${options.ingressIP} # [!code highlight]
      - ingress_domain=${options.ingressDomain} # [!code highlight]

components:
  - https://github.com/metalkast/metalkast//manifests/system/base${remoteParams(options)}
  - https://github.com/metalkast/metalkast//manifests/system/base/ironic/components/insecure${remoteParams(options)}
  - https://github.com/metalkast/metalkast//manifests/system/base/cilium/components/issuers/self-signed${remoteParams(options)}
`.trim()
}
