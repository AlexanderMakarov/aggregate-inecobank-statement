package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func (m *MoneyWith2DecimalPlaces) UnmarshalText(text []byte) error {
	sanitizedText := strings.Replace(string(text), ",", "", -1)
	floatVal, err := strconv.ParseFloat(sanitizedText, 64)
	if err != nil {
		return err
	}
	m.int = int(floatVal * 100)
	return nil
}

const MyAmeriaDateFormat = "02/01/2006"

var (
	headers = []string{
		"Ամսաթիվ",
		"Փաստ N",
		"ԳՏ",
		"Ելքագրվող հաշիվ",
		"Շահառուի հաշիվ",
		"Վճարող/Շահառու",
		"Մանրամասներ",
		"Կարգավիճակ",
		"Մեկնաբանություն",
		"Գումար",
		"Արժույթ",
	}
)

type MyAmeriaTransaction struct {
	Date               time.Time
	FactN              string
	PO                 string
	OutgoingAccount    string
	BeneficiaryAccount string
	PayerOrBeneficiary string
	Details            string
	Status             string
	Comment            string
	Amount             MoneyWith2DecimalPlaces
	Currency           string
}

type MyAmeriaExcelFileParser struct {
	MyAccounts              []string
	DetailsIncomeSubstrings []string
}

func (p MyAmeriaExcelFileParser) ParseRawTransactionsFromFile(filePath string) ([]Transaction, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Find first sheet.
	firstSheet := f.WorkBook.Sheets.Sheet[0].Name
	fmt.Printf("%s: '%s' is first sheet of %d sheets.\n", filePath, firstSheet, f.SheetCount)

	rows, err := f.GetRows(firstSheet)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	// Parse myAmeriaTransactions.
	var myAmeriaTransactions []MyAmeriaTransaction
	var isHeaderRowFound bool
	for i, row := range rows {
		// Find header row.
		if !isHeaderRowFound {
			if len(row) < len(headers) {
				continue
			}
			var isCellMatches = false
			for cellIndex := range headers {
				if strings.TrimSpace(row[cellIndex]) != headers[cellIndex] {
					isCellMatches = false
					break
				}
				isCellMatches = true
			}
			if isCellMatches {
				isHeaderRowFound = true
				continue
			}
		}
		// Stop if row doesn't have enough cells.
		if len(row) < len(headers) {
			break
		}

		// Parse date and amount.
		date, err := time.Parse(MyAmeriaDateFormat, row[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse date from 1st cell of %d row: %w", i, err)
		}
		var amount MoneyWith2DecimalPlaces
		if err := amount.UnmarshalText([]byte(row[9])); err != nil {
			return nil, fmt.Errorf("failed to parse amount from 10th cell of %d row: %w", i, err)
		}

		transaction := MyAmeriaTransaction{
			Date:               date,
			FactN:              row[1],
			PO:                 row[2],
			OutgoingAccount:    row[3],
			BeneficiaryAccount: row[4],
			PayerOrBeneficiary: row[5],
			Details:            row[6],
			Status:             row[7],
			Comment:            row[8],
			Amount:             amount,
			Currency:           row[10],
		}
		myAmeriaTransactions = append(myAmeriaTransactions, transaction)
	}

	// Convert MyAmeria rows to unified transactions and separate expenses from incomes.
	transactions := make([]Transaction, len(myAmeriaTransactions))
	for i, transaction := range myAmeriaTransactions {
		isExpense := true
		if len(p.MyAccounts) > 0 {
			if slices.Contains(p.MyAccounts, transaction.BeneficiaryAccount) {
				isExpense = false
			}
		} else if len(p.DetailsIncomeSubstrings) > 0 {
			for _, substring := range p.DetailsIncomeSubstrings {
				if strings.Contains(transaction.Details, substring) {
					isExpense = false
					break
				}
			}
		}
		transactions[i] = Transaction{
			IsExpense: isExpense,
			Date:      transaction.Date,
			Details:   transaction.Details,
			Amount:    transaction.Amount,
		}
	}

	return transactions, nil
}
