package fetcher

import (
	"github.com/COSAE-FR/ripradius/pkg/local/client"
	"github.com/COSAE-FR/ripradius/pkg/updater/binding"
)

type HttpFetcher struct {
	client *client.Client
}

func (f *HttpFetcher) GetRemoteCertificate() (*binding.RadiusCertificate, error) {
	return f.client.GetCertificate()
}

func NewHttpFetcher(client *client.Client) *HttpFetcher {
	return &HttpFetcher{client: client}
}

