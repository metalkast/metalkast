# SOPS integration

You can encrypt secrets with [SOPS][sops]:

```shell
sops \
  --age <age_public_key> \
  --encrypted-regex '^(data|stringData)$' \
  nodes-secrets.yaml
```

Setup `.sops.yaml` in the root of your repo to automatically apply options for paths.

```yaml
creation_rules:
  - age: <age_public_key>
    encrypted_regex: '^(data|stringData)$'
    path_regex: ...
```

To use a different editor (e.g. VSCode):

```shell
export EDITOR='code --wait'
```

[sops]: https://github.com/getsops/sops
