import { systemManifest } from "../../../../../docs/utils/manifests";

async function main() {
    console.log(systemManifest({
        manifestsRef: "main",
        controlPlaneHostname: "192.168.123.104",
        ingressIP: "192.168.123.105",
        ingressDomain: "192.168.123.105.nip.io"
    }))
}

main()
