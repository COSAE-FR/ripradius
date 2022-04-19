package local

import (
	"github.com/COSAE-FR/ripradius/pkg/utils"
	"github.com/COSAE-FR/riputils/common"
	"github.com/creasty/defaults"
)

type Configuration struct {
	Interface string `yaml:"interface"`
	IPAddress string `yaml:"-"`
	Port      uint32 `yaml:"port" default:"8812"`
	Token     string `yaml:"token"`
}

func (c *Configuration) Check() error {
	if err := defaults.Set(c); err != nil {
		return err
	}
	if len(c.Interface) == 0 {
		c.Interface = utils.LoopbackInterfaceName
	}
	ifIP, err := common.GetIPForInterface(c.Interface)
	if err != nil {
		return err
	}
	c.IPAddress = ifIP.IP.String()
	if len(c.Token) == 0 {
		c.Token = common.RandomHexString(32)
	}
	return nil
}
