package daemon

import (
	"fmt"
	"github.com/COSAE-FR/ripradius/pkg/api/local"
	"github.com/COSAE-FR/ripradius/pkg/freeradius"
	"github.com/COSAE-FR/ripradius/pkg/local/cache"
	"github.com/COSAE-FR/ripradius/pkg/local/client"
	"github.com/COSAE-FR/ripradius/pkg/local/token"
	"github.com/COSAE-FR/ripradius/pkg/updater"
	"github.com/COSAE-FR/ripradius/pkg/utils"
	"github.com/COSAE-FR/riputils/common"
	"github.com/COSAE-FR/riputils/common/logging"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Configuration struct {
	*logging.Config `yaml:"logging"`
	Cache           cache.Configuration      `yaml:"cache"`
	Api             local.Configuration      `yaml:"api"`
	Client          client.Configuration     `yaml:"client"`
	Radius          freeradius.Configuration `yaml:"radius"`
	Fetcher         *updater.Configuration   `yaml:"fetcher,omitempty"`
	Log             *logrus.Entry            `yaml:"-"`
	logFileWriter   *os.File
	path            string
}

func (c *Configuration) Check() error {
	if err := c.Cache.Check(); err != nil {
		return err
	}
	if err := c.Api.Check(); err != nil {
		return err
	}
	if c.Radius.ApiHost == "" {
		c.Radius.ApiHost = c.Api.IPAddress
	}
	if c.Radius.ApiPort == 0 {
		c.Radius.ApiPort = c.Api.Port
	}
	if c.Radius.ApiToken == "" {
		c.Radius.ApiToken = c.Api.Token
	}
	if err := c.Radius.Check(); err != nil {
		return err
	}
	if c.Client.Token == "" {
		clientToken, err := token.ComputeToken(c.Radius.Secret)
		if err != nil {
			return err
		}
		c.Client.Token = clientToken
	}
	if err := c.Client.Check(); err != nil {
		return err
	}
	if c.Fetcher != nil {
		if err := c.Fetcher.Check(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Configuration) Read() error {
	if _, err := os.Stat(c.path); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	yamlFile, err := os.Open(c.path)
	if err != nil {
		return err
	}
	defer func() {
		_ = yamlFile.Close()
	}()
	byteValue, err := ioutil.ReadAll(yamlFile)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(byteValue, c)
}

func NewConfiguration(path string) (*Configuration, error) {
	if !common.FileExists(path) {
		return nil, fmt.Errorf("configuration file %s does not exist", path)
	}
	config := &Configuration{
		path: path,
	}
	err := config.Read()
	if err != nil {
		return config, err
	}
	config.Log = config.SetupLog(utils.Name, utils.Version)
	err = config.Check()
	return config, err
}
