package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/jszwec/csvutil"
)

const DATE_FORMAT = "2006-01-02"

type TransactionCsv struct {
	Nn                     string    `csv:"-"`
	Number                 string    `csv:"-"`
	Date                   time.Time `csv:",omitempty"`
	Currency               string    `csv:"-"`
	Income                 float32   `csv:",omitempty"`
	Expense                float32   `csv:",omitempty"`
	RecieverOrPayerAccount string    `csv:"-"`
	RecieverOrPayer        string    `csv:",omitempty"`
	Details                string    `csv:",omitempty"`
}

func buildStatisticFromRecords(records [][]string, monthStart uint) ([]*MonthStatistics, error) {
	// monthlyStatistic := []*MonthStatistics{}
	// for _, record := range records {
	// }

	res := []*MonthStatistics{}
	// for _, month := range monthlyStatistic {

	// }
	return res, nil
}

func ParseInecoStatementCSV(filePath string, monthStart uint) ([]*MonthStatistics, error) {
	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Join(errors.New("Error opening '"+filePath+"' file: "), err)
	}
	defer file.Close()

	// Scan lines between header (inclusive, needed for csvutils input)
	// scanner := bufio.NewScanner(file)
	// dataLines := []string{}
	// isDataStarted := false
	// for scanner.Scan() {
	// 	line := scanner.Text()
	// 	if isDataStarted {
	// 		if strings.HasPrefix("Total", line) {
	// 			break
	// 		}
	// 		dataLines = append(dataLines, line)
	// 	}
	// 	if line == "n/n,Number,Date,Currency,Income,Expense,Receiver/Payer Account,Receiver/Payer,Details" {
	// 		// Scan next line to skip header.
	// 		scanner.Text()
	// 		isDataStarted = true
	// 		continue
	// 	}
	// }

	reader := InsideCSVReader(
		file,
		"n/n,Number,Date,Currency,Income,Expense,Receiver/Payer Account,Receiver/Payer,Details",
		"Total",
	)

	// Build csvutil reader.
	csvReader := csv.NewReader(reader)

	// Provide csvutil with header.
	transactionsHeader, err := csvutil.Header(TransactionCsv{}, "csv")
	if err != nil {
		return nil, errors.Join(errors.New("Wrong header struct is provided for csvutil lib: "), err)
	}

	dec, err := csvutil.NewDecoder(csvReader, transactionsHeader...)
	if err != nil {
		return nil, errors.Join(
			errors.New("Mismatched to '"+filePath+"' content header struct is provided for csvutil lib: "),
			err,
		)
	}
	var transactions []TransactionCsv
	for {
		var u TransactionCsv
		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.Join(
				errors.New("Wrong line from '"+filePath+"' is recieved by csvutil lib: "),
				err,
			)
		}
		transactions = append(transactions, u)
	}

	// // Create a new CSV reader
	// reader := csv.NewReader(file)

	// // Read the CSV records
	// records, err := reader.ReadAll()
	// if err != nil {
	// 	return nil, errors.Join(errors.New("Error reading CSV from '"+filePath+"' file: "), err)
	// }

	return nil, nil //buildStatisticFromRecords(records, monthStart)
}

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

	stat, err := ParseInecoStatementCSV(args.FilePath, args.MonthStart)
	if err != nil {
		fmt.Println(err)
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
