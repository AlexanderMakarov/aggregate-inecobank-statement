package main

import "time"

// MoneyWith2DecimalPlaces is a wrapper to parse money from "1,500.00" or "1,500" to 150000.
type MoneyWith2DecimalPlaces struct {
	int
}

const OutputDateFormat = "2006-01-02"

type Transaction struct {
	IsExpense bool
	Date      time.Time
	Details   string
	Amount    MoneyWith2DecimalPlaces
}

type Group struct {
	Name         string
	Total        MoneyWith2DecimalPlaces
	Transactions []Transaction
}

type IntervalStatistic struct {
	Start   time.Time
	End     time.Time
	Income  map[string]*Group
	Expense map[string]*Group
}
