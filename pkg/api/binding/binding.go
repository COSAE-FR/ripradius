package binding

import (
	"strings"
)

type UserRequest struct {
	Username      string `json:"username" binding:"required"`
	Password      string `json:"password"`
	ClientIp      string `json:"ip" binding:"required"`
	VirtualServer string `json:"realm" binding:"required"`
	AuthType      string `json:"type"`
	Authenticator string `json:"called"`
	ClientMac     string `json:"calling"`
}

func (r UserRequest) GetClientMac() string {
	if r.ClientMac != "" {
		return r.ClientMac
	}
	parts := strings.Split(r.Authenticator, ":")
	if len(parts) >= 2 {
		return strings.Join(parts[:len(parts)-1], ":")
	}
	return ""
}

type RadiusUserResponse struct {
	TunnelType   string `json:"reply:Tunnel-Type" default:"VLAN"`
	TunnelMedium string `json:"reply:Tunnel-Medium-Type" default:"IEEE-802"`
	VLAN         uint16 `json:"reply:Tunnel-Private-Group-Id" default:"0"`
	Password     string `json:"config:Password-With-Header" binding:"required"`
}

type RadiusAdminResponse struct {
	Password string `json:"config:Password-With-Header" binding:"required"`
	Class    string `json:"reply:Class"`
}

type RadiusRejectResponse struct {
	AuthType string `json:"control:Auth-Type" default:"Reject"`
}
