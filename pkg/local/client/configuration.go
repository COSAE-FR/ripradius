package client

import (
	"fmt"
	"github.com/COSAE-FR/riputils/common"
	"github.com/creasty/defaults"
	"io/ioutil"
	"strings"
)

type Configuration struct {
	Server          string `yaml:"server"`
	ApiVersion      uint16 `yaml:"api_version" default:"1"`
	Token           string `yaml:"token"`
	CA              string `yaml:"ca"`
	Certificate     string `yaml:"certificate"`
	Key             string `yaml:"key"`
	SourceInterface string `yaml:"source_interface"`
}

func (c *Configuration) Check() error {
	if err := defaults.Set(c); err != nil {
		return err
	}
	if len(c.Server) == 0 {
		return fmt.Errorf("authenticator server URL is mandatory")
	}
	if len(c.CA) > 0 {
		if !strings.Contains(c.CA, "BEGIN CERTIFICATE") {
			if !common.FileExists(c.CA) {
				return fmt.Errorf("ca is not a PEM string nor a valid file")
			}
			if content, err := ioutil.ReadFile(c.CA); err != nil {
				return fmt.Errorf("cannot reqd CA file: %s", err)
			} else {
				c.CA = string(content)
			}
		}
	}
	if len(c.Certificate) > 0 {
		if !strings.Contains(c.Certificate, "BEGIN CERTIFICATE") {
			if !common.FileExists(c.Certificate) {
				return fmt.Errorf("certificate is not a PEM string nor a valid file")
			}
			if content, err := ioutil.ReadFile(c.Certificate); err != nil {
				return fmt.Errorf("cannot reqd certificate file: %s", err)
			} else {
				c.Certificate = string(content)
			}
		}
	}
	if len(c.Key) > 0 {
		if !strings.Contains(c.Key, "BEGIN RSA PRIVATE KEY") {
			if !common.FileExists(c.Key) {
				return fmt.Errorf("key is not a PEM string nor a valid file")
			}
			if content, err := ioutil.ReadFile(c.Key); err != nil {
				return fmt.Errorf("cannot reqd key file: %s", err)
			} else {
				c.Key = string(content)
			}
		}
	}
	return nil
}
