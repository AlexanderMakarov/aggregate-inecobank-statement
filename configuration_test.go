package main

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/thlib/go-timezone-local/tzlocal"
)

func TestReadConfig_ValidYAML(t *testing.T) {
	// Arrange
	tempFile := createTempFileWithContent(
		`inecobankStatementFilesGlob: "*.xml"
ameriaCsvFilesGlob: "*.csv"
myAmeriaHistoryFilesGlob: "*.xls"
myAmeriaMyAccounts: 
  - Account1
  - Account2
myAmeriaIncomeSubstrings:
  - Income
  - Salary
detailedOutput: true
monthStartDayNumber: 1
timeZoneLocation: "America/New_York"
groupAllUnknownTransactions: true
groupNamesToSubstrings:
  g1:
    - Sub1
    - Sub2
  g2:
    - Sub3
ignoreSubstrings:
  - Ignore1
  - Ignore2
`,
	)
	defer os.Remove(tempFile.Name())

	// Act
	cfg, err := readConfig(tempFile.Name())

	// Assert
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if cfg == nil {
		t.Error("Expected config, but got nil")
	}
	if cfg.InecobankStatementFilesGlob != "*.xml" {
		t.Errorf(
			"Expected InecobankStatementFilesGlob to be '*.xml', got '%s'",
			cfg.InecobankStatementFilesGlob,
		)
	}
	if cfg.AmeriaCsvFilesGlob != "*.csv" {
		t.Errorf(
			"Expected AmeriaCsvFilesGlob to be '*.csv', got '%s'",
			cfg.InecobankStatementFilesGlob,
		)
	}
	if cfg.MyAmeriaHistoryFilesGlob != "*.xls" {
		t.Errorf(
			"Expected MyAmeriaHistoryFilesGlob to be '*.xls', got '%s'",
			cfg.MyAmeriaHistoryFilesGlob,
		)
	}
	if len(cfg.MyAmeriaMyAccounts) != 2 || cfg.MyAmeriaMyAccounts[0] != "Account1" || cfg.MyAmeriaMyAccounts[1] != "Account2" {
		t.Errorf(
			"Expected MyAmeriaMyAccounts to be ['Account1', 'Account2'], got '%v'",
			cfg.MyAmeriaMyAccounts,
		)
	}
	if len(cfg.MyAmeriaIncomeSubstrings) != 2 || cfg.MyAmeriaIncomeSubstrings[0] != "Income" || cfg.MyAmeriaIncomeSubstrings[1] != "Salary" {
		t.Errorf(
			"Expected MyAmeriaIncomeSubstrings to be ['Income', 'Salary'], got '%v'",
			cfg.MyAmeriaIncomeSubstrings,
		)
	}
	if !cfg.DetailedOutput {
		t.Error("Expected DetailedOutput to be true")
	}
	if cfg.MonthStartDayNumber != 1 {
		t.Errorf("Expected MonthStartDayNumber to be 1, got '%d'", cfg.MonthStartDayNumber)
	}
	if cfg.TimeZoneLocation != "America/New_York" {
		t.Errorf("Expected TimeZoneLocation to be 'America/New_York', got '%s'", cfg.TimeZoneLocation)
	}
	if !cfg.GroupAllUnknownTransactions {
		t.Error("Expected GroupAllUnknownTransactions to be true")
	}
	if len(cfg.GroupNamesToSubstrings) != 2 || len(cfg.GroupNamesToSubstrings["g1"]) != 2 || cfg.GroupNamesToSubstrings["g1"][0] != "Sub1" || cfg.GroupNamesToSubstrings["g1"][1] != "Sub2" || len(cfg.GroupNamesToSubstrings["g2"]) != 1 || cfg.GroupNamesToSubstrings["g2"][0] != "Sub3" {
		t.Errorf(
			"Expected GroupNamesToSubstrings to have correct mappings, got '%v'",
			cfg.GroupNamesToSubstrings,
		)
	}
	if len(cfg.IgnoreSubstrings) != 2 || cfg.IgnoreSubstrings[0] != "Ignore1" || cfg.IgnoreSubstrings[1] != "Ignore2" {
		t.Errorf(
			"Expected IgnoreSubstrings to be ['Ignore1', 'Ignore2'], got '%v'",
			cfg.IgnoreSubstrings,
		)
	}
}

func TestReadConfig_InvalidYAML(t *testing.T) {
	// Arrange. Note that "myAmeriaIncomeSubstrings" doesn't have ":" at the end.
	tempFile := createTempFileWithContent(
		`inecobankStatementFilesGlob: "*.xml"
ameriaCsvFilesGlob: "*.csv"
myAmeriaHistoryFilesGlob: "*.csv"
myAmeriaMyAccounts: 
  - Account1
  - Account2
myAmeriaIncomeSubstrings
  - Income
  - Salary
detailedOutput: true
monthStartDayNumber: 1
timeZoneLocation: "America/New_York"
groupAllUnknownTransactions: true
groupNamesToSubstrings:
  g1:
    - Sub1
    - Sub2
  g2:
    - Sub3
ignoreSubstrings:
  - Ignore1
  - Ignore2
`,
	)
	defer os.Remove(tempFile.Name())

	// Act
	_, err := readConfig(tempFile.Name())

	// Assert
	if err == nil {
		t.Fatal("Expected error, but got no error")
	}
	checkErrorContainsSubstring(t, err, "yaml: line 7: could not find expected ':'")
}

