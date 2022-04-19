package updater

import (
	"github.com/COSAE-FR/ripradius/pkg/freeradius"
	"github.com/COSAE-FR/ripradius/pkg/local/client"
	"github.com/COSAE-FR/ripradius/pkg/updater/fetcher"
	log "github.com/sirupsen/logrus"
	"time"
)

type Server struct {
	config          *Configuration
	fetcher         fetcher.Fetcher
	radius          *freeradius.Freeradius
	certificateDate *time.Time
	tick            *time.Ticker
	done            chan bool
	manual          chan bool
	log             *log.Entry
}

func New(logger *log.Entry, config *Configuration, client *client.Client) (*Server, error) {
	f := fetcher.NewHttpFetcher(client)
	return &Server{config: config, log: logger.WithField("component", "updater"), fetcher: f}, nil
}

func (s *Server) Start() error {
	if err := s.startWithoutRemote(); err != nil {
		s.log.Errorf("cannot start initial radius server with default certificate")
	}
	go func() {
		for {
			select {
			case <-s.done:
				return
			case <-s.tick.C:
				if err := s.applyUpdate(); err != nil {
					s.log.Errorf("cannot update radius certificate: %s", err)
					if err := s.startWithoutRemote(); err != nil {
						s.log.Errorf("cannot start radius server with default certificate")
					}
				}
			case <-s.manual:
				if err := s.applyUpdate(); err != nil {
					s.log.Errorf("cannot update radius certificate: %s", err)
					if err := s.startWithoutRemote(); err != nil {
						s.log.Errorf("cannot start radius server with default certificate")
					}
				}
			}
		}
	}()
	s.manual <- true
	return nil
}

func (s *Server) Stop() error {
	if s.tick != nil {
		s.tick.Stop()
	}
	if s.done != nil {
		s.done <- true
	}
	if s.radius != nil {
		return s.radius.Stop()
	}
	return nil
}

func (s *Server) Configure() error {
	if s.done != nil {
		_ = s.Stop()
	}
	s.tick = time.NewTicker(s.config.Interval)
	s.done = make(chan bool)
	s.manual = make(chan bool)
	if err := s.createCacheDirectory(); err != nil {
		return err
	}
	return nil
}
