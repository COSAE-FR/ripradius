package daemon

import (
	"github.com/COSAE-FR/ripradius/pkg/api/local"
	"github.com/COSAE-FR/riputils/svc"
	"github.com/sirupsen/logrus"
)

type Daemon struct {
	Configuration *Configuration
	Freeradius    svc.Configurable
	Api           *local.Server
	Log           *logrus.Entry
}

func (d *Daemon) Start() error {
	d.Log.Debug("Starting API server")
	if err := d.Api.Start(); err != nil {
		return err
	}
	if d.Freeradius != nil {
		d.Log.Debug("Starting Freeradius service")
		if err := d.Freeradius.Start(); err != nil {
			return err
		}
	}
	d.Log.Trace("Services started")
	return nil
}

func (d *Daemon) Stop() error {
	if d.Freeradius != nil {
		d.Log.Debug("Stopping Freeradius service")
		if err := d.Freeradius.Stop(); err != nil {
			d.Log.Errorf("Error while stopping API server: %s", err)
		}
	}
	d.Log.Debug("Stopping API server")
	err := d.Api.Stop()
	d.Log.Trace("Services stopped")
	return err
}

func (d *Daemon) Configure() error {
	if d.Freeradius != nil {
		if err := d.Freeradius.Configure(); err != nil {
			d.Log.Errorf("Cannot configure Freeradius service: %s", err)
			return err
		}
	}
	if d.Api != nil {
		if err := d.Api.Configure(); err != nil {
			d.Log.Errorf("Cannot configure API service: %s", err)
			return nil
		}
	}
	return nil
}
