# Annotations

## `metalkast.io/redfish-urls`

This annotation is specified on `Secret` objects to allow `kast generate` command
to discover and generate `BareMetalHost` objects.

URLs are specified one per line. Secret should contain `username` and `password`
which are credentials to the specified Redfish URLs.

## `metalkast.io/bootstrap-cluster-apply`

`kast bootstrap` applies all the given manifests to two different clusters: bootstrap and target.
Bootstrap cluster is temporary and is started on one of the baremetal nodes. The purpose of this cluster is to initialize Cluster API. Since this cluster is not managed by Cluster API, all the objects are later moved to target cluster and the bootstrap cluster is destroyed.

Set the value of this annotation to `"false"` to any manifests that you don't want to be applied to the bootstrap cluster. Typical scenarios for this would be:
  * Skipping manifests that are not needed
on the bootstrap cluster (e.g. your production applications that don't aid in the bootstrap process)
  * Skipping manifests that are not compatible with bootstrap cluster. One such example is CNI plugin manifests. Bootstrap cluster already has a working CNI so you should annotate your CNI plugin manifests with `metalkast.io/bootstrap-cluster-apply: "false"` to avoid breaking the bootstrap cluster.
