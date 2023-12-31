package main

import "time"

// MoneyWith2DecimalPlaces is a wrapper to parse money from "1,500.00" or "1,500" to 150000.
type MoneyWith2DecimalPlaces struct {
	int
}

const InecoDateFormat = "02/01/2006"

type InecoTransaction struct {
	Nn                     string
	Number                 string
	Date                   time.Time
	Currency               string
	Income                 MoneyWith2DecimalPlaces
	Expense                MoneyWith2DecimalPlaces
	RecieverOrPayerAccount string
	RecieverOrPayer        string
	Details                string
}

const OutputDateFormat = "2006-01-02"

type Transaction struct {
	Date    time.Time
	Details string
	Amount  MoneyWith2DecimalPlaces
}

type Group struct {
	Name         string
	Total        MoneyWith2DecimalPlaces
	Transactions []*Transaction
}

type IntervalStatistic struct {
	Start   time.Time
	End     time.Time
	Income  map[string]*Group
	Expense map[string]*Group
}
