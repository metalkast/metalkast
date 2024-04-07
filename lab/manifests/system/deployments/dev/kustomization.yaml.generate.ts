import { systemManifest } from "../../../../../docs/utils/manifests";

async function main() {
    console.log(systemManifest({
        manifestsRef: "main",
        ingressIP: "192.168.123.105",
        ingressDomain: "192.168.123.105.nip.io",
        extraComponents: [
            "https://github.com/metalkast/metalkast//manifests/system/base/ironic/components/insecure",
            "https://github.com/metalkast/metalkast//manifests/system/base/cilium/components/issuers/self-signed",
        ]
    }))
}

main()
