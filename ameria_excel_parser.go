package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tealeg/xlsx"
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
const giveUpFindHeaderAfterEmpty1Cells = 15

var (
	xlsxHeaders = []string{
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

func (p MyAmeriaExcelFileParser) ParseRawTransactionsFromFile(
	filePath string,
) ([]Transaction, error) {
	f, err := xlsx.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Find first sheet.
	firstSheet := f.Sheets[0]
	fmt.Printf("%s: parsing first sheet '%s', total %d sheets.\n",
		filePath, firstSheet.Name, len(f.Sheets))

	// Parse myAmeriaTransactions.
	var myAmeriaTransactions []MyAmeriaTransaction
	var isHeaderRowFound bool
	for i, row := range firstSheet.Rows {
		cells := row.Cells
		if len(cells) < len(xlsxHeaders) {
			return nil,
				fmt.Errorf(
					"%s: %d row has only %d cells while need to find information for headers %v",
					filePath, i, len(cells), xlsxHeaders,
				)
		}
		// Find header row.
		if !isHeaderRowFound {
			if i > giveUpFindHeaderAfterEmpty1Cells {
				return nil, fmt.Errorf(
					"%s: after scanning %d rows can't find headers %v",
					filePath, i, xlsxHeaders,
				)
			}
			var isCellMatches = true
			for cellIndex, header := range xlsxHeaders {
				if strings.TrimSpace(cells[cellIndex].String()) != header {
					isCellMatches = false
					break
				}
			}
			if isCellMatches {
				isHeaderRowFound = true
			}

			// Skip this row anyway.
			continue
		}

		// Stop if row doesn't have enough cells or first cell is empty.
		if len(cells) < len(xlsxHeaders) || cells[0].String() == "" {
			break
		}

		// Parse date and amount.
		date, err := time.Parse(MyAmeriaDateFormat, cells[0].String())
		if err != nil {
			return nil, fmt.Errorf("failed to parse date from 1st cell of %d row: %w", i, err)
		}
		var amount MoneyWith2DecimalPlaces
		if err := amount.UnmarshalText([]byte(cells[9].String())); err != nil {
			return nil, fmt.Errorf("failed to parse amount from 10th cell of %d row: %w", i, err)
		}

		transaction := MyAmeriaTransaction{
			Date:               date,
			FactN:              cells[1].String(),
			PO:                 cells[2].String(),
			OutgoingAccount:    cells[3].String(),
			BeneficiaryAccount: cells[4].String(),
			PayerOrBeneficiary: cells[5].String(),
			Details:            cells[6].String(),
			Status:             cells[7].String(),
			Comment:            cells[8].String(),
			Amount:             amount,
			Currency:           cells[10].String(),
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
