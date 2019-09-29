package base

type Host struct {
	Name     string   `json:"name"`
	Ip       string   `json:"ip"`
	Port     int      `json:"port"`
	User     string   `json:"user"`
	Password string   `json:"password"`
	KeyFile  string   `json:"key_file"`
	Platform platform `json:"platform"`
	AuthType authType `json:"auth_type"`
}

type platform string

const (
	LinuxPlatform platform = "Linux"
	AIXPlatform   platform = "AIX"
	SunOsPlatform platform = "SunOS"
	HPPlatform    platform = "HP-UX"
)

type authType string

const (
	PasswordAuth authType = "password"
	KeyFileAuth  authType = "keyFile"
)
