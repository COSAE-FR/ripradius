package fetcher

import (
	"github.com/COSAE-FR/ripradius/pkg/updater/binding"
)

type RenewedHandler func(cert *binding.RadiusCertificate) error

type Fetcher interface {
	GetRemoteCertificate() (*binding.RadiusCertificate, error)
}

type UpdateFetcher interface {
	Start() error
	Stop() error
	Configure() error
	AddHandler(key string, h RenewedHandler) string
	RemoveHandler(key string)
}
