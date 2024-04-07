import { loadCommits } from "../../../../../utils/commits"
import { systemManifest } from "../../../../../utils/manifests"

async function main() {
    console.log(systemManifest({
        manifestsRef: (await loadCommits())[0].commit,
        ingressIP: "",
        ingressDomain: "",
        extraComponents: [
            "https://github.com/metalkast/metalkast//manifests/system/base/ironic/components/insecure",
            "https://github.com/metalkast/metalkast//manifests/system/base/cilium/components/issuers/self-signed",
        ]
    }))
}

main()
