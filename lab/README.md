# Lab

Emulates baremetal nodes in docker to support testing `kast` on a local machine.

**Requirements:**
  * Docker
  * 32GB RAM
  * 60GB storage

## Usage

Run emulated hosts and bootstrap them with `kast bootstrap`. This will destroy the previous environment if it exists.
```shell
./run
```

Connect to the lab
```
./connect
```

From there you can connect to bootstrap cluster or target cluster
```shell
# bootstrap cluster
k ctx bootstrap
# target cluster
k ctx target
```
