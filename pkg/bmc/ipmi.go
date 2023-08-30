package bmc

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"time"

	"github.com/Netflix/go-expect"
	"github.com/go-logr/logr"
	kastlogr "github.com/metalkast/metalkast/pkg/logr"
)

type IpmiSolClient interface {
	Activate() error
	Wait() error
	Console() *expect.Console
}

type ipmiTool struct {
	*exec.Cmd
	c            *expect.Console
	ipmiAddress  string
	ipmiUsername string
	ipmiPassword string
}

var _ IpmiSolClient = &ipmiTool{}

func (t *ipmiTool) Activate() error {
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	// Ignore exit code since deactivate doesn't work inside lab
	exec.CommandContext(ctx, "ipmitool", "-I", "lanplus", "-H", t.ipmiAddress, "-U", t.ipmiUsername, "-P", t.ipmiPassword, "sol", "deactivate").Run()
	cancel()
	cmd := exec.CommandContext(context.TODO(), "ipmitool", "-I", "lanplus", "-H", t.ipmiAddress, "-U", t.ipmiUsername, "-P", t.ipmiPassword, "sol", "activate")
	cmd.Stdin = t.c.Tty()
	cmd.Stdout = t.c.Tty()
	cmd.Stderr = t.c.Tty()
	t.Cmd = cmd
	return cmd.Start()
}

func (t *ipmiTool) Console() *expect.Console {
	return t.c
}

func (t *ipmiTool) Wait() error {
	return errors.Join(t.c.Close(), t.Cmd.Wait())
}

func newIpmiTool(ipmiAddress, ipmiUsername, ipmiPassword string, logger logr.Logger) (*ipmiTool, error) {
	c, err := expect.NewConsole(expect.WithStdout(kastlogr.NewLogWriter(logger)), expect.WithDefaultTimeout(10*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to configure console: %w", err)
	}

	return &ipmiTool{
		c:            c,
		ipmiAddress:  ipmiAddress,
		ipmiUsername: ipmiUsername,
		ipmiPassword: ipmiPassword,
	}, nil
}

type BMC struct {
	IpmiClient    IpmiSolClient
	RedfishClient *RedFish
}

func NewBMC(redfishUrl, username, password string, logger logr.Logger) (*BMC, error) {
	redfishUrlParsed, err := url.Parse(redfishUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get ipmi host from redfish url: %w", err)
	}
	ipmiClient, err := newIpmiTool(redfishUrlParsed.Host, username, password, logger.WithName("ipmi console"))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize IPMI client: %w", err)
	}

	redfishClient, err := NewRedFish(redfishUrl, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redfish client: %w", err)
	}

	return &BMC{
		IpmiClient:    ipmiClient,
		RedfishClient: redfishClient,
	}, nil
}
