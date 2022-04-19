package cache

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/COSAE-FR/riputils/cache"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

type User struct {
	Username string
	Password string
	Mac      string
	VlanId   uint16
}

// Status provides statistics for cache
type Status struct {
	Hits    int  `json:"hits"`
	Misses  int  `json:"misses"`
	Added   int  `json:"added"`
	Evicted int  `json:"evicted"`
	Entries int  `json:"entries"`
	Offline bool `json:"offline"`
}

type Cache struct {
	cache   cache.Cache
	config  *Configuration
	offline bool
	log     *log.Entry
	sync.Mutex
}

func getUserKey(username string, mac string) string {
	return fmt.Sprintf("user|%x", sha256.Sum256([]byte(fmt.Sprintf("%s|%s", strings.ToLower(username), strings.ToLower(mac)))))
}

func New(logger *log.Entry, config *Configuration) (*Cache, error) {
	cacheLogger := logger.WithField("component", "cache")
	c, err := cache.New(cache.MaxKeys(config.MaxSize), cache.TTL(config.TTL), cache.LRU())
	if err != nil {
		cacheLogger.Errorf("Cannot create cache backend: %s", err)
		return nil, err
	}
	return &Cache{
		cache:  c,
		config: config,
		log:    cacheLogger,
	}, nil
}

func (c *Cache) Status() Status {
	stats := c.cache.Stat()
	return Status{
		Hits:    stats.Hits,
		Misses:  stats.Misses,
		Added:   stats.Added,
		Evicted: stats.Evicted,
		Entries: c.cache.Len(),
		Offline: c.offline,
	}
}

func (c *Cache) SetOffline() {
	c.Lock()
	defer c.Unlock()
	c.log.Trace("Setting cache offline")
	c.cache.ChangeTTL(c.config.OfflineTTL)
	c.offline = true
}

func (c *Cache) SetOnline() {
	c.Lock()
	defer c.Unlock()
	c.log.Trace("Setting cache online")
	c.cache.ChangeTTL(c.config.TTL)
	c.offline = false
}

func (c *Cache) GetUser(username string, mac string) (User, bool) {
	user, _, found := c.GetUserWithAge(username, mac)
	return user, found
}

func (c *Cache) GetUserWithAge(username string, mac string) (User, time.Duration, bool) {
	logger := c.log.WithFields(map[string]interface{}{
		"user":    username,
		"src_mac": mac,
	})
	if c.cache == nil {
		logger.Error("Cache is not ready")
		return User{}, 0, false
	}
	entry, age, found := c.cache.GetWithAge(getUserKey(username, mac))
	if !found {
		logger.Trace("Entry not in cache")
		return User{}, 0, found
	}
	user, ok := entry.(User)
	if !ok {
		logger.Error("Cannot parse cached entry")
		return User{}, 0, ok
	}
	return user, age, true
}

func (c *Cache) GetUserWithRefreshNeed(username string, mac string) (User, bool, bool) {
	user, age, found := c.GetUserWithAge(username, mac)
	if !found {
		return user, true, found
	}
	if age > c.config.RefreshTTL {
		return user, true, found
	}
	return user, false, found
}

func (c *Cache) HasUser(username string, mac string) bool {
	if c.cache == nil {
		c.log.WithFields(map[string]interface{}{
			"user":    username,
			"src_mac": mac,
		}).Error("Cache is not ready")
		return false
	}
	return c.cache.Has(getUserKey(username, mac))
}

func (c *Cache) AddUser(user User) error {
	if c.cache == nil {
		c.log.WithFields(map[string]interface{}{
			"user":    user.Username,
			"src_mac": user.Mac,
		}).Error("Cache is not ready")
		return errors.New("cache is not ready")
	}
	c.cache.Set(getUserKey(user.Username, user.Mac), user)
	return nil
}
