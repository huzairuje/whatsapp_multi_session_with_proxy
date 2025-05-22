package config

var (
	Conf Config
	Env  string

	EnvironmentLocal = "LOCAL"
	EnvironmentDev   = "DEV"
	EnvironmentUAT   = "UAT"
	EnvironmentProd  = "PROD"
	searchPath       = []string{
		"/etc/whatsapp_multi_session_with_proxies",
		"$HOME/.whatsapp_multi_session_with_proxies",
		".",
	}
	configDefaults = map[string]interface{}{
		"port":       1234,
		"logLevel":   "DEBUG",
		"logFormat":  "text",
		"signString": "supersecret",
	}
	configName = map[string]string{
		"local": "config.local",
		"dev":   "config.dev",
		"uat":   "config.uat",
		"prod":  "config.prod",
		"test":  "config.test",
	}
)

type Config struct {
	Env             string          `mapstructure:"env"`
	Port            int             `mapstructure:"port"`
	Pprof           Pprof           `mapstructure:"pprof"`
	Proxy           Proxy           `mapstructure:"proxy"`
	StartUp         StartUp         `mapstructure:"startUp"`
	ShutDown        ShutDown        `mapstructure:"shutDown"`
	AutoLogout      bool            `mapstructure:"autoLogout"`
	AutoDisconnect  bool            `mapstructure:"autoDisconnect"`
	Cronjob         Cronjob         `mapstructure:"cronjob"`
	DeleteAfterSend DeleteAfterSend `mapstructure:"deleteAfterSend"`
	Postgres        PostgresConfig  `mapstructure:"postgres"`
	Redis           RedisConfig     `mapstructure:"redis"`
	Dashboard       DashboardConfig `mapstructure:"dashboard"` // Added for dashboard configuration
}

// DashboardConfig holds configuration specific to the dashboard feature
type DashboardConfig struct {
	JwtSecretKey string `mapstructure:"jwt_secret_key"`
	OtpSenderJID string `mapstructure:"otp_sender_jid"`
}

type Pprof struct {
	Enable       bool   `mapstructure:"enable"`
	PprofPort    string `mapstructure:"pprofPort"`
	PprofAddress string `mapstructure:"pprofAddress"`
}

type Proxy struct {
	Enable    bool   `mapstructure:"enable"`
	Directory string `mapstructure:"directory"`
}

type StartUp struct {
	EnableAutoLogin bool `mapstructure:"enableAutoLogin"`
}

type ShutDown struct {
	EnableAutoShutDown bool `mapstructure:"enableAutoShutDown"`
}

type Cronjob struct {
	AutoPresence AutoPresence `mapstructure:"autoPresence"`
}

type AutoPresence struct {
	Enable          bool   `mapstructure:"enable"`
	CronJobSchedule string `mapstructure:"cronJobSchedule"`
}

type DeleteAfterSend struct {
	Enable bool `mapstructure:"enable"`
}

// PostgresConfig ...
type PostgresConfig struct {
	Host           string `mapstructure:"host"`
	Port           string `mapstructure:"port"`
	Schema         string `mapstructure:"schema"`
	DBName         string `mapstructure:"dbName"`
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	EnablePostgres bool   `mapstructure:"enablePostgres"`
}

// RedisConfig ...
type RedisConfig struct {
	Host        string `mapstructure:"host"`
	Password    string `mapstructure:"password"`
	DB          int    `mapstructure:"db"`
	Port        int    `mapstructure:"port"`
	EnableRedis bool   `mapstructure:"enableRedis"`
}

type Proxies struct {
	Directory string `mapstructure:"directory"`
}
