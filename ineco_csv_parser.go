package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
)

type TransactionCsv struct {
	Nn                     string                  `csv:"nn"`
	Number                 string                  `csv:"number"`
	Date                   DateTime                `csv:"omitempty"`
	Currency               string                  `csv:"currency"`
	Income                 MoneyWith2DecimalPlaces `csv:"omitempty"`
	Expense                MoneyWith2DecimalPlaces `csv:"omitempty"`
	RecieverOrPayerAccount string                  `csv:"omitempty"`
	RecieverOrPayer        string                  `csv:"omitempty"`
	Details                string                  `csv:"omitempty"`
}

// DateTime is a wrapper for standard Time to be parsed from custom format.
type DateTime struct {
	time.Time
}

func (date *DateTime) UnmarshalCSV(field string) (err error) {
	date.Time, err = time.Parse(InecoDateFormat, field)
	return err
}

func (m *MoneyWith2DecimalPlaces) UnmarshalCSV(field string) (err error) {
	floatVal, err := strconv.ParseFloat(strings.ReplaceAll(field, ",", ""), 32)
	if err != nil {
		return err
	}
	m.int = int(floatVal * 100)
	return nil
}

type CSVParser struct{}

func (p CSVParser) ParseRawTransactionsFromFile(args Args) ([]InecoTransaction, error) {

	// Open the CSV file.
	file, err := os.Open(args.FilePath)
	if err != nil {
		return nil, fmt.Errorf("Error opening '%s' file: %w", args.FilePath, err)
	}
	defer file.Close()

	// Prepare reader which may find start and end of data in the CSV.
	csvReader := InsideCSVReader(
		file,
		"n/n,Number,Date,Currency,Income,Expense,Receiver/Payer Account,Receiver/Payer,Details",
		"Total",
	)

	// Provide header to the csvutil library.
	// Note that standard "encoding/csv" fails because CSV doesn't match standard CSV format.
	// Ineco CSV-s have fields enclosed in single double-quote characters with comma inside.
	// "github.com/jszwec/csvutil", "github.com/gocarina/gocsv" are based on "encoding/csv" so don't work as well.

	// Configure custom transformations.
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		return gocsv.LazyCSVReader(in) // Allows use of quotes in CSV
	})
	// Time format is custom.
	// unmarshalTime := func(data []byte, t *time.Time) error {
	// 	tt, err := time.Parse(inecoCsvDateFormat, string(data))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	*t = tt
	// 	return nil
	// }
	// decoder.Register(unmarshalTime)
	// // Money amounts contains commas as thousands separators.
	// decoder.Map = func(field, column string, v any) string {
	// 	if _, ok := v.(float32); ok {
	// 		return strings.ReplaceAll(field, ",", "")
	// 	}
	// 	return field
	// }

	transactions := []*TransactionCsv{}
	if err := gocsv.UnmarshalWithoutHeaders(csvReader, &transactions); err != nil {
		return nil, fmt.Errorf("Invalid CSV in '%s' file: %w", args.FilePath, err)
	}

	// Read and parse the remaining lines
	// for {
	// 	var transaction TransactionCsv
	// 	if err := decoder.Decode(&transaction); err == io.EOF {
	// 		break
	// 	} else if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	transactions = append(transactions, &transaction)
	// }
	// for {
	// 	fields, err := csvReader.Read()
	// 	if err == io.EOF {
	// 		break
	// 	} else if err != nil {
	// 		return nil, err
	// 	}

	// 	date, err := parseDate(fields[2])
	// 	if err != nil {
	// 		return nil, fmt.Errorf(
	// 			"Invalid CSV in '%s' file. On the 3 cell of '%v' expected date, but failed with: %w",
	// 			filePath, fields, err,
	// 		)
	// 	}
	// 	income, err := parseFloat(fields[4])
	// 	if err != nil {
	// 		return nil, fmt.Errorf(
	// 			"Invalid CSV in '%s' file. On the 5 cell of '%v' expected money amount, but failed with: %w",
	// 			filePath, fields, err,
	// 		)
	// 	}
	// 	expense, err := parseFloat(fields[5])
	// 	if err != nil {
	// 		return nil, fmt.Errorf(
	// 			"Invalid CSV in '%s' file. On the 6 cell of '%v' expected money amount, but failed with: %w",
	// 			filePath, fields, err,
	// 		)
	// 	}
	// 	transaction := &TransactionCsv{
	// 		Nn:                     fields[0],
	// 		Number:                 fields[1],
	// 		Date:                   *date,
	// 		Currency:               fields[3],
	// 		Income:                 *income,
	// 		Expense:                *expense,
	// 		RecieverOrPayerAccount: fields[6],
	// 		RecieverOrPayer:        fields[7],
	// 		Details:                fields[8],
	// 	}
	// 	transactions = append(transactions, transaction)
	// }

	inecoTransactions := make([]InecoTransaction, 0, len(transactions))
	for _, t := range transactions {
		inecoTransactions = append(inecoTransactions, InecoTransaction{
			Nn:                     t.Nn,
			Number:                 t.Number,
			Date:                   t.Date.Time,
			Currency:               t.Currency,
			Income:                 t.Income,
			Expense:                t.Expense,
			RecieverOrPayerAccount: t.RecieverOrPayer,
			RecieverOrPayer:        t.RecieverOrPayer,
			Details:                t.Details,
		})
	}
	return inecoTransactions, nil
}

// func (t *TransactionCsv) UnmarshalCSV(fields []string, filePath string) error {
// 	if len(fields) != 9 {
// 		return errors.New("invalid number of fields in record")
// 	}

// 	date, err := parseDate(fields[2])
// 	if err != nil {
// 		return fmt.Errorf(
// 			"Invalid CSV in '%s' file. On the 3 cell of '%v' expected date, but failed with: %w",
// 			filePath, fields, err,
// 		)
// 	}
// 	income, err := parseFloat(fields[4])
// 	if err != nil {
// 		return fmt.Errorf(
// 			"Invalid CSV in '%s' file. On the 5 cell of '%v' expected money amount, but failed with: %w",
// 			filePath, fields, err,
// 		)
// 	}
// 	expense, err := parseFloat(fields[5])
// 	if err != nil {
// 		return fmt.Errorf(
// 			"Invalid CSV in '%s' file. On the 6 cell of '%v' expected money amount, but failed with: %w",
// 			filePath, fields, err,
// 		)
// 	}
// 	t.Nn = fields[0]
// 	t.Number = fields[1]
// 	t.Date = *date
// 	t.Currency = fields[3]
// 	t.Income = *income
// 	t.Expense = *expense
// 	t.RecieverOrPayerAccount = fields[6]
// 	t.RecieverOrPayer = fields[7]
// 	t.Details = fields[8]

// 	return nil
// }

func parseDate(dateStr string) (*time.Time, error) {
	date, err := time.Parse(InecoDateFormat, dateStr)
	if err != nil {
		return nil, err
	}
	return &date, nil
}

func parseFloat(floatStr string) (*float32, error) {
	floatVal, err := strconv.ParseFloat(strings.ReplaceAll(floatStr, ",", ""), 32)
	if err != nil {
		return nil, err
	}
	value := float32(floatVal)
	return &value, nil
}
