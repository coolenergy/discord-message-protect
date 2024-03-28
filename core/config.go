package core

import "time"

type Config struct {
	AppPath string `json:"app_path"`
	LogPath string `json:"log_path"`

	HttpConfig                  *HttpConfig      `json:"http"`
	DiscordConfig               *DiscordConfig   `json:"discord"`
	SecretsConfig               *SecretsConfig   `json:"secrets"`
	SessionsConfig              *SessionsConfig  `json:"sessions"`
	PollutionConfig             *PollutionConfig `json:"pollution"`
	DatabaseConfig              *DatabaseConfig  `json:"database"`
	AckActionOnProtectedMessage bool
}

type HttpConfig struct {
	Hostname       string `json:"hostname"`
	Port           int    `json:"port"`
	Path           string `json:"path"`
	Scheme         string `json:"scheme"`
	ChallengePath  string `json:"challenge_path"`
	Args           map[string]interface{}
	CaptchaService string `json:"captcha_service"`
}

type DiscordConfig struct {
	BotToken           string `json:"bot_token"`
	AppId              string
	GuildId            string
	ProtectCommandName string `json:"command_name"`
}

type SessionsConfig struct {
	// Ttl time to live in minutes
	Ttl int `json:"ttl"`
	// Args is used to pass arguments specific to each implementation
	// for instance a file system based secret implementation may need to know the max number of files to keep
	// at the same time, etc. Not used for now.
	Args []byte `json:"extra"`
	// The time a user can be considered authenticated after he passes the challenge, while the user is authenticated
	// he won't be challenged anymore with captcha, this field is parsed from Ttl that's why it is not passed
	// from json config (-)
	SessionDuration time.Duration `json:"-"`

	DisconnectedUsersToRotateMap int `json:"-"`
}

type SecretsConfig struct {
	// Ttl time to live in minutes, how much time a protected message should be remembered
	Ttl int
	// Args is used to pass arguments specific to each implementation
	// for instance a file system based secret implementation may need to know the max number of files to keep
	// at the same time, etc. Not used for now.
	Args []byte
}

type PollutionConfig struct {
	StrategyName string `json:"strategy"`
	// Args is used to pass arguments specific to each implementation
	Args map[string]interface{} `json:"args"`
}

type DatabaseConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
}
