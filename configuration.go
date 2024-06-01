package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/thlib/go-timezone-local/tzlocal"
	"gopkg.in/yaml.v3"
)

type Config struct {
	InecobankStatementFilesGlob string              `yaml:"inecobankStatementFilesGlob" validate:"required,filepath,min=1"`
	MyAmeriaHistoryFilesGlob    string              `yaml:"myAmeriaHistoryFilesGlob" validate:"required,filepath,min=1"`
	MyAmeriaMyAccounts          []string            `yaml:"myAmeriaMyAccounts,omitempty"`
	MyAmeriaIncomeSubstrings    []string            `yaml:"myAmeriaIncomeSubstrings,omitempty"`
	DetailedOutput              bool                `yaml:"detailedOutput"`
	MonthStartDayNumber         uint                `yaml:"monthStartDayNumber,omitempty" validate:"min=1,max=31" default:"1"`
	TimeZoneLocation            string              `yaml:"timeZoneLocation,omitempty" validate:"timezone"`
	GroupAllUnknownTransactions bool                `yaml:"groupAllUnknownTransactions"`
	IgnoreSubstrings            []string            `yaml:"ignoreSubstrings,omitempty"`
	GroupNamesToSubstrings      map[string][]string `yaml:"groupNamesToSubstrings"`
}

func readConfig(filename string) (*Config, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	decoder := yaml.NewDecoder(strings.NewReader(string(buf)))
	decoder.KnownFields(true) // Disallow unknown fields
	if err = decoder.Decode(cfg); err != nil {
		if err.Error() == "EOF" {
			return nil, fmt.Errorf("can't decode YAML from configuration file '%s': %v", filename, err)
		}
		return nil, err
	}

	// Set default values.
	if cfg.MonthStartDayNumber == 0 {
		cfg.MonthStartDayNumber = 1
	}
	if len(cfg.TimeZoneLocation) == 0 {
		tzname, err := tzlocal.RuntimeTZ()
		if err != nil {
			return nil, err
		}
		cfg.TimeZoneLocation = tzname
	}

	// Validate.
	validate := validator.New()
	if err = validate.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
