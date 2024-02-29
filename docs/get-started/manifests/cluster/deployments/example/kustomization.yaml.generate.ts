import { latestRelease } from "../../../../../utils/releases";
import { clusterManifest } from "../../../../../utils/manifests";
import { loadCommits } from "../../../../../utils/commits";

async function main() {
    console.log(clusterManifest({
        k8sVersion: (await latestRelease()).url,
        manifestsRef: (await loadCommits())[0].abbreviatedCommit,
    }))
}

main()
