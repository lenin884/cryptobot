package config

import (
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

// Config содержит конфигурацию бота.
type Config struct {
	TelegramToken string `yaml:"telegram_token"`
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
