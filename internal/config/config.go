package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Config содержит конфигурацию бота.
type Config struct {
	Telegram Telegram `yaml:"telegram"`
	Bybit    Bybit    `yaml:"bybit"`
	DBPath   string   `yaml:"db_path"`
}

type Telegram struct {
	Token string `yaml:"token"`
}

type Bybit struct {
	Key     string `yaml:"api_key"`
	Secret  string `yaml:"api_secret"`
	Testnet bool   `yaml:"testnet"`
}

// NewConfig загружает конфигурацию из файла.
func NewConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
