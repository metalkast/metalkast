import {
    S3Client,
    ListObjectsV2Command,
    _Object,
} from "@aws-sdk/client-s3";
import _ from 'lodash';
import { loadCommits } from "./commits";

export interface Release {
    kubernetesVersion: string;
    version: string;
    release: string;
    url: string;
    dateFormatted: string;
    date: Date;
}

export async function loadReleases(): Promise<Release[]> {
    return (await loadReleaseConfigs()).map((c) => ({
        release: c.Key!!.split('/')[1],
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
    })
}

async function loadReleaseConfigs(): Promise<_Object[]> {
    if (process.env.CI === "true" || process.env.ACCESS_KEY_ID && process.env.SECRET_ACCESS_KEY && process.env.S3_ENDPOINT) {
        const client = new S3Client({
            credentials: {
                accessKeyId: process.env.ACCESS_KEY_ID!!,
                secretAccessKey: process.env.SECRET_ACCESS_KEY!!,
            },
            endpoint: process.env.S3_ENDPOINT!!,
            region: "eeur"
        });
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

export async function latestRelease(): Promise<Release> {
    const commits = await loadCommits();
    const releases = await loadReleases();
    return commits.reduce<Release | undefined>((v, c) => {
        if (v) {
            return v
        }
        return releases.find(r => r.version === c.abbreviatedCommit)
    }, undefined)!!
}
