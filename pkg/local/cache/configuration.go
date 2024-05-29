package cache

import (
	"github.com/creasty/defaults"
	"time"
)

type Configuration struct {
	MaxSize    int           `yaml:"size" default:"1500"`
	TTL        time.Duration `yaml:"ttl" default:"12h"`
	RefreshTTL time.Duration `yaml:"refresh_ttl" default:"1h"`
	OfflineTTL time.Duration `yaml:"offline_ttl"`
}

func (c *Configuration) Check() error {
	if err := defaults.Set(c); err != nil {
		return err
	}
	return nil
}
