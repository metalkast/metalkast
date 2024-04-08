import { _Object } from "@aws-sdk/client-s3";
import _ from 'lodash';

import path from 'path';
import { exec } from "./exec";

async function gitLog(format: string) {
    return (await exec(`git --git-dir ${path.join(__dirname, "../..", ".git")} log --pretty="format:${format}"`)).stdout.split("\n")
}

export async function loadCommits() {
    const changelogCommits = (await exec(`git --git-dir ${path.join(__dirname, "../..", ".git")} log --pretty="format:%h" -- ${path.join(__dirname, "../..", "image-build")}`)).stdout.split("\n");
    return _.zipWith(
        await gitLog("%h"),
        await gitLog("%H"),
        await gitLog("%s"),
        await gitLog("%ct"),
        (abbreviatedCommit, commit, message, time) => ({
            abbreviatedCommit,
            commit,
            message,
            time: new Date(Number(time) * 1000),
            imageChange: changelogCommits.includes(abbreviatedCommit),
        })
    )
}
