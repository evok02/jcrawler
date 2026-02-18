package config

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Keywords []string
	Worker   *WorkerConfig
	DB       *DBConfig
	Seed     []string
	Log      *LogConfig
}

type LogConfig struct {
	Path string
}

type DBConfig struct {
	ConnString string
}

type WorkerConfig struct {
	Timeout time.Duration
	Delay   time.Duration
}

func setUpConfig(dir string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(dir)
}

func NewConfig(dirPath string) (*Config, error) {
	setUpConfig(dirPath)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("NewConfig: %s", err.Error())
	}
	var c = Config{
		Keywords: []string{},
		Worker:   new(WorkerConfig),
		DB:       new(DBConfig),
		Log:      new(LogConfig),
	}
	err = extractValues(&c)
	if err != nil {
		return nil, fmt.Errorf("NewConfig: %s", err.Error())
	}
	return &c, nil
}

func extractValues(c *Config) error {
	extractKeywords(c)
	err := extractWorkerConfig(c.Worker)
	if err != nil {
		return err
	}
	extractDBConfig(c)
	extractSeed(c)
	extractLogConfig(c)
	return nil
}

func extractKeywords(c *Config) {
	c.Keywords = viper.GetStringSlice("keywords")
}

func extractWorkerConfig(wc *WorkerConfig) error {
	timeout, err := time.ParseDuration(viper.GetString("worker.timeout"))
	if err != nil {
		return fmt.Errorf("extractValue: %s", err.Error())
	}
	delay, err := time.ParseDuration(viper.GetString("worker.delay"))
	if err != nil {
		return fmt.Errorf("extractValue: %s", err.Error())
	}
	wc.Timeout = timeout
	wc.Delay = delay
	return nil
}

func extractDBConfig(c *Config) error {
	viper.SetEnvPrefix("db")
	err := viper.BindEnv("conn_string")
	if err != nil {
		return fmt.Errorf("extractDBConfig: %s", err.Error())
	}
	connString := viper.GetString("conn_string")
	c.DB.ConnString = connString
	return nil
}

func extractSeed(c *Config) {
	c.Seed = viper.GetStringSlice("seed")
}

func extractLogConfig(c *Config) {
	c.Log.Path = viper.GetString("log.path")
}
