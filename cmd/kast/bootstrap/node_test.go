package bootstrap

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"strings"
	"testing"

	expect "github.com/Netflix/go-expect"
	"github.com/metalkast/metalkast/pkg/bmc"
	"github.com/metalkast/metalkast/pkg/bmc/fake"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

type fakeIpmiTool struct {
	c *expect.Console
}

func TestConfigureSSH(t *testing.T) {
	const consoleOutput = `
[  OK  ] Finished Permit User Sessions.
         Starting Hold until boot process finishes up...
         Starting Terminate Plymouth Boot Screen...
cron.service
systemd-user-sessions.service

Ubuntu 22.04.2 LTS ubuntu ttyS0

ubuntu login: [   17.791447] overlayfs: filesystem on '/var/lib/docker/check-overlayfs-support112685843/upper' not supported as upperdir
bootstrap
Password:
Welcome to Ubuntu 22.04.2 LTS (GNU/Linux 5.15.0-76-generic x86_64)

 * Documentation:  https://help.ubuntu.com
 * Management:     https://landscape.canonical.com
 * Support:        https://ubuntu.com/advantage

 System information disabled due to load higher than 2.0

Expanded Security Maintenance for Applications is not enabled.

72 updates can be applied immediately.
43 of these updates are standard security updates.
To see these additional updates run: apt list --upgradable

Enable ESM Apps to receive additional future security updates.
See https://ubuntu.com/esm or run: sudo pro status

The programs included with the Ubuntu system are free software;
the exact distribution terms for each program are described in the
individual files in /usr/share/doc/*/copyright.

Ubuntu comes with ABSOLUTELY NO WARRANTY, to the extent permitted by
applicable law.

$ sudo ssh-keygen -A && sudo systemctl enable --now ssh
ssh-keygen: generating new host keys: RSA DSA ECDSA ED25519
Synchronizing state of ssh.service with SysV service script with /lib/systemd/systemd-sysv-install.
Executing: /lib/systemd/systemd-sysv-install enable ssh
Created symlink /etc/systemd/system/sshd.service → /lib/systemd/system/ssh.service.
Created symlink /etc/systemd/system/multi-user.target.wants/ssh.service → /lib/systemd/system/ssh.service.
$ mkdir -p ~/.ssh && echo 'ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBINKN3oG/aCA2rgNU6nTv2XS0WNfvZM+LMNeKVgSNIH/ouUWB6ILlfDtPoUuIfJmnhepEdLtS8zuWqg5qExOnA4=" > ~/.ssh/authorized_keys
$ cat /etc/ssh/ssh_host_ecdsa_key.pub
ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBB1MuQWdl85Q//TSrvwrUgeprrEBoASD7VD5qkY25IvSl04eKmCSw2uxwtNw6YImFb3/xu9o1rqijr8sFXJpw28= root@k8s-node-1
$ hostname -I | cut -d' ' -f1
192.168.123.101
$ exit
~.
`
	hostPublicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(
		"ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBB1MuQWdl85Q//TSrvwrUgeprrEBoASD7VD5qkY25IvSl04eKmCSw2uxwtNw6YImFb3/xu9o1rqijr8sFXJpw28= root@k8s-node-1",
	))
	assert.NoError(t, err)
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), strings.NewReader("6rif6asdf6yftai6sditgi6astdgisa6drig6asdgsadg"))
	assert.NoError(t, err)
	want := sshConfig{
		user:          BootstrapNodeLiveCdUsername,
		userAuthKey:   *privateKey,
		hostIP:        "192.168.123.101",
		hostPublicKey: hostPublicKey,
	}

	console, err := expect.NewConsole(expect.WithStdin(strings.NewReader(consoleOutput)))
	assert.Nil(t, err, "failed to create console: %w", err)
	ipmiClient := fake.NewFakeIpmiTool(console)
	node := BootstrapNode{
		bmc: &bmc.BMC{
			IpmiClient: &ipmiClient,
		},
	}

	config, err := node.configureSSH(*privateKey)
	assert.NoError(t, err, "failed to configure ssh: %w", err)
	assert.Equal(t, want, *config)
}
