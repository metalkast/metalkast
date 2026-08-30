# Lab

Emulates baremetal nodes in docker to support testing `kast` on a local machine.

**Requirements:**
  * Docker
  * 32GB RAM
  * 60GB storage

## Usage

Run emulated hosts and bootstrap them with `kast bootstrap`. Pass a lab id so multiple lab environments can coexist for the same user. Reusing the same lab id will destroy the previous environment for that id if it exists.
```shell
./run codex-validation-run
```

Connect to the lab with the same id:
```shell
./connect codex-validation-run
```

From there you can connect to bootstrap cluster or target cluster
```shell
# bootstrap cluster
k ctx bootstrap
# target cluster
k ctx target
```
