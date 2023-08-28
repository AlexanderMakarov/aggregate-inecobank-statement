package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
)

type Args struct {
	ConfigPath   string `arg:"positional" help:"Path to the configuration YAML file. By default is used 'config.yaml' path."`
	DontOpenFile bool   `arg:"-n" help:"Flag to don't open result file in OS at the end, only print in STDOUT."`
}

type FileParser interface {
	ParseRawTransactionsFromFile(filePath string) ([]InecoTransaction, error)
}

// Version is application version string and should be updated with `go build -ldflags`.
var Version = "development"

const resultFilePath = "Inecobank Aggregated Statement.txt"

func main() {
	log.Printf("Version %s", Version)
	configPath := "config.yaml"

	// Parse arguments and set configPath.
	var args Args
	arg.MustParse(&args)
	if args.ConfigPath != "" {
		configPath = args.ConfigPath
	}
	configPath, err := getAbsolutePath(configPath)
	if err != nil {
		log.Fatalf("Can't find configuration file '%s': %#v\n", configPath, err)
	}

	// Parse configuration.
	config, err := readConfig(configPath)
	if err != nil {
		log.Fatalf("Configuration file '%s' is wrong: %#v\n", configPath, err)
	}

	// Parse timezone or set system.
	timeZone, err := time.LoadLocation(config.TimeZoneLocation)
	if err != nil {
		log.Fatalf("Unknown TimeZoneLocation: %#v.\n", config.TimeZoneLocation)
	}

	// Build groupsExtractor earlier to check for configuration errors.
	groupExtractorFactory, err := NewStatisticBuilderByDetailsSubstrings(
		config.GroupNamesToSubstrings,
		config.GroupAllUnknownTransactions,
		config.IgnoreSubstrings,
	)
	if err != nil {
		log.Fatalln("Can't create statistic builder:", err)
	}

	// Log settings.
	log.Printf("Using configuration: %+v", config)

	// Parse files to raw transactions.
	rawTransactions, err := parseFiles(config.StatementFilesGlob, XmlParser{})
	if err != nil {
		log.Fatalln("Can't parse transactions:", err)
	}
	if len(rawTransactions) < 1 {
		log.Fatal("Can't find transactions.")
	}
	log.Printf("Total found %d transactions.", len(rawTransactions))

	// Build statistic.
	statistics, err := BuildMonthlyStatisticFromInecoTransactions(
		rawTransactions,
		groupExtractorFactory,
		config.MonthStartDayNumber,
		timeZone,
	)
	if err != nil {
		log.Fatalln("Can't build statistic:", err)
	}

	// Process received statistics.
	result := ""
	for _, s := range statistics {
		if config.DetailedOutput {
			result = fmt.Sprintf("%s\nTotal %d months.", s.String(), len(statistics))
			continue
		}

		// Note that this logic is intentionally separated from `func (s *IntervalStatistic) String()`.
		income := MapOfGroupsToString(s.Income)
		expense := MapOfGroupsToString(s.Expense)
		result = fmt.Sprintf(
			"\n%s..%s:\n  Income (%d, sum=%s):%s\n  Expenses (%d, sum=%s):%s\nTotal %d months.",
			s.Start.Format(OutputDateFormat),
			s.End.Format(OutputDateFormat),
			len(income),
			MapOfGroupsSum(s.Income),
			strings.Join(income, ""),
			len(s.Expense),
			MapOfGroupsSum(s.Expense),
			strings.Join(expense, ""),
			len(statistics),
		)
	}

	// Always print result into logs and conditionally into the file which open through the OS.
	log.Print(result)
	if !args.DontOpenFile { // Twice no here, but we need in good default value for the flag and too lazy.
		if err := os.WriteFile(resultFilePath, []byte(result), 0644); err != nil {
			log.Fatalf("Can't write result file into %s: %#v", resultFilePath, err)
		}
		openFileInOS(resultFilePath)
	}
}

func parseFiles(glog string, parser FileParser) ([]InecoTransaction, error) {
	files, err := getFilesByGlob(glog)
	if err != nil {
		return nil, err
	}

	result := make([]InecoTransaction, 0)
	for _, file := range files {
		log.Printf("Parsing '%s' with %v parser.", file, parser)
		rawTransactions, err := parser.ParseRawTransactionsFromFile(file)
		if err != nil {
			log.Printf("!!! Can't parse transactions from '%s' file: %#v", file, err)
		}
		if len(rawTransactions) < 1 {
			log.Printf("!!! Can't find transactions in '%s' file.", file)
		}
		log.Printf("Found %d transactions in '%s' file.", len(rawTransactions), file)
		result = append(result, rawTransactions...)
	}
	return result, nil
}
