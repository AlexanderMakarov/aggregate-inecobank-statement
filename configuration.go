package main

import (
	"os"

	_ "github.com/go-playground/validator/v10"

	"gopkg.in/yaml.v3"
)

type Config struct {
	StatementFilesGlob          string              `yaml:"StatementFilesGlob" validate:"required,filepath"`
	DetailedOutput              bool                `yaml:"detailedOutput"`
	MonthStartDayNumber         uint                `yaml:"monthStartDayNumber" validate:"min=1,max=31"`
	TimeZoneLocation            string              `yaml:"TimeZoneLocation" validate:"timezone"`
	GroupAllUnknownTransactions bool                `yaml:"groupAllUnknownTransactions"`
	GroupNamesToSubstrings      map[string][]string `yaml:"groupNamesToSubstrings"`
	IgnoreSubstrings            []string            `yaml:"ignoreSubstrings"`
}

func readConfig(filename string) (*Config, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = yaml.Unmarshal(buf, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
