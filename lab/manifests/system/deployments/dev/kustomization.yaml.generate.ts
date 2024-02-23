import { systemManifest } from "../../../../../docs/utils/manifests";

async function main() {
    console.log(systemManifest({
        manifestsRef: "main",
    }))
}

main()
