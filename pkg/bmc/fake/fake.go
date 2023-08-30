package fake

import (
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

func (t *FakeIpmiTool) Activate() error {
	return nil
}

func (t *FakeIpmiTool) Wait() error {
	t.c.Close()
	return nil
}

func (t *FakeIpmiTool) Console() *expect.Console {
	return t.c
}
