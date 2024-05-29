package freeradius

import (
	"embed"
	"fmt"
	"github.com/COSAE-FR/ripradius/pkg/utils"
	"github.com/COSAE-FR/riputils/common"
	"github.com/Luzifer/go-dhparam"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

//go:embed assets/files
var files embed.FS

//go:embed assets/templates
var templates embed.FS

type TemplatesConfiguration struct {
	RadiusConfDir              string
	RadiusLibDir               string
	RadiusPrivateKey           string
	RadiusCertificateBundle    string
	RadiusCertificateAuthority string
	RadiusAutoChain            string
	RadiusDHParam              string
	RadiusSecret               string
	ApiServer                  string
	ApiToken                   string
	ApiAuthorizePath           string
	ApiDynamicPath             string
	FreeradiusChangeUser       bool
	FreeRadiusUser             string
	FreeRadiusGroup            string
	ListenAddress              string
	ListenPort                 uint32
	ClientNet                  string
	PrefixDirectory            string
	MaxRequestTime             uint8
	CleanupDelay               uint8
	MaxRequests                uint64
	LogAuth                    string
	StartServers               uint16
	MaxServers                 uint16
	MinSpareServers            uint16
	MaxSpareServers            uint16
	MaxQueueSize               uint32
}

func (f *Freeradius) prepareConfiguration() error {
	var err error
	var configurationBase = path.Join(f.config.RunDirectory, "radius")
	if err = os.MkdirAll(configurationBase, 0750); err != nil {
		f.log.Errorf("cannot create base configurtion directory %s: %s", configurationBase, err)
		return err
	}
	defer func() {
		if err != nil && f.config.CleanOnStop {
			if e := os.RemoveAll(configurationBase); e != nil {
				f.log.Errorf("Cannot remove base configurtion directory %s: %s", configurationBase, e)
			}
		}
	}()

	// Copy static files
	err = fs.WalkDir(files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			f.log.Error(err)
			return err
		}
		if path == "." || path == "assets" || path == "assets/files" {
			return nil
		}
		target := filepath.Join(configurationBase, strings.TrimPrefix(path, "assets/files"))
		if d.IsDir() {
			if !common.IsDirectory(target) {
				err := os.MkdirAll(target, 0755)
				if err != nil {
					f.log.Errorf("Static: cannot create %s: %s", target, err)
					return err
				}
			}
		} else {
			content, err := fs.ReadFile(files, path)
			if err != nil {
				f.log.Errorf("Static: cannot read %s: %s", path, err)
				return err
			}
			err = os.WriteFile(target, content, 0644)
			if err != nil {
				f.log.Errorf("Static: cannot write %s: %s", target, err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	configs, err := template.ParseFS(templates, "assets/templates/*.tmpl")
	if err != nil {
		f.log.Errorf("Templates: cannot parse templates: %s", err)
		return err
	}
	prefixDir, err := GetDefaultConfigurationValue("prefix")
	if err != nil {
		f.log.Tracef("cannot get value for %s, using default %s: %s", "prefix", utils.PrefixDirectory, err)
		prefixDir = utils.PrefixDirectory
	}
	libDir, err := GetDefaultConfigurationValue("libdir")
	if err != nil {
		f.log.Tracef("cannot get value for %s, using default %s: %s", "libdir", utils.FreeradiusLibDirectory, err)
		libDir = utils.FreeradiusLibDirectory
	}
	userId := "0"
	groupId := "0"
	userName := "root"
	group := "root"
	if !f.config.StayRoot {
		userName, err = GetDefaultConfigurationValue("user")
		if err != nil {
			f.log.Tracef("cannot get value for %s, using default %s: %s", "user", utils.FreeradiusUser, err)
			userName = utils.FreeradiusUser
		}
		userObj, err := user.Lookup(userName)
		if err != nil {
			f.log.Tracef("cannot get uid of %s: %s", userName, err)
		} else {
			userId = userObj.Uid
		}
		group, err = GetDefaultConfigurationValue("group")
		if err != nil {
			f.log.Tracef("cannot get value for %s, using default %s: %s", "group", utils.FreeradiusGroup, err)
			group = utils.FreeradiusGroup
		}
		groupObj, err := user.LookupGroup(group)
		if err != nil {
			f.log.Tracef("cannot get gid of %s: %s", group, err)

		} else {
			groupId = groupObj.Gid
		}
	} else {

	}
	clientNet := f.config.ClientNet
	if clientNet == "" {
		clientNet = f.config.InterfaceNet
	}
	listenIP := "127.0.0.1"
	if f.config.EnableAdmin {
		listenIP = f.config.InterfaceIP
	}
	autoCAChain := "no"
	if f.config.EnableAutoChain {
		autoCAChain = "yes"
	}

	logAuth := "no"
	if f.config.LogAuth {
		logAuth = "yes"
	}

	templatesConfig := TemplatesConfiguration{
		RadiusConfDir:           configurationBase,
		RadiusLibDir:            libDir,
		RadiusPrivateKey:        path.Join(configurationBase, "tls", "private.pem"),
		RadiusCertificateBundle: path.Join(configurationBase, "tls", "bundle.pem"),
		RadiusDHParam:           path.Join(configurationBase, "tls", "dhparam.pem"),
		RadiusAutoChain:         autoCAChain,
		RadiusSecret:            f.config.Secret,
		ApiToken:                f.config.ApiToken,
		ApiServer:               fmt.Sprintf("http://%s:%d", f.config.ApiHost, f.config.ApiPort),
		ApiAuthorizePath:        "/api/v1/authorize",
		ApiDynamicPath:          "/api/v1/dynamic-client",
		FreeradiusChangeUser:    userId != "0",
		FreeRadiusUser:          userName,
		FreeRadiusGroup:         group,
		ListenAddress:           listenIP,
		ListenPort:              f.config.Port,
		ClientNet:               clientNet,
		PrefixDirectory:         prefixDir,
		MaxRequestTime:          f.config.MaxRequestTime,
		CleanupDelay:            f.config.CleanupDelay,
		MaxRequests:             f.config.MaxRequests,
		LogAuth:                 logAuth,
		StartServers:            f.config.StartServers,
		MaxServers:              f.config.MaxServers,
		MinSpareServers:         f.config.MinSpareServers,
		MaxSpareServers:         f.config.MaxSpareServers,
		MaxQueueSize:            f.config.MaxQueueSize,
	}

	if len(f.config.CA) > 0 {
		templatesConfig.RadiusCertificateAuthority = path.Join(configurationBase, "tls", "ca.pem")
	} else {
		templatesConfig.RadiusCertificateAuthority = templatesConfig.RadiusCertificateBundle
	}

	if err = f.prepareTlsConfiguration(configurationBase, templatesConfig); err != nil {
		return err
	}

	if err = writeTemplateFile(configs, "radiusd.conf.tmpl", configurationBase, templatesConfig); err != nil {
		return err
	}
	if err = os.MkdirAll(path.Join(configurationBase, "sites-enabled"), 0755); err != nil {
		f.log.Errorf("cannot create sites-enabled directory %s: %s", configurationBase, err)
		return err
	}
	if f.config.EnableAdmin {
		if err = writeTemplateFile(configs, "default.tmpl", path.Join(configurationBase, "sites-enabled"), templatesConfig); err != nil {
			return err
		}
	}
	if err = writeTemplateFile(configs, "inner-tunnel.tmpl", path.Join(configurationBase, "sites-enabled"), templatesConfig); err != nil {
		return err
	}
	if err = writeTemplateFile(configs, "apn.tmpl", path.Join(configurationBase, "sites-enabled"), templatesConfig); err != nil {
		return err
	}
	if err = writeTemplateFile(configs, "eap.tmpl", path.Join(configurationBase, "mods-enabled"), templatesConfig); err != nil {
		return err
	}
	if err = writeTemplateFile(configs, "rest.tmpl", path.Join(configurationBase, "mods-enabled"), templatesConfig); err != nil {
		return err
	}
	if f.config.EnableAdmin {
		if err = writeTemplateFile(configs, "dynamic-clients.tmpl", path.Join(configurationBase, "sites-enabled"), templatesConfig); err != nil {
			return err
		}
		if err = writeTemplateFile(configs, "rip.tmpl", path.Join(configurationBase, "sites-enabled"), templatesConfig); err != nil {
			return err
		}
		if err = writeTemplateFile(configs, "dynamic-clients.mods.tmpl", path.Join(configurationBase, "mods-enabled"), templatesConfig); err != nil {
			return err
		}
	}
	if templatesConfig.FreeradiusChangeUser {
		f.log.Tracef("changing ownership of %s for %s(%s):%s(%s)", configurationBase, userName, userId, group, groupId)
		uid, err := strconv.Atoi(userId)
		if err != nil {
			return fmt.Errorf("cannot get uid of user %s: %w", userName, err)
		}
		gid, err := strconv.Atoi(groupId)
		if err != nil {
			return fmt.Errorf("cannot get uid of user %s: %w", group, err)
		}
		if err = ChownR(configurationBase, uid, gid, f.log); err != nil {
			return fmt.Errorf("cannot change owner of %s to %s:%s: %w", configurationBase, userName, group, err)
		}
	}
	f.log.Tracef("Configuration generated at %s", configurationBase)
	return err
}

func writeTemplateFile(tmpls *template.Template, name string, target string, config TemplatesConfiguration) error {
	targetName := strings.TrimSuffix(name, ".tmpl")
	f, err := os.OpenFile(path.Join(target, targetName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpls.ExecuteTemplate(f, name, config)
}

func (f *Freeradius) prepareTlsConfiguration(configurationBase string, templatesConfig TemplatesConfiguration) error {
	var err error
	if err = os.MkdirAll(path.Join(configurationBase, "tls"), 0755); err != nil {
		f.log.Errorf("cannot create tls directory %s: %s", configurationBase, err)
		return err
	}
	if len(f.config.CA) > 0 {
		if err = ioutil.WriteFile(templatesConfig.RadiusCertificateAuthority, []byte(f.config.CA), 0664); err != nil {
			f.log.Errorf("TLS: cannot write CA %s: %s", templatesConfig.RadiusCertificateAuthority, err)
			return err
		}
	}
	if err = ioutil.WriteFile(templatesConfig.RadiusPrivateKey, []byte(f.config.Key), 0660); err != nil {
		f.log.Errorf("TLS: cannot write key %s: %s", templatesConfig.RadiusPrivateKey, err)
		return err
	}
	bundle := f.config.Certificate
	if f.config.NoBundle == false {
		bundle = fmt.Sprintf("%s\n%s", bundle, f.config.CA)
	}
	if err = ioutil.WriteFile(templatesConfig.RadiusCertificateBundle, []byte(bundle), 0664); err != nil {
		f.log.Errorf("TLS: cannot write bundle %s: %s", templatesConfig.RadiusCertificateBundle, err)
		return err
	}

	dhFilePath := templatesConfig.RadiusDHParam
	if !common.FileExists(dhFilePath) {
		dh, err := dhparam.Generate(1024, dhparam.GeneratorTwo, nil)
		if err == nil {
			dhPem, err := dh.ToPEM()
			if err == nil {
				err = ioutil.WriteFile(dhFilePath, dhPem, 0660)
				if err != nil {
					f.log.Errorf("cannot open parameters PEM file: %s", err)
				}
			} else {
				f.log.Errorf("cannot create DH parameters PEM: %s", err)
			}
		} else {
			f.log.Errorf("cannot create DH parameters: %s", err)
		}
		err = nil
	}
	return nil
}

func GetDefaultConfigurationValue(name string) (string, error) {
	for _, base := range []string{utils.FreeradiusEtcDirectory, "/etc/raddb"} {
		confFileName := filepath.Join(base, "radiusd.conf")
		if !common.FileExists(confFileName) {
			continue
		}
		search := fmt.Sprintf(`(?m)^[ \t]*%s[ \t]*=[ \t]*\"?([^"\n]+)\"?`, name)
		valueRegex, err := regexp.Compile(search)
		if err != nil {
			return "", err
		}
		b, err := ioutil.ReadFile(confFileName)
		if err != nil {
			return "", err
		}
		match := valueRegex.FindStringSubmatch(string(b))
		if len(match) == 2 {
			return match[1], nil
		}
		return "", fmt.Errorf("%s not in %s", name, confFileName)
	}
	return "", fmt.Errorf("no valid configuration file found")
}

func ChownR(path string, uid, gid int, logger *log.Entry) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Chown(name, uid, gid)
			if err != nil && logger != nil {
				logger.Errorf("cannot chown %s to %d:%d: %s", path, uid, gid, err)
			}
		}
		return nil
	})
}
