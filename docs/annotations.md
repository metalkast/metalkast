# Annotations

## `metalkast.io/redfish-urls`

This annotation is specified on `Secret` objects to allow `kast generate` command
to discover and generate `BareMetalHost` objects.

URLs are specified one per line. Secret should contain `username` and `password`
which are credentials to the specified Redfish URLs.

## `metalkast.io/bootstrap-cluster-apply`

`kast bootstrap` applies all the given manifests to two different clusters: bootstrap and target.
Bootstrap cluster is temporary and is started on one of the baremetal nodes. The purpose of this cluster is to initialize Cluster API. Since this cluster is not managed by Cluster API, all the objects are later moved to target cluster and the bootstrap cluster is destroyed.

Set the value of this annotation to `"false"` to any manifests that you don't want to be applied to the bootstrap cluster. For example, metalkast uses this internally to ensure that networking config is not broken during bootstrap.