func TestReadConfig_MisstypedField(t *testing.T) {
	// Arrange. Note that "groupsNamesToSubstrings" is a wrong name.
	tempFile := createTempFileWithContent(
		`inecobankStatementFilesGlob: "*.xml"
ameriaCsvFilesGlob: "*.csv"
myAmeriaHistoryFilesGlob: "*.csv"
myAmeriaMyAccounts: 
  - Account1
  - Account2
myAmeriaIncomeSubstrings:
  - Income
  - Salary
detailedOutput: true
monthStartDayNumber: 1
timeZoneLocation: "America/New_York"
groupAllUnknownTransactions: true
groupsNamesToSubstrings:
  g1:
    - Sub1
    - Sub2
  g2:
    - Sub3
ignoreSubstrings:
  - Ignore1
  - Ignore2
`,
	)
	defer os.Remove(tempFile.Name())

	// Act
	_, err := readConfig(tempFile.Name())

	// Assert
	if err == nil {
		t.Fatal("Expected error, but got no error")
	}
	checkErrorContainsSubstring(t, err, "line 14: field groupsNamesToSubstrings not found in type")
}

func TestReadConfig_FileNotFound(t *testing.T) {
	// Arrange
	nonexistentFile := "nonexistent_file.yaml"

	// Act
	_, err := readConfig(nonexistentFile)

	// Assert
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected os.ErrNotExist error, but got: %v", err)
	}
}

func TestReadConfig_EmptyFile(t *testing.T) {
	// Arrange
	tempFile := createTempFileWithContent("")
	defer os.Remove(tempFile.Name())

	// Act
	_, err := readConfig(tempFile.Name())

	// Assert
	if err == nil {
		t.Fatal("Expected error, but got no error")
	}
	checkErrorContainsSubstring(t, err, "EOF")
	checkErrorContainsSubstring(t, err, "can't decode YAML from configuration file")
}

func TestReadConfig_NotAllFields(t *testing.T) {
	// Arrange
	tempFile := createTempFileWithContent(
		`inecobankStatementFilesGlob: "*.xml"
ameriaCsvFilesGlob: "*.csv"
myAmeriaHistoryFilesGlob: "*.xls"
detailedOutput: false
groupAllUnknownTransactions: true
groupNamesToSubstrings:
  g1:
    - Sub1
    - Sub2
  g2:
    - Sub3
`,
	)
	defer os.Remove(tempFile.Name())

	// Act
	cfg, err := readConfig(tempFile.Name())

	// Assert
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if cfg == nil {
		t.Error("Expected config, but got nil")
	}
	if cfg.InecobankStatementFilesGlob != "*.xml" {
		t.Errorf(
			"Expected InecobankStatementFilesGlob to be '*.xml', got '%s'",
			cfg.InecobankStatementFilesGlob,
		)
	}
	if cfg.AmeriaCsvFilesGlob != "*.csv" {
		t.Errorf(
			"Expected AmeriaCsvFilesGlob to be '*.csv', got '%s'",
			cfg.InecobankStatementFilesGlob,
		)
	}
	if cfg.MyAmeriaHistoryFilesGlob != "*.xls" {
		t.Errorf(
			"Expected MyAmeriaHistoryFilesGlob to be '*.xls', got '%s'",
			cfg.MyAmeriaHistoryFilesGlob,
		)
	}
	if len(cfg.MyAmeriaMyAccounts) != 0 {
		t.Errorf(
			"Expected MyAmeriaMyAccounts to be empty, got '%v'",
			cfg.MyAmeriaMyAccounts,
		)
	}
	if len(cfg.MyAmeriaIncomeSubstrings) != 0 {
		t.Errorf(
			"Expected MyAmeriaIncomeSubstrings to be empty, got '%v'",
			cfg.MyAmeriaIncomeSubstrings,
		)
	}
	if cfg.DetailedOutput {
		t.Error("Expected DetailedOutput to be false")
	}
	if cfg.MonthStartDayNumber != 1 {
		t.Errorf("Expected MonthStartDayNumber to be 1, got '%d'", cfg.MonthStartDayNumber)
	}
	tzname, _ := tzlocal.RuntimeTZ()
	if cfg.TimeZoneLocation != tzname {
		t.Errorf("Expected TimeZoneLocation to be '%s', got '%s'", tzname, cfg.TimeZoneLocation)
	}
	if !cfg.GroupAllUnknownTransactions {
		t.Error("Expected GroupAllUnknownTransactions to be true")
	}
	if len(cfg.GroupNamesToSubstrings) != 2 || len(cfg.GroupNamesToSubstrings["g1"]) != 2 || cfg.GroupNamesToSubstrings["g1"][0] != "Sub1" || cfg.GroupNamesToSubstrings["g1"][1] != "Sub2" || len(cfg.GroupNamesToSubstrings["g2"]) != 1 || cfg.GroupNamesToSubstrings["g2"][0] != "Sub3" {
		t.Errorf(
			"Expected GroupNamesToSubstrings to have correct mappings, got '%v'",
			cfg.GroupNamesToSubstrings,
		)
	}
	if len(cfg.IgnoreSubstrings) != 0 {
		t.Errorf(
			"Expected IgnoreSubstrings to be empty, got '%v'",
			cfg.IgnoreSubstrings,
		)
	}
}

// createTempFileWithContent creates a temporary file with the given content.
func createTempFileWithContent(content string) *os.File {
	tempFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		panic(err)
	}
	if _, err := tempFile.WriteString(content); err != nil {
		panic(err)
	}
	return tempFile
}

func checkErrorContainsSubstring(t *testing.T, err error, substring string) {
	if !strings.Contains(err.Error(), substring) {
		t.Errorf(
			"Expected error message to contain '%s', got '%s'",
			substring,
			err.Error(),
		)
	}
}
