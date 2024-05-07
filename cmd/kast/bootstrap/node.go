package bootstrap

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/Netflix/go-expect"
	"github.com/metalkast/metalkast/cmd/kast/log"
	"github.com/metalkast/metalkast/pkg/bmc"
	"github.com/metalkast/metalkast/pkg/cluster"
	"github.com/metalkast/metalkast/pkg/logr"
	"golang.org/x/crypto/ssh"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	BootstrapNodeLiveCdUsername = "bootstrap"
	BootstrapNodeLiveCdPassword = "bootstrap"
)

type BootstrapNode struct {
	bmc         *bmc.BMC
	liveIsoUrl  string
	kubeCfgDest string
	sshKeyDest  string
}

type BootstrapNodeOptions struct {
	RedfishUrl      string
	RedfishUsername string
	RedfishPassword string
	LiveIsoUrl      string
	KubeCfgDestPath string
	SSHKeyDestPath  string
}

func NewBootstrapNode(options BootstrapNodeOptions) (*BootstrapNode, error) {
	bmc, err := bmc.NewBMC(options.RedfishUrl, options.RedfishUsername, options.RedfishPassword, log.Log.V(1).WithName("bmc"))
	if err != nil {
		return nil, fmt.Errorf("failed to init BMC for bootstrap node: %w", err)
	}

	kubeCfgDest := options.KubeCfgDestPath
	if kubeCfgDest == "" {
		kubeCfgDest = "bootstrap.kubeconfig"
	}

	sshKeyDest := options.SSHKeyDestPath
	if sshKeyDest == "" {
		sshKeyDest = "ssh.key"
	}

	return &BootstrapNode{
		bmc:         bmc,
		liveIsoUrl:  options.LiveIsoUrl,
		kubeCfgDest: kubeCfgDest,
		sshKeyDest:  sshKeyDest,
	}, nil
}

func (n *BootstrapNode) start() error {
	if err := n.bmc.RedfishClient.InsertMedia(n.liveIsoUrl); err != nil {
		return err
	}

	if err := n.bmc.RedfishClient.SetBootMedia(); err != nil {
		return err
	}

	if err := n.bmc.RedfishClient.Boot(); err != nil {
		return err
	}

	return nil
}

type sshConfig struct {
	user          string
	userAuthKey   ecdsa.PrivateKey
	hostIP        string
	hostPublicKey ssh.PublicKey
}

func (c *sshConfig) sshClient() (*ssh.Client, error) {
	signer, err := ssh.NewSignerFromKey(&c.userAuthKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer from a private key: %w", err)
	}
	config := &ssh.ClientConfig{
		User: c.user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.FixedHostKey(c.hostPublicKey),
	}
	sshConn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.hostIP, 22), config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %s", err)
	}

	return sshConn, nil
}

