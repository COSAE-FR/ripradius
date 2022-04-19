package helpers

import (
	"github.com/COSAE-FR/ripradius/pkg/api/binding"
	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func logFromVariableArgs(logger *logrus.Entry, defaultMessage string, args ...interface{}) {
	switch len(args) {
	case 0:
		logger.Info(defaultMessage)
	case 1:
		msg, ok := args[0].(string)
		if ok {
			logger.Info(msg)
		} else {
			logger.Info(defaultMessage)
		}
	default:
		msg, ok := args[0].(string)
		if ok {
			logger.Infof(msg, args[1:]...)
		} else {
			logger.Info(defaultMessage)
		}
	}
}

func RadiusReject(c *gin.Context, logger *logrus.Entry, args ...interface{}) {
	logFromVariableArgs(logger, "Rejecting user", args...)
	c.AbortWithStatusJSON(http.StatusUnauthorized, binding.RadiusRejectResponse{AuthType: "Reject"})

}

func RadiusAcceptUser(c *gin.Context, password string, vlanId uint16, logger *logrus.Entry, args ...interface{}) {
	response := &binding.RadiusUserResponse{
		VLAN:     vlanId,
		Password: password,
		TunnelMedium: "IEEE-802",
		TunnelType: "VLAN",
	}
	if err := defaults.Set(c); err != nil {
		logger.Errorf("Cannot populate authorize response with defaults: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{
			"message": "Cannot create user response",
		})
		return
	}
	logFromVariableArgs(logger, "Accepting user", args...)
	c.AbortWithStatusJSON(http.StatusOK, response)
}

func RadiusAcceptAdmin(c *gin.Context, password string, class string, logger *logrus.Entry, args ...interface{}) {
	response := &binding.RadiusAdminResponse{
		Password: password,
		Class: class,
	}
	if err := defaults.Set(c); err != nil {
		logger.Errorf("Cannot populate authorize response with defaults: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{
			"message": "Cannot create admin response",
		})
		return
	}
	logFromVariableArgs(logger, "Accepting user", args...)
	c.AbortWithStatusJSON(http.StatusOK, response)
}
