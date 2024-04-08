import { clusterManifest } from "../../../../../docs/utils/manifests";

async function main() {
    console.log(clusterManifest({
        k8sVersion: "../../configs/version/dev",
        controlPlaneHostname: "192.168.123.104.nip.io",
        controlPlaneIP: "192.168.123.104",
        extraComponents: [
            "../../components/debug",
        ],
        manifestsRef: "main",
    }))
}

main()
