package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
)

var args struct {
	FilePath         string `arg:"positional" help:"'Statement #############.csv' downloaded from https://online.inecobank.am"`
	MonthStart       uint   `arg:"-s" default:"1" help:"Day of month to treat as a month start."`
	IsDetailedOutput bool   `arg:"-d" default:"false" help:"Print detailed statistic."`
}

func main() {
	argsParser := arg.MustParse(&args)

	// Check if the file path argument is provided.
	if args.FilePath == "" {
		argsParser.WriteHelp(os.Stdout)
		fmt.Println("Please provide the path to a file with 'Statement #############.csv' downloaded" +
			"from https://online.inecobank.am. First open account from which you want analyze expenses," +
			"next put into 'From' and 'To' fields dates you want to analyze, press 'Search', scroll" +
			"page to bottom and here at right corner will be 5 icons to download statement." +
			"Press CSV button and save file. Next specify path to this file to the script.")
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

	rawTransactions, err := ParseTransactionCsvFromInecoCsvFile(args.FilePath)
	if err != nil {
		fmt.Println("Can't parse transactions:", err)
		os.Exit(2)
	}
	stat, err := BuildStatisticFromRecords(rawTransactions, args.MonthStart)
	if err != nil {
		fmt.Println("Can't build statistic:", err)
		os.Exit(2)
	}

	// Process the received statistics.
	for _, record := range stat {
		if args.IsDetailedOutput {
			fmt.Println(record)
			continue
		}
		// for _, value := range record {
		// 	fmt.Printf("%s\t", value)
		// }
		// fmt.Println()
	}
}
