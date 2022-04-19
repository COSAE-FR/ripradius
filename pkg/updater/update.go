package updater

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/COSAE-FR/ripradius/pkg/freeradius"
	"github.com/COSAE-FR/ripradius/pkg/updater/binding"
	"github.com/COSAE-FR/riputils/common"
	"io/ioutil"
	"os"
	"path/filepath"
)

func (s *Server) update() {

}

func (s *Server) fetchUpdate() {

}

func (s *Server) applyUpdate() error {
	cert, err := s.getRemoteCertificate()
	if err != nil {
		cert, err = s.getLocalCertificate()
		if err != nil {
			return err
		}
	} else {
		if err := s.writeCache(*cert); err != nil {
			s.log.Errorf("cannot write new certificates to cache: %s", err)
		}
	}
	cfg := *s.config.Radius
	cfg.CA = cert.CA
	cfg.Certificate = cert.Certificate
	cfg.Key = cert.Key
	return s.configureAndStartRadius(&cfg)
}

func (s *Server) startWithoutRemote() error{
	cert, err := s.getLocalCertificate()
	if err != nil {
		return err
	}
	cfg := *s.config.Radius
	cfg.CA = cert.CA
	cfg.Certificate = cert.Certificate
	cfg.Key = cert.Key
	return s.configureAndStartRadius(&cfg)
}

func (s *Server) configureAndStartRadius(config *freeradius.Configuration) error {
	var err error
	var radius *freeradius.Freeradius
	defer func() {
		// If no error the new configuration is OK, save it
		if err == nil {
			s.config.Radius = config
			s.radius = radius
		}
	}()
	radius, err = freeradius.New(s.log, config)
	if err != nil {
		return err
	}
	if s.radius != nil {
		if err := s.radius.Stop(); err != nil {
			s.log.Errorf("cannot stop freeradius server: %s", err)
		}
	}
	err = radius.Configure()
	if err != nil {
		return err
	}
	return radius.Start()
}

func (s *Server) createCacheDirectory() error {
	if common.IsDirectory(s.config.CacheDir) {
		return nil
	}
	if err := os.MkdirAll(s.config.CacheDir, 0700); err != nil {
		s.log.Errorf("cannot create cache directory %s: %s", s.config.CacheDir, err)
		return err
	}
	return nil
}

func (s *Server) writeCache(certificate binding.RadiusCertificate) error {
	if err := s.createCacheDirectory(); err != nil {
		return err
	}
	target := filepath.Join(s.config.CacheDir, "certificate.json")
	data, err := json.Marshal(&certificate)
	if err != nil {
		s.log.Errorf("cannot marshall certificate: %s", err)
		return err
	}
	return ioutil.WriteFile(target, data, 0600)
}

func (s *Server) readCache() (*binding.RadiusCertificate, error) {
	target := filepath.Join(s.config.CacheDir, "certificate.json")
	if !common.FileExists(target) {
		return nil, fmt.Errorf("no cache file at: %s", target)
	}
	data, err := ioutil.ReadFile(target)
	if err != nil {
		return nil, err
	}
	var certificate *binding.RadiusCertificate
	err = json.Unmarshal(data, certificate)
	return certificate, err
}

func (s *Server) getConfigCertificate() (*binding.RadiusCertificate, error) {
	if s.config.Radius == nil {
		return nil, fmt.Errorf("no Radius configuration")
	}
	if s.config.Radius.Certificate == "" {
		return nil, fmt.Errorf("no certificate in Radius configuration")
	}
	if s.config.Radius.Key == "" {
		return nil, fmt.Errorf("no private key in Radius configuration")
	}
	block, _ := pem.Decode([]byte(s.config.Radius.Certificate))
	if block == nil {
		return nil, fmt.Errorf("cannot decode certificate")
	}
	certObject, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	return &binding.RadiusCertificate{
		SignatureDate: &certObject.NotBefore,
		CA:            s.config.Radius.CA,
		Certificate:   s.config.Radius.Certificate,
		Key:           s.config.Radius.Key,
	}, nil
}

func (s *Server) getLocalCertificate() (*binding.RadiusCertificate, error) {
	cert, err := s.readCache()
	if err == nil {
		return cert, nil
	}
	return s.getConfigCertificate()
}

func (s *Server) getRemoteCertificate() (*binding.RadiusCertificate, error) {
	return s.fetcher.GetRemoteCertificate()
}

