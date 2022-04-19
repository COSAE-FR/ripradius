package freeradius

import (
	"context"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path"
)

type Freeradius struct {
	config *Configuration
	log *log.Entry
	process *exec.Cmd
	kill context.CancelFunc
	output *os.File

}

func New(logger *log.Entry, config *Configuration) (*Freeradius, error) {
	return &Freeradius{
		config: config,
		log:    logger.WithField("component", "freeradius"),
	}, nil
}

func (f *Freeradius) openRadiusLogFile() error {
	if len(f.config.BinaryLog) > 0 && f.output == nil {
		if out, err := os.OpenFile(f.config.BinaryLog, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0640); err != nil {
			return err
		} else {
			f.output = out
		}
	}
	return nil
}

func (f *Freeradius) Configure() error {
	if err := f.prepareConfiguration(); err != nil {
		return err
	}
	return f.openRadiusLogFile()
}

func (f *Freeradius) Start() error {
	f.log.Debug("Starting Freeradius process")
	if f.process != nil {
		if err := f.Stop(); err != nil {
			f.log.Errorf("stopping freeradius service: %s", err)
		}
	}
	if err := f.openRadiusLogFile(); err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, f.config.Binary, getFreeradiusArgs(f.config)...)
	f.log.Tracef("Freeradius arguments: %v", getFreeradiusArgs(f.config))
	if f.output != nil {
		cmd.Stdout = f.output
	} else {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr
	f.kill = cancel
	return cmd.Start()
}

func (f *Freeradius) Stop() error {
	f.log.Debug("Stopping Freeradius process")
	if f.kill != nil {
		f.kill()
	}
	f.process = nil
	f.kill = nil
	if f.output != nil {
		if err := f.output.Close(); err != nil {
			f.log.Errorf("cannot close log file: %s", err)
		}
	}
	if f.config.CleanOnStop {
		configurationBase := path.Join(f.config.RunDirectory, "radius")
		if e := os.RemoveAll(configurationBase); e != nil {
			f.log.Errorf("Cannot remove base configurtion directory %s: %s", configurationBase, e)
		}
	}
	return nil
}

func getFreeradiusArgs(config *Configuration) []string {
	var args = []string{"-d",  path.Join(config.RunDirectory, "radius"), "-f", "-l", "stdout"}
	if config.BinaryDebug {
		args = append(args, "-xx")
	} else {
		args = append(args, "-x")
	}
	return args
}
