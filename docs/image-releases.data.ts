import {
    S3Client,
    ListObjectsV2Command,
    _Object,
} from "@aws-sdk/client-s3";
import _ from 'lodash';
import { createMarkdownRenderer } from "vitepress";

import path from 'path';
import { exec as execInternal } from 'child_process';
import { promisify } from 'util';
const exec = promisify(execInternal)

async function gitLog(format: string) {
    return (await exec(`git --git-dir ${path.join(__dirname, "..", ".git")} log --pretty="format:${format}"`)).stdout.split("\n")
}

const changelogCommits = (await exec(`git --git-dir ${path.join(__dirname, "..", ".git")} log --pretty="format:%h" -- ${path.join(__dirname, "..", "image-build")}`)).stdout.split("\n")

const commits = _.zipWith(
    await gitLog("%h"),
    await gitLog("%s"),
    await gitLog("%ct"),
    (abbreviatedCommit, message, time) => ({
        abbreviatedCommit,
        message,
        time: new Date(Number(time) * 1000),
        include: changelogCommits.includes(abbreviatedCommit),
    })
)

async function loadReleases(): Promise<_Object[]> {
    const client = (process.env.ACCESS_KEY_ID && process.env.SECRET_ACCESS_KEY && process.env.S3_ENDPOINT) ? new S3Client({
        credentials: {
            accessKeyId: process.env.ACCESS_KEY_ID ?? "",
            secretAccessKey: process.env.SECRET_ACCESS_KEY ?? "",
        },
        endpoint: process.env.S3_ENDPOINT,
        region: "eeur"
    }) : undefined;
    if (client) {
        const command = new ListObjectsV2Command({
            Bucket: "metalkast",
        });

        try {
            let isTruncated = true;

            let contents: _Object[] = [];
            while (isTruncated) {
                const { Contents, IsTruncated, NextContinuationToken } =
                    await client.send(command);
                contents = contents.concat(Contents?.filter((c) => c.Key?.endsWith("config.yaml")) ?? [])
                isTruncated = IsTruncated ?? false;
                command.input.ContinuationToken = NextContinuationToken;
            }
            return contents;
        } catch (err) {
            console.error(err);
            throw err;
        }
    }

    // return dev list otherwise
    return [{
        Key: 'node-images/k8s-v1.28.4-ubuntu-22.04-20230719-amd64-3bffbf1/config.yaml',
        LastModified: new Date("2023-08-29T20:10:57.125Z"),
    },
    {
        Key: 'node-images/k8s-v1.29.1-ubuntu-22.04-20230719-amd64-3bffbf1/config.yaml',
        LastModified: new Date("2024-01-29T20:39:15.481Z"),
    },
    {
        Key: 'node-images/k8s-v1.28.4-ubuntu-22.04-20230719-amd64-c140d49/config.yaml',
        LastModified: new Date("2023-08-21T20:10:57.125Z"),
    },
    {
        Key: 'node-images/k8s-v1.29.1-ubuntu-22.04-20230719-amd64-c140d49/config.yaml',
        LastModified: new Date("2024-01-21T20:39:15.481Z"),
    }]
}

export default {
    async load() {
        let releases = (await loadReleases()).map((c) => ({
            release: c.Key?.split('/')[1],
            url: `https://dl.metalkast.io/${c.Key}`,
            dateFormatted: c.LastModified!!.toLocaleDateString("en-US", { year: 'numeric', month: 'short', day: 'numeric' }),
            date: c.LastModified!!,
            ...(/\/k8s-v(?<kubernetesVersion>\d+\.\d+).*?(?<version>[a-z0-9]+)\/config\.yaml$/.exec(
                c.Key ?? "",
            )?.groups as { kubernetesVersion: string, version: string }),
        })).sort((a, b) => {
            let cmp = b.kubernetesVersion.localeCompare(a.kubernetesVersion)
            if (cmp != 0) {
                return cmp
            }
            return b.date.getTime() - a.date.getTime()
        });

        const config = global.VITEPRESS_CONFIG;
        let renderer = await createMarkdownRenderer(
            config.srcDir,
            config.markdown,
            config.site.base,
            config.logger
        );

        let content = `# Releases

:::warning
There's currently no Long Term Support (LTS) for any of the releases.
:::
`
            +
            commits.filter(c => releases.some(r => r.version === c.abbreviatedCommit))
                .map((c, i, releaseCommits) =>
                    // `## Version [${c.abbreviatedCommit}](https://github.com/metalkast/metalkast/commits/${c.abbreviatedCommit}/)\n` +
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

${"```yaml{5}"}
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ${r.url}

components:
  - https://github.com/metalkast/metalkast//manifests/cluster/base?ref=${c.abbreviatedCommit}
${"```"}

:::
`).join("")).join("");
        return renderer.render(content);
    }
}
