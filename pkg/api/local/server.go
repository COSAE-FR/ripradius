package local

import (
	"context"
	"fmt"
	"github.com/COSAE-FR/ripradius/pkg/local/cache"
	"github.com/COSAE-FR/ripradius/pkg/local/client"
	"github.com/COSAE-FR/riputils/gin/ginlog"
	"github.com/COSAE-FR/riputils/gin/token"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"time"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

type Server struct {
	server   *http.Server
	listener net.Listener
	started  bool
	config   *Configuration
	client   *client.Client
	cache    *cache.Cache
	log      *log.Entry
}

func New(logger *log.Entry, config *Configuration, userCache *cache.Cache, upstreamClient *client.Client) (*Server, error) {
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	srv := Server{
		server: &http.Server{
			Handler: router,
		},
		config: config,
		client: upstreamClient,
		cache:  userCache,
		log:    logger.WithField("component", "api_server"),
	}
	router.Use(ginlog.Logger(srv.log), gin.Recovery())
	// router.GET("/api/v1/status", srv.status)
	operational := router.Group("/")
	if len(config.Token) > 0 {
		srv.log.Debug("Configuring token authentication")
		operational.Use(token.StaticTokenMiddleware(config.Token, srv.log))
	}
	operational.POST("/api/v1/authorize", srv.userAuthorize)
	return &srv, nil
}

func (s *Server) Configure() error {
	var err error
	s.listener, err = net.Listen("tcp4", fmt.Sprintf("%s:%d", s.config.IPAddress, s.config.Port))
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Start() error {
	if s.started {
		if err := s.Stop(); err != nil {
			s.log.Errorf("stopping API server: %s", err)
		}
	}
	if s.listener == nil {
		if err := s.Configure(); err != nil {
			s.log.Errorf("configuring API server: %s", err)
			return err
		}
	}
	s.started = true
	go func() {
		_ = s.server.Serve(s.listener)
	}()
	return nil
}

func (s *Server) Stop() error {
	if s.started {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer func() {
			cancel()
			s.started = false
		}()
		return s.server.Shutdown(ctx)
	}
	return nil
}
