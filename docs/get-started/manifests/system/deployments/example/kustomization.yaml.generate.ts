import { loadCommits } from "../../../../../utils/commits"
import { systemManifest } from "../../../../../utils/manifests"

async function main() {
    console.log(systemManifest({
        manifestsRef: (await loadCommits())[0].abbreviatedCommit,
    }))
}

main()
