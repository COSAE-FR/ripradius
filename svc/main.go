package main

import (
	"github.com/COSAE-FR/ripradius/pkg/api/local"
	"github.com/COSAE-FR/ripradius/pkg/freeradius"
	"github.com/COSAE-FR/ripradius/pkg/local/cache"
	"github.com/COSAE-FR/ripradius/pkg/local/client"
	"github.com/COSAE-FR/ripradius/pkg/updater"
	"github.com/COSAE-FR/ripradius/pkg/utils"
	"github.com/COSAE-FR/ripradius/svc/daemon"
	"github.com/COSAE-FR/riputils/svc"
	"github.com/COSAE-FR/riputils/svc/shared"
	log "github.com/sirupsen/logrus"
)

func New(logger *log.Entry, cfg shared.Config) (svc.Daemonizer, error) {
	config, err := daemon.NewConfiguration(cfg.Conf)
	if err != nil {
		return nil, err
	}
	dmn := daemon.Daemon{Configuration: config, Log: config.Log}
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
			dmn.Freeradius = fetch
		} else {
			fr, err := freeradius.New(logger, &config.Radius)
			if err != nil {
				logger.Errorf("Cannot create Freeradius service: %s", err)
				return nil, err
			}
			dmn.Freeradius = fr
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
	dmn.Api = srv
	return &dmn, nil
}

func main() {
	svc.StartService(utils.Name, New, svc.WithVersion(utils.Version))
}
