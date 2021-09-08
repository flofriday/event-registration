package eventregistration

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Template      string `toml:"template"`
	Filename      string `toml:"filename"`
	Port          int    `toml:"port"`
	AdminEnable   bool   `toml:"admin-enable"`
	AdminPassword string `toml:"admin-password"`
}

func LoadConfig(filename string) (*Config, error) {
	tomlData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var conf Config
	if _, err := toml.Decode(string(tomlData), &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
