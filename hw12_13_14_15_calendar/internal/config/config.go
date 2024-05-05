package config

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.

import (
	"log"
	"os"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Logger     LoggerConf
	DB         DBConf
	Queue      QueueConf
	Server     ServerConf
	GRPCServer GRPCServerConf
	Scheduler  SchedulerConf
}

type LoggerConf struct {
	Level string
}

type DBConf struct {
	Type string // "memory", "sql"
	SQL  SQLConf
}

type QueueConf struct {
	Type string // "rmq" only
	RMQ  RMQConf
}

type ServerConf struct {
	Host string
	Port string
}

type GRPCServerConf struct {
	Host string
	Port string
}

type SQLConf struct {
	Driver   string
	Name     string
	User     string
	Password string
	Host     string
	Port     string
}

type RMQConf struct {
	Name     string
	User     string
	Password string
	Host     string
	Port     string
}

type SchedulerConf struct {
	PurgeIntervalDays int `yaml:"purgeIntervalDays"`
}

func NewConfig() *Config {
	return &Config{}
}

func Parse(filePath string) (*Config, error) {
	configData, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	cfg := NewConfig()
	err = yaml.Unmarshal(configData, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, err
}
