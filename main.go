package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alexflint/go-arg"
)

type Args struct {
	FilePath         string `arg:"positional" help:"'Statement #############' CSV (not supported yet) or XML downloaded from https://online.inecobank.am"`
	MonthStart       uint   `arg:"-s" default:"1" help:"Day of month to treat as a month start."`
	IsDetailedOutput bool   `arg:"-d" default:"false" help:"Print detailed statistic."`
}

type FileParser interface {
	ParseRawTransactionsFromFile(args Args) ([]InecoTransaction, error)
}

func main() {
	var args Args
	argsParser := arg.MustParse(&args)

	// Check if the file path argument is provided.
	if args.FilePath == "" {
		argsParser.WriteHelp(os.Stdout)
		fmt.Println("Please provide the path to a file with 'Statement #############.xml' downloaded" +
			"from https://online.inecobank.am. First open account from which you want analyze expenses," +
			"next put into 'From' and 'To' fields dates you want to analyze, press 'Search', scroll" +
			"page to bottom and here at right corner will be 5 icons to download statement." +
			"Press XML button and save file. Next specify path to this file to the script.")
		os.Exit(1)
	}
	_, err := os.Stat(args.FilePath)
	if os.IsNotExist(err) {
		fmt.Printf("File '%s' does not exist.\n", args.FilePath)
		os.Exit(1)
	}

	// Validate month start.
	if args.MonthStart < 1 || args.MonthStart > 31 {
		argsParser.WriteHelp(os.Stdout)
		fmt.Println("Error: Month start must be between 1 and 31.")
		os.Exit(1)
	}

	dotItems := strings.Split(args.FilePath, ".")
	fileExtension := dotItems[len(dotItems)-1]
	var parser FileParser
	switch fileExtension {
	case "xml":
		parser = XmlParser{}
	case "csv":
		parser = CSVParser{}
	}
	log.Printf("Going to parse with %v parser with settings %+v", parser, args)
	rawTransactions, err := parser.ParseRawTransactionsFromFile(args)
	if err != nil {
		fmt.Println("Can't parse transactions:", err)
		os.Exit(2)
	}
	if len(rawTransactions) < 1 {
		log.Fatal("Can't find transactions.")
	}
	log.Printf("Found %d transactions.", len(rawTransactions))

	// Create statistics builder.
	ge, err := NewGroupExtractorByDetailsSubstrings(
		map[string][]string{
			"Yandex Taxi":   {"YANDEX"},
			"Vika's health": {"ARABKIR JMC"},
			"Groceries":     {"CHEESE MARKET"},
		},
		args.IsDetailedOutput, // Make "group per uknown transaction" only if "verbose" output requested.
	)
	if err != nil {
		fmt.Println("Can't create statistic builder:", err)
		os.Exit(2)
	}

	// Build statistic.
	statistics, err := BuildStatisticFromInecoTransactions(rawTransactions, ge, args.MonthStart)
	if err != nil {
		fmt.Println("Can't build statistic:", err)
		os.Exit(2)
	}

	// Process the received statistics.
	for _, stat := range statistics {
		if args.IsDetailedOutput {
			log.Println(stat)
			continue
		}

		log.Printf(
			"%s..%s:\n  Income:%s\n  Expenses:%s",
			stat.MonthStartTimestamp.Format(OutputDateFormat),
			stat.MonthEndTimestamp.Format(OutputDateFormat),
			MapOfGroupsToString(stat.Income),
			MapOfGroupsToString(stat.Expense),
		)
	}
	log.Printf("Total %d month.", len(statistics))
}