func (n *BootstrapNode) configureSSH(privateKey ecdsa.PrivateKey) (*sshConfig, error) {
	var (
		hostPublicKey ssh.PublicKey
		hostIp        string
	)
	err := wait.PollUntilContextTimeout(context.TODO(), time.Second, 20*time.Minute, true, func(ctx context.Context) (bool, error) {
		err := n.bmc.IpmiClient.Run(ctx, func(c *expect.Console) error {
			if _, err := c.Write([]byte("\000")); err != nil {
				return err
			}
			if _, err := c.SendLine("\n"); err != nil {
				return err
			}
			if _, err := c.ExpectString("login:"); err != nil {
				return err
			}
			if _, err := c.SendLine(BootstrapNodeLiveCdUsername); err != nil {
				return err
			}
			if _, err := c.ExpectString("Password:"); err != nil {
				return err
			}
			if _, err := c.SendLine(BootstrapNodeLiveCdPassword); err != nil {
				return err
			}

			const prompt = "$ "
			if _, err := c.ExpectString(prompt); err != nil {
				return err
			}
			if _, err := c.SendLine("sudo ssh-keygen -A && sudo systemctl enable --now ssh"); err != nil {
				return err
			}
			if _, err := c.ExpectString(prompt); err != nil {
				return err
			}

			publicKey, err := ssh.NewPublicKey(privateKey.Public())
			if err != nil {
				return err
			}
			if _, err := c.SendLine(fmt.Sprintf("mkdir -p ~/.ssh && echo '%s' > ~/.ssh/authorized_keys", ssh.MarshalAuthorizedKey(publicKey))); err != nil {
				return err
			}
			if _, err := c.ExpectString(prompt); err != nil {
				return err
			}

			const printHostPublicKeyCmd = "cat /etc/ssh/ssh_host_ecdsa_key.pub"
			if _, err := c.SendLine(printHostPublicKeyCmd); err != nil {
				return err
			}
			hostPublicKeyOutput, err := c.ExpectString(prompt)
			if err != nil {
				return err
			}
			hostPublicKey, _, _, _, err = ssh.ParseAuthorizedKey([]byte(
				strings.TrimSpace(strings.TrimPrefix(
					strings.TrimSuffix(hostPublicKeyOutput, prompt),
					printHostPublicKeyCmd,
				)),
			))
			if err != nil {
				return err
			}

			// https://unix.stackexchange.com/a/167040
			const printHostIpCmd = "ip route get 1.1.1.1 | grep -oP 'src \\K\\S+'"
			if _, err := c.SendLine(printHostIpCmd); err != nil {
				return err
			}
			hostIpOutput, err := c.ExpectString(prompt)
			if err != nil {
				return err
			}
			hostIp = strings.TrimSpace(strings.TrimPrefix(
				strings.TrimSuffix(hostIpOutput, prompt),
				printHostIpCmd,
			))

			if _, err := c.SendLine("exit"); err != nil {
				return err
			}
			if _, err := c.Send("~."); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			log.Log.V(1).Error(err, "failed to configure ssh via IPMI")
		}
		return err == nil, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to configure ssh access to bootstrap node: %w", err)
	}

	return &sshConfig{
		user:          BootstrapNodeLiveCdUsername,
		userAuthKey:   privateKey,
		hostIP:        hostIp,
		hostPublicKey: hostPublicKey,
	}, nil
}

func initKubeadm(c sshConfig) error {
	sshClient, err := c.sshClient()
	if err != nil {
		return fmt.Errorf("failed to init ssh: %s", err)
	}

	initSession, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %s", err)
	}

	defer initSession.Close()

	kubeadmInitCommandTemplate := `sudo bash -c '
		set -x
		set -eEuo pipefail
		disk=$(lsblk | awk '"'"'/disk/ {print $1; exit}'"'"')
		mkfs.ext4 /dev/$disk -F
		mkdir /tmp/containerd
		mount /dev/$disk /tmp/containerd
		systemctl stop containerd
		cp -r /var/lib/containerd/* /tmp/containerd/
		umount /tmp/containerd
		mount /dev/$disk /var/lib/containerd
		systemctl start containerd
		kubeVersion=$(ctr -n k8s.io i ls | grep -o -P "(?<=kube-apiserver:)v1\.[0-9]+\.[0-9]+")
		# TODO: parse skip phases and pod network cidr from manifests
		kubeadm init --kubernetes-version $kubeVersion --pod-network-cidr 10.244.0.0/16 --control-plane-endpoint {{ .hostname }} --skip-phases=addon/kube-proxy
		export KUBECONFIG=/etc/kubernetes/admin.conf
		kubectl taint nodes --all node-role.kubernetes.io/control-plane:NoSchedule-
		kubectl -n kube-system create configmap cilium-apiserver-endpoint --from-literal=KUBERNETES_SERVICE_HOST={{ .hostname }}
	'`
	tmpl := template.Must(template.New("notImportant").Parse(kubeadmInitCommandTemplate))
	kubeadmInitCommandBuilder := strings.Builder{}

	if err = tmpl.Execute(&kubeadmInitCommandBuilder, map[string]interface{}{
		"hostname": c.hostIP,
	}); err != nil {
		return fmt.Errorf("failed to execute template: %s", err)
	}

	logger := log.Log.V(1).WithName("ssh init cluster")
	initSession.Stderr = logr.NewLogWriter(logger)
	initSession.Stdout = logr.NewLogWriter(logger)

	return initSession.Run(kubeadmInitCommandBuilder.String())
}

func getBootstrapClusterKubeconfig(c sshConfig) ([]byte, error) {
	sshClient, err := c.sshClient()
	if err != nil {
		return nil, fmt.Errorf("failed to init ssh: %s", err)
	}

	readKubeconfigSession, err := sshClient.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %s", err)
	}

	defer readKubeconfigSession.Close()

	kubeconfig, err := readKubeconfigSession.Output("sudo cat /etc/kubernetes/admin.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeconfig: %s", err)
	}

	return kubeconfig, nil
}

func (n *BootstrapNode) BootstrapCluster() (*cluster.Cluster, error) {
	var err error

	if err = n.start(); err != nil {
		return nil, fmt.Errorf("failed to start bootstrap node: %w", err)
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate temporary SSH key to use for bootstrap node: %w", err)
	}

	sshKeyFileContent, err := ssh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize ssh private key: %w", err)
	}

	if err = os.WriteFile(n.sshKeyDest, pem.EncodeToMemory(sshKeyFileContent), 0600); err != nil {
		return nil, fmt.Errorf("failed to save ssh private key: %w", err)
	}

	sshConfig, err := n.configureSSH(*privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to configure bootstrap node ssh via IPMI: %w", err)
	}

	if err = initKubeadm(*sshConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize kubeadm on bootstrap node: %w", err)
	}

	bootstrapClusterKubeconfig, err := getBootstrapClusterKubeconfig(*sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch kubeconfig from bootstrap node cluster: %w", err)
	}

	bootstrapCluster, err := cluster.NewCluster(
		bootstrapClusterKubeconfig,
		n.kubeCfgDest,
		log.Log.V(1).WithName("bootstrap cluster"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cluster setup: %w", err)
	}

	return bootstrapCluster, nil
}
