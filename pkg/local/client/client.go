package client

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/COSAE-FR/ripradius/pkg/api/binding"
	ubinding "github.com/COSAE-FR/ripradius/pkg/updater/binding"
	"github.com/COSAE-FR/ripradius/pkg/utils"
	"github.com/COSAE-FR/riputils/common"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/url"
	"time"
)

var UserRejectedError = errors.New("user rejected")
var UserNotFoundError = errors.New("user not found")

type Client struct {
	client *resty.Client
	config *Configuration
}

func New(config Configuration) (*Client, error) {
	client := resty.New()
	client.SetBaseURL(config.Server)
	client.SetHeaders(map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
		"User-Agent":   fmt.Sprintf("%s/%s", utils.Name, utils.Version),
	})
	u, err := url.Parse(config.Server)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "https" {
		if len(config.Certificate) > 0 && len(config.Key) > 0 {
			cert, err := tls.X509KeyPair([]byte(config.Certificate), []byte(config.Key))
			if err != nil {
				return nil, err
			}
			client.SetCertificates(cert)
		}
		if len(config.CA) > 0 {
			client.SetRootCertificate(config.CA)
		}
	}
	transport := &http.Transport{}
	if len(config.SourceInterface) > 0 {
		ip, err := common.GetIPForInterface(config.SourceInterface)
		if err == nil {
			dialer := &net.Dialer{
				Timeout:   10 * time.Second,
				LocalAddr: &net.TCPAddr{IP: ip.IP.To4()},
			}
			if u.Scheme == "https" {
				transport.DialTLSContext = dialer.DialContext
			} else {
				transport.DialContext = dialer.DialContext
			}
		} else {
			log.Errorf("Cannot get interface %s IP: %s", config.SourceInterface, err)
		}
	}
	client.SetTransport(transport)
	if len(config.Token) > 0 {
		client.SetAuthToken(config.Token)
	}
	return &Client{
		client: client,
		config: &config,
	}, nil
}

func (c *Client) getUrl(path string) string {
	return fmt.Sprintf("api/v%d/%s", c.config.ApiVersion, path)
}

/*
func (c *Client) GetStatus() (*StatusResponse, error) {
	resp, err := c.client.R().Get(c.getUrl("status"))
	if err != nil {
		return nil, err
	}
	statusCode := resp.StatusCode()
	if statusCode >= 200 && statusCode < 300 {
		status := &StatusResponse{}
		if err := json.Unmarshal(resp.Body(), status); err != nil {
			return nil, err
		}
		return status, nil
	}
	return nil, fmt.Errorf("cannot get server status: %d: %s", statusCode, resp.Status())
} */

func (c *Client) GetUser(userRequest *binding.UserRequest) (*binding.RadiusUserResponse, error) {
	resp, err := c.client.R().SetBody(userRequest).Post(c.getUrl("authorize"))
	if err != nil {
		return nil, err
	}
	statusCode := resp.StatusCode()
	switch statusCode {
	case 200:
		user := &binding.RadiusUserResponse{}
		if err := json.Unmarshal(resp.Body(), user); err != nil {
			return nil, err
		}
		return user, nil
	case 401:
		return nil, UserRejectedError
	case 404:
		return nil, UserNotFoundError
	default:
		return nil, fmt.Errorf("cannot get user authorization: %d: %s", statusCode, resp.Status())
	}
}

func (c *Client) GetCertificate() (*ubinding.RadiusCertificate, error) {
	resp, err := c.client.R().Get(c.getUrl("certificate"))
	if err != nil {
		return nil, err
	}
	statusCode := resp.StatusCode()
	switch statusCode {
	case 200:
		cert := &ubinding.RadiusCertificate{}
		if err := json.Unmarshal(resp.Body(), cert); err != nil {
			return nil, err
		}
		return cert, nil
	default:
		return nil, fmt.Errorf("cannot get certificate: %d: %s", statusCode, resp.Status())
	}
}
