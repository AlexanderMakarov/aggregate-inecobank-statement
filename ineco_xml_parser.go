package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

type XmlDate struct {
	time.Time
}

type XmlTransaction struct {
	NN                   string                  `xml:"n-n"`
	Number               string                  `xml:"Number"`
	Date                 XmlDate                 `xml:"Date"`
	Currency             string                  `xml:"Currency"`
	Income               MoneyWith2DecimalPlaces `xml:"Income"`
	Expense              MoneyWith2DecimalPlaces `xml:"Expense"`
	ReceiverPayerAccount string                  `xml:"Receiver-PayerAccount"`
	ReceiverPayer        string                  `xml:"Receiver-Payer"`
	Details              string                  `xml:"Details"`
}

type Operations struct {
	Transactions []XmlTransaction `xml:"Operation"`
}

type Statement struct {
	Client         string     `xml:"Client" validate:"required"`
	AccountNumber  string     `xml:"AccountNumber" validate:"required"`
	Currency       string     `xml:"Currency" validate:"required"`
	Period         string     `xml:"Period" validate:"required"`
	OpeningBalance string     `xml:"Openingbalance" validate:"required"`
	ClosingBalance string     `xml:"Closingbalance" validate:"required"`
	Operations     Operations `xml:"Operations" validate:"required"`
}

func (m *MoneyWith2DecimalPlaces) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	d.DecodeElement(&v, &start)
	v = strings.Replace(v, ",", "", -1)
	floatVal, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return err
	}
	m.int = int(floatVal * 100)
	return nil
}

func (xd *XmlDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	d.DecodeElement(&v, &start)

	parse, err := time.Parse(InecoDateFormat, v)
	if err != nil {
		return err
	}

	xd.Time = parse
	return nil
}

type XmlParser struct {
}

// ParseRawTransactionsFromFile implements FileParser.
func (XmlParser) ParseRawTransactionsFromFile(filePath string) ([]InecoTransaction, error) {

	// Open XML file.
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening '%s' file: %w", filePath, err)
	}
	defer file.Close()

	// Read the file content
	xmlData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Unmarshal XML.
	var stmt Statement
	err = xml.Unmarshal(xmlData, &stmt)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling XML: %v", err)
	}

	// Validate that all fields are set.
	validate := validator.New()
	for i, operation := range stmt.Operations.Transactions {
		err = validate.Struct(operation)
		if err != nil {
			return nil, fmt.Errorf("error in %d transaction in '%s': %v", i+1, filePath, err)
		}
	}

	inecoTransactions := make([]InecoTransaction, 0, len(stmt.Operations.Transactions))
	for _, t := range stmt.Operations.Transactions {
		inecoTransactions = append(inecoTransactions, InecoTransaction{
			Nn:                     t.NN,
			Number:                 t.Number,
			Date:                   t.Date.Time,
			Currency:               t.Currency,
			Income:                 t.Income,
			Expense:                t.Expense,
			RecieverOrPayerAccount: t.ReceiverPayerAccount,
			RecieverOrPayer:        t.ReceiverPayer,
			Details:                t.Details,
		})
	}
	return inecoTransactions, nil
}

var _ FileParser = XmlParser{}
