import { clusterManifest } from "../../../../../docs/utils/manifests";

async function main() {
    console.log(clusterManifest({
        k8sVersion: "../../configs/version/dev",
        controlPlaneHostname: "192.168.123.104",
        extraCompoents: [
            "../../components/debug",
        ],
        manifestsRef: "main",
    }))
}

main()
