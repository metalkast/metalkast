package fake

import (
	"context"

	expect "github.com/Netflix/go-expect"
	"github.com/metalkast/metalkast/pkg/bmc"
)

type FakeIpmiTool struct {
	c *expect.Console
}

func NewFakeIpmiTool(c *expect.Console) FakeIpmiTool {
	return FakeIpmiTool{
		c: c,
	}
}

var _ bmc.IpmiSolClient = &FakeIpmiTool{}

func (t *FakeIpmiTool) Run(ctx context.Context, f func(c *expect.Console) error) error {
	return f(t.c)
}
