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
	BulkSend        BulkSendConfig  `mapstructure:"bulkSend"`
	Postgres        PostgresConfig  `mapstructure:"postgres"`
	Redis           RedisConfig     `mapstructure:"redis"`
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

// BulkSendConfig controls anti-ban behavior for bulk messaging.
type BulkSendConfig struct {
	// MinDelay is the minimum delay between messages in milliseconds (default: 15000 = 15s)
	MinDelay int `mapstructure:"minDelay"`
	// MaxDelay is the maximum delay between messages in milliseconds (default: 45000 = 45s)
	MaxDelay int `mapstructure:"maxDelay"`
	// BatchSize is the number of messages to send before taking a batch pause (default: 10)
	BatchSize int `mapstructure:"batchSize"`
	// BatchPauseMin is the minimum batch pause in seconds (default: 300 = 5 minutes)
	BatchPauseMin int `mapstructure:"batchPauseMin"`
	// BatchPauseMax is the maximum batch pause in seconds (default: 600 = 10 minutes)
	BatchPauseMax int `mapstructure:"batchPauseMax"`
	// DailyLimit is the max messages per sender per day (default: 50)
	DailyLimit int `mapstructure:"dailyLimit"`
	// TypingDelayMin is the minimum "composing" presence duration in milliseconds before sending (default: 2000)
	TypingDelayMin int `mapstructure:"typingDelayMin"`
	// TypingDelayMax is the maximum "composing" presence duration in milliseconds before sending (default: 5000)
	TypingDelayMax int `mapstructure:"typingDelayMax"`
	// EnablePresenceSimulation enables sending "composing" before each message (default: true)
	EnablePresenceSimulation bool `mapstructure:"enablePresenceSimulation"`
	// AllowedHourStart is the earliest hour to send messages (0-23, default: 8 = 8 AM)
	AllowedHourStart int `mapstructure:"allowedHourStart"`
	// AllowedHourEnd is the latest hour to send messages (0-23, default: 22 = 10 PM)
	AllowedHourEnd int `mapstructure:"allowedHourEnd"`
	// Timezone for time-of-day restrictions (default: "Local")
	Timezone string `mapstructure:"timezone"`
	// EnableTimeRestrictions enables time-of-day sending restrictions (default: true)
	EnableTimeRestrictions bool `mapstructure:"enableTimeRestrictions"`
	// ErrorBackoffMinutes is how long to pause after rate limit error (default: 30)
	ErrorBackoffMinutes int `mapstructure:"errorBackoffMinutes"`
	// EnableRecipientValidation validates recipients before sending (default: true)
	EnableRecipientValidation bool `mapstructure:"enableRecipientValidation"`
	// ValidationCacheDuration is how long to cache validation results in hours (default: 24)
	ValidationCacheDuration int `mapstructure:"validationCacheDuration"`
	// EnableHealthCheck checks session health before bulk send (default: true)
	EnableHealthCheck bool `mapstructure:"enableHealthCheck"`
	// MaxErrorRate is the maximum acceptable error rate (0.0-1.0, default: 0.3 = 30%)
	MaxErrorRate float64 `mapstructure:"maxErrorRate"`
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
