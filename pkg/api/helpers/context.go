package helpers

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func GetLogger(defaultLogger *log.Entry, ctx *gin.Context) *log.Entry {
	requestLogger, found := ctx.Get("logger")
	if !found {
		return defaultLogger
	}
	logger, ok := requestLogger.(*log.Entry)
	if !ok {
		return defaultLogger
	}
	return logger
}
