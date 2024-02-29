import { _Object } from "@aws-sdk/client-s3";
import _ from 'lodash';

import { clusterManifest } from "./utils/manifests";
import { renderMarkdown } from "./utils/render";
import { loadCommits } from "./utils/commits";
import { loadReleases } from "./utils/releases";

export default {
    async load() {
        const commits = await loadCommits();
        let releases = await loadReleases();

        let content = `# Releases

:::warning
There's currently no Long Term Support (LTS) for any of the releases.
:::
`
            +
            commits.filter(c => releases.some(r => r.version === c.abbreviatedCommit))
                .map((c, i, releaseCommits) =>
                    `## Release ${c.time.toLocaleDateString("en-US", { year: 'numeric', month: 'long', day: 'numeric' })}\n` +
                    `Version: [${c.abbreviatedCommit}](https://github.com/metalkast/metalkast/tree/${c.abbreviatedCommit}/)\n` +
                    ((i < releaseCommits.length - 1) ?
                        (
                            "### Changelog\n"
                            +
                            commits.slice(
                                commits.findIndex(cs => c.abbreviatedCommit === cs.abbreviatedCommit),
                                commits.findIndex(cs => releaseCommits[i + 1].abbreviatedCommit === cs.abbreviatedCommit)
                            )
                                .filter(c => c.include)
                                .map(c => `* [\`${c.abbreviatedCommit}\`](https://github.com/metalkast/metalkast/commit/${c.abbreviatedCommit}) ${c.message}`).join("\n")
                        )
                        : "")
                    +
                    releases.filter(r => r.version === c.abbreviatedCommit).map(r => `

:::details ${r.kubernetesVersion}

Example configuration:

${"```yaml"}
${clusterManifest({ k8sVersion: r.url, manifestsRef: c.abbreviatedCommit })}
${"```"}

:::
`).join("")).join("");
        return renderMarkdown(content);
    }
}
