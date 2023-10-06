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
		fatalError(fmt.Sprintf("Can't find configuration file '%s': %#v\n", configPath, err), true)
	}
	isOpenFileWithResult := !args.DontOpenFile

	// Parse configuration.
	config, err := readConfig(configPath)
	if err != nil {
		fatalError(fmt.Sprintf("Configuration file '%s' is wrong: %#v\n", configPath, err), isOpenFileWithResult)
	}

	// Parse timezone or set system.
	timeZone, err := time.LoadLocation(config.TimeZoneLocation)
	if err != nil {
		fatalError(fmt.Sprintf("Unknown TimeZoneLocation: %#v.\n", config.TimeZoneLocation), isOpenFileWithResult)
	}

	// Build groupsExtractor earlier to check for configuration errors.
	groupExtractorFactory, err := NewStatisticBuilderByDetailsSubstrings(
		config.GroupNamesToSubstrings,
		config.GroupAllUnknownTransactions,
		config.IgnoreSubstrings,
	)
	if err != nil {
		fatalError(fmt.Sprintf("Can't create statistic builder: %#v", err), isOpenFileWithResult)
	}

	// Log settings.
	log.Printf("Using configuration: %+v", config)

	// Parse files to raw transactions.
	rawTransactions, err := parseFiles(config.StatementFilesGlob, XmlParser{})
	if err != nil {
		fatalError(fmt.Sprintf("Can't parse transactions: %#v", err), isOpenFileWithResult)
	}
	if len(rawTransactions) < 1 {
		fatalError(fmt.Sprintf("Can't find transactions, check that '%s' matches something",
			config.StatementFilesGlob), isOpenFileWithResult)
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
		fatalError(fmt.Sprintf("Can't build statistic: %#v", err), isOpenFileWithResult)
	}

	// Process received statistics.
	result := ""
	for _, s := range statistics {
		if config.DetailedOutput {
			result = result + "\n" + s.String()
			continue
		}

		// Note that this logic is intentionally separated from `func (s *IntervalStatistic) String()`.
		income := MapOfGroupsToString(s.Income)
		expense := MapOfGroupsToString(s.Expense)
		result = result + "\n" + fmt.Sprintf(
			"\n%s..%s:\n  Income (%d, sum=%s):%s\n  Expenses (%d, sum=%s):%s",
			s.Start.Format(OutputDateFormat),
			s.End.Format(OutputDateFormat),
			len(income),
			MapOfGroupsSum(s.Income),
			strings.Join(income, ""),
			len(s.Expense),
			MapOfGroupsSum(s.Expense),
			strings.Join(expense, ""),
		)
	}
	result = fmt.Sprintf("%s\nTotal %d months.", result, len(statistics))

	// Always print result into logs and conditionally into the file which open through the OS.
	log.Print(result)
	if !args.DontOpenFile { // Twice no here, but we need in good default value for the flag and too lazy.
		writeAndOpenFile(resultFilePath, result)
	}
}

func fatalError(err string, inFile bool) {
	if inFile {
		writeAndOpenFile(resultFilePath, err)
	}
	log.Fatalf(err)
}

func writeAndOpenFile(resultFilePath, content string) {
	if err := os.WriteFile(resultFilePath, []byte(content), 0644); err != nil {
		log.Fatalf("Can't write result file into %s: %#v", resultFilePath, err)
	}
	if err := openFileInOS(resultFilePath); err != nil {
		log.Fatalf("Can't open result file %s: %#v", resultFilePath, err)
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
