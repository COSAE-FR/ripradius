package binding

import (
	"time"
)

type RadiusCertificate struct {
	SignatureDate *time.Time `json:"signature_date"`
	CA            string     `json:"ca"`
	Certificate   string     `json:"certificate"`
	Key           string     `json:"key"`
}
