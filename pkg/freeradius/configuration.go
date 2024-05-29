package freeradius

import (
	"fmt"
	"github.com/COSAE-FR/ripradius/pkg/utils"
	"github.com/COSAE-FR/riputils/common"
	"github.com/COSAE-FR/riputils/tls"
	"github.com/creasty/defaults"
	"io/ioutil"
	"os/exec"
	"strings"
)

// Configuration holds the parameters needed to manage a dedicated Freeradius daemon
type Configuration struct {
	// Path to the FreeRadius binary
	Binary string `yaml:"binary"`
	// Launch Freeradius in debug mode (-X)
	BinaryDebug bool `yaml:"debug"`
	// Write Freeradius std{out,err} to this file
	BinaryLog string `yaml:"log_file"`
	// Make Freeradius listen on this interface
	Interface    string `yaml:"interface"`
	InterfaceNet string `yaml:"-"`
	InterfaceIP  string `yaml:"-"`
	// Make Freeradius listen on this port
	Port uint32 `yaml:"port" default:"1812"`
	// Base directory for configuration files
	RunDirectory string `yaml:"run_directory"`
	CleanOnStop  bool   `yaml:"clean_on_stop"`
	StayRoot     bool   `yaml:"stay_root"`
	// Freeradius secret
	Secret          string `yaml:"secret"`
	CA              string `yaml:"ca"`
	EnableAutoChain bool   `yaml:"enable_auto_chain"`
	NoBundle        bool   `yaml:"no_bundle"`
	Certificate     string `yaml:"certificate"`
	Key             string `yaml:"key"`
	ApiToken        string
	ApiHost         string
	ApiPort         uint32
	EnableAdmin     bool   `yaml:"enable_admin"`
	ClientNet       string `yaml:"client_net" validate:"isdefault|cidrv4"`
	// radiusd.conf tuning
	// MaxRequestTime The maximum time (in seconds) to handle a request (5 to 120).
	MaxRequestTime uint8 `yaml:"max_request_time" default:"30"`
	// CleanupDelayThe time to wait (in seconds) before cleaning up (2 to 10)
	CleanupDelay uint8 `yaml:"cleanup_delay" default:"5"`
	// MaxRequests: The maximum number of requests which the server keeps
	// track of.  This should be 256 multiplied by the number of clients.
	// (min 256)
	MaxRequests uint64 `yaml:"max_requests" default:"16384"`
	LogAuth     bool   `yaml:"log_auth"`
	// StartServers: Number of servers to start initially
	StartServers uint16 `yaml:"start_servers" default:"5"`
	// MaxServers: Limit on the total number of servers running.
	MaxServers uint16 `yaml:"max_servers" default:"32"`
	// Server-pool size regulation
	MinSpareServers uint16 `yaml:"min_spare_servers" default:"3"`
	MaxSpareServers uint16 `yaml:"max_spare_servers" default:"10"`
	MaxQueueSize    uint32 `yaml:"max_queue_size" default:"65536"`
}

func (c *Configuration) Check() error {
	if len(c.Binary) == 0 {
		var err error
		c.Binary, err = exec.LookPath(utils.FreeradiusBinaryName)
		if err != nil {
			return err
		}
	}
	if !common.FileExists(c.Binary) {
		return fmt.Errorf("freeradius binary %s does not exist", c.Binary)
	}
	if len(c.Secret) == 0 && !c.EnableAdmin {
		return fmt.Errorf("radius secret is mandatory")
	}
	if len(c.RunDirectory) == 0 {
		c.RunDirectory = utils.RunDirectory
	}
	if len(c.Interface) == 0 {
		c.Interface = utils.LoopbackInterfaceName
	}
	ifIP, err := common.GetIPForInterface(c.Interface)
	if err != nil {
		return err
	}
	c.InterfaceNet = ifIP.String()
	c.InterfaceIP = ifIP.IP.To4().String()
	if len(c.CA) > 0 {
		if !strings.Contains(c.CA, "BEGIN CERTIFICATE") {
			if !common.FileExists(c.CA) {
				return fmt.Errorf("ca is not a PEM string nor a valid file")
			}
			if content, err := ioutil.ReadFile(c.CA); err != nil {
				return fmt.Errorf("cannot reqd CA file: %s", err)
			} else {
				c.CA = string(content)
			}
		}
	}
	if len(c.Certificate) > 0 {
		if !strings.Contains(c.Certificate, "BEGIN CERTIFICATE") {
			if !common.FileExists(c.Certificate) {
				return fmt.Errorf("certificate is not a PEM string nor a valid file")
			}
			if content, err := ioutil.ReadFile(c.Certificate); err != nil {
				return fmt.Errorf("cannot reqd certificate file: %s", err)
			} else {
				c.Certificate = string(content)
			}
		}
	}
	if len(c.Key) > 0 {
		if !strings.Contains(c.Key, "BEGIN RSA PRIVATE KEY") {
			if !common.FileExists(c.Key) {
				return fmt.Errorf("key is not a PEM string nor a valid file")
			}
			if content, err := ioutil.ReadFile(c.Key); err != nil {
				return fmt.Errorf("cannot reqd key file: %s", err)
			} else {
				c.Key = string(content)
			}
		}
	}
	if len(c.Key) == 0 || len(c.Certificate) == 0 {
		cert, key, err := tls.GenerateSelfSignedCertificate()
		if err != nil {
			return fmt.Errorf("cannot generate self-signed certificate for Freeradius server: %w", err)
		}
		c.Certificate = string(cert)
		c.Key = string(key)
	}
	if err := defaults.Set(c); err != nil {
		return err
	}
	return nil
}
