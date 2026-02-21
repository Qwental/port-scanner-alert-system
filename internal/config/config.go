package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ProjectName string          `yaml:"project_name"`
	Masscan     MasscanConfig   `yaml:"masscan"`
	Targets     []string        `yaml:"targets"`
	Database    DatabaseConfig  `yaml:"database"`
	Scheduler   SchedulerConfig `yaml:"scheduler"`
	Telegram    TelegramConfig  `yaml:"telegram"`
	SMTP        SMTPConfig      `yaml:"smtp"`
}

type MasscanConfig struct {
	Rate      string `yaml:"rate"`
	Interface string `yaml:"interface"`
	Ports     string `yaml:"ports"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type SchedulerConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Interval string `yaml:"interval"`
}

type TelegramConfig struct {
	Enabled bool   `yaml:"enabled"`
	Token   string `yaml:"token"`   
	ChatID  int64  `yaml:"chat_id"`
}

type SMTPConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	To       string `yaml:"to"`
}

func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}


	_ = godotenv.Load()

	if token := os.Getenv("TELEGRAM_TOKEN"); token != "" {
		config.Telegram.Token = token
	}

	if chatIDStr := os.Getenv("TELEGRAM_CHAT_ID"); chatIDStr != "" {
		if chatID, err := strconv.ParseInt(chatIDStr, 10, 64); err == nil {
			config.Telegram.ChatID = chatID
		}
	}

	if smtpHost := os.Getenv("SMTP_HOST"); smtpHost != "" {
		config.SMTP.Host = smtpHost
	}

	if smtpPortStr := os.Getenv("SMTP_PORT"); smtpPortStr != "" {
		if port, err := strconv.Atoi(smtpPortStr); err == nil {
			config.SMTP.Port = port
		}
	}

	if smtpUser := os.Getenv("SMTP_USER"); smtpUser != "" {
		config.SMTP.User = smtpUser
	}

	if smtpPass := os.Getenv("SMTP_PASSWORD"); smtpPass != "" {
		config.SMTP.Password = smtpPass
	}

	if smtpTo := os.Getenv("SMTP_TO"); smtpTo != "" {
    config.SMTP.To = smtpTo
	}
	
	if smtpFrom := os.Getenv("SMTP_FROM"); smtpFrom != "" {
		config.SMTP.From = smtpFrom
	}

	return config, nil
}