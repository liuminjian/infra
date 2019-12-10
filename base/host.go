package base

type Host struct {
	Name     string   `json:"name"`
	Ip       string   `json:"ip"`
	Port     int      `json:"port"`
	User     string   `json:"user"`
	Password string   `json:"password"`
	KeyFile  string   `json:"keyFile"`
	Platform Platform `json:"platform"`
	AuthType authType `json:"authType"`
}

type Platform string

const (
	LinuxPlatform Platform = "Linux"
	AIXPlatform   Platform = "AIX"
	SunOsPlatform Platform = "SunOS"
	HPPlatform    Platform = "HP-UX"
)

type authType string

const (
	PasswordAuth authType = "password"
	KeyFileAuth  authType = "keyFile"
)
