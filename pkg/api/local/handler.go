package local

import (
	"errors"
	"github.com/COSAE-FR/ripradius/pkg/api/binding"
	"github.com/COSAE-FR/ripradius/pkg/api/helpers"
	"github.com/COSAE-FR/ripradius/pkg/local/cache"
	"github.com/COSAE-FR/ripradius/pkg/local/client"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Server) refreshUser(c *gin.Context, requestedUser *binding.UserRequest, errorFunc gin.HandlerFunc) {
	var serverOffline bool
	logger := s.log.WithFields(map[string]interface{}{
		"user":    requestedUser.Username,
		"src_mac": requestedUser.GetClientMac(),
		"src_ip":  requestedUser.ClientIp,
	})
	defer func() {
		if serverOffline {
			logger.Debug("Enabling offline mode")
			s.cache.SetOffline()
		} else {
			logger.Trace("Enabling online mode")
			s.cache.SetOnline()
		}
	}()
	user, err := s.client.GetUser(requestedUser)
	if err != nil {
		if errors.Is(err, client.UserRejectedError) {
			logger.Debugf("User rejected by authenticator")
			errorFunc(c)
			return
		}
		if errors.Is(err, client.UserNotFoundError) {
			logger.Debugf("User not found by authenticator")
			errorFunc(c)
			return
		}
		logger.Errorf("Error with authenticator: %s", err)
		serverOffline = true
		errorFunc(c)
		return
	}
	logger.Trace("Adding user to cache")
	if err := s.cache.AddUser(cache.User{
		Username: requestedUser.Username,
		Password: user.Password,
		Mac:      requestedUser.GetClientMac(),
		VlanId:   user.VLAN,
	}); err != nil {
		logger.Errorf("Cannot add user to cache: %s", err)
	}
	helpers.RadiusAcceptUser(c, user.Password, user.VLAN, logger)
}

func (s *Server) userAuthorize(c *gin.Context) {
	userRequest := binding.UserRequest{}
	if err := c.ShouldBindJSON(&userRequest); err != nil {
		s.log.Errorf("Cannot decode JSON client request: %s", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	logger := s.log.WithFields(map[string]interface{}{
		"user":    userRequest.Username,
		"src_mac": userRequest.GetClientMac(),
		"src_ip":  userRequest.ClientIp,
	})
	cachedUser, mustRefresh, found := s.cache.GetUserWithRefreshNeed(userRequest.Username, userRequest.GetClientMac())
	if !found {
		logger.Trace("User not in cache, refreshing")
		s.refreshUser(c, &userRequest, func(c *gin.Context) {
			helpers.RadiusReject(c, logger)
		})
		return
	}
	if mustRefresh {
		logger.Trace("User in cache for a while, refreshing")
		s.refreshUser(c, &userRequest, func(c *gin.Context) {
			helpers.RadiusAcceptUser(c, cachedUser.Password, cachedUser.VlanId, logger)
		})
		return
	}
	helpers.RadiusAcceptUser(c, cachedUser.Password, cachedUser.VlanId, logger)
}

func (s *Server) status(c *gin.Context) {
	cacheStatus := s.cache.Status()
	c.AbortWithStatusJSON(http.StatusOK, &binding.ServerStatus{Cache: cacheStatus})
}
