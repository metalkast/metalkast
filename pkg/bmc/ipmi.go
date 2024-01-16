package bmc

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/Netflix/go-expect"
	"github.com/go-logr/logr"
	kastlogr "github.com/metalkast/metalkast/pkg/logr"
	"golang.org/x/sync/errgroup"
)

type IpmiSolClient interface {
	Run(context.Context, func(c *expect.Console) error) error
}

type ipmiTool struct {
	logger       logr.Logger
	ipmiAddress  string
	ipmiUsername string
	ipmiPassword string
}

var _ IpmiSolClient = &ipmiTool{}

func (t *ipmiTool) Run(ctx context.Context, f func(c *expect.Console) error) error {
	c, err := expect.NewConsole(expect.WithStdout(kastlogr.NewLogWriter(t.logger)), expect.WithDefaultTimeout(10*time.Second))
	if err != nil {
		return fmt.Errorf("failed to configure console: %w", err)
	}

	activateCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	out, err := exec.CommandContext(activateCtx, "ipmitool", "-I", "lanplus", "-H", t.ipmiAddress, "-U", t.ipmiUsername, "-P", t.ipmiPassword, "sol", "deactivate").CombinedOutput()
	cancel()
	if err != nil && !strings.Contains(string(out), "already de-activated") {
		t.logger.Error(err, "failed to deactivate previous IPMI SOL Session")
	}

	g, ctx := errgroup.WithContext(ctx)
	cmd := exec.CommandContext(ctx, "ipmitool", "-I", "lanplus", "-H", t.ipmiAddress, "-U", t.ipmiUsername, "-P", t.ipmiPassword, "sol", "activate")
	cmd.Stdin = c.Tty()
	cmd.Stdout = c.Tty()
	cmd.Stderr = c.Tty()
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start IPMI SOL Session: %w", err)
	}

	if _, err := c.ExpectString("SOL Session operational"); err != nil {
		return fmt.Errorf("failed to start SOL Session")
	}

	g.Go(func() error {
		return cmd.Wait()
	})
	g.Go(func() error {
		return f(c)
	})

	return g.Wait()
}

func newIpmiTool(ipmiAddress, ipmiUsername, ipmiPassword string, logger logr.Logger) (*ipmiTool, error) {
	return &ipmiTool{
		logger:       logger,
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
