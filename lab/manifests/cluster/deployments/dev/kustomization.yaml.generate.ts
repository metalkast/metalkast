import { clusterManifest } from "../../../../../docs/utils/manifests";

async function main() {
    console.log(clusterManifest({
        k8sVersion: "k8s-version.yaml",
        extraCompoents: [
            "../../components/debug",
        ],
        manifestsRef: "main",
    }))
}

main()
