package main

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
	"github.com/COSAE-FR/riputils/svc"
	"github.com/COSAE-FR/riputils/svc/shared"
	log "github.com/sirupsen/logrus"
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
	Fetcher *updater.Configuration `yaml:"fetcher,omitempty"`
	Log             *log.Entry               `yaml:"-"`
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

type Daemon struct {
	Configuration *Configuration
	Freeradius    svc.Configurable
	Api           *local.Server
	log           *log.Entry
}

func (d *Daemon) Start() error {
	d.log.Debug("Starting API server")
	if err := d.Api.Start(); err != nil {
		return err
	}
	if d.Freeradius != nil {
		d.log.Debug("Starting Freeradius service")
		if err := d.Freeradius.Start(); err != nil {
			return err
		}
	}
	d.log.Trace("Services started")
	return nil
}

func (d *Daemon) Stop() error {
	if d.Freeradius != nil {
		d.log.Debug("Stopping Freeradius service")
		if err := d.Freeradius.Stop(); err != nil {
			d.log.Errorf("Error while stopping API server: %s", err)
		}
	}
	d.log.Debug("Stopping API server")
	err := d.Api.Stop()
	d.log.Trace("Services stopped")
	return err
}

func (d *Daemon) Configure() error {
	if d.Freeradius != nil {
		if err := d.Freeradius.Configure(); err != nil {
			d.log.Errorf("Cannot configure Freeradius service: %s", err)
			return err
		}
	}
	if d.Api != nil {
		if err := d.Api.Configure(); err != nil {
			d.log.Errorf("Cannot configure API service: %s", err)
			return nil
		}
	}
	return nil
}

func New(logger *log.Entry, cfg shared.Config) (svc.Daemonizer, error) {
	config, err := NewConfiguration(cfg.Conf)
	if err != nil {
		return nil, err
	}
	daemon := Daemon{Configuration: config, log: config.Log}
	logger = config.Log.WithField("component", "create_svc")
	clt, err := client.New(config.Client)
	if err != nil {
		logger.Errorf("Cannot create authentication API client: %s", err)
		return nil, err
	}
	if &config.Radius != nil {
		logger.Debug("Freeradius configuration found, configuring")
		if config.Fetcher != nil {
			logger.Debug("Fetcher configuration found, configuring")
			config.Fetcher.Radius = &config.Radius
			fetch, err := updater.New(logger, config.Fetcher, clt)
			if err != nil {
				return nil, err
			}
			daemon.Freeradius = fetch
		} else {
			fr, err := freeradius.New(logger, &config.Radius)
			if err != nil {
				logger.Errorf("Cannot create Freeradius service: %s", err)
				return nil, err
			}
			daemon.Freeradius = fr
		}
	} else {
		logger.Debug("No Freeradius configuration")
	}
	userCache, err := cache.New(logger, &config.Cache)
	if err != nil {
		logger.Errorf("Cannot create user cache: %s", err)
		return nil, err
	}
	srv, err := local.New(logger, &config.Api, userCache, clt)
	if err != nil {
		logger.Errorf("Cannot create API service: %s", err)
		return nil, err
	}
	daemon.Api = srv
	return &daemon, nil
}

func main() {
	svc.StartService(utils.Name, New, svc.WithVersion(utils.Version))
}
