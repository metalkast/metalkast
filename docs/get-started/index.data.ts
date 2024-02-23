export default {
    load() {
        return {
            clusterManifestsPath: 'manifests/cluster/deployments/<cluster>',
            systemManifestsPath: 'manifests/system/deployments/<cluster>',
        }
    }
}
