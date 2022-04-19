package updater

import (
	"github.com/COSAE-FR/ripradius/pkg/freeradius"
	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"time"
)

type Configuration struct {
	Interval time.Duration             `yaml:"interval" default:"240h"`
	CacheDir string                    `yaml:"cache" validate:"required"`
	Radius   *freeradius.Configuration `yaml:"-"`
}

func (c *Configuration) Check() error {
	if err := defaults.Set(c); err != nil {
		return err
	}
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return err
	}
	return nil
}