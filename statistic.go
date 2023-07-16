package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Transaction struct {
	Date                 string
	Details              string
	Amount2DigitAfterDot uint
}

type Group struct {
	Name                      string
	TotalAmount2DigitAfterDot int
	Transactions              []Transaction
}

type GroupList []Group

func (g GroupList) Len() int {
	return len(g)
}

func (g GroupList) Less(i, j int) bool {
	return g[i].TotalAmount2DigitAfterDot < g[j].TotalAmount2DigitAfterDot
}

func (g GroupList) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

func directionToString(direction map[string]Group) string {
	groupList := make(GroupList, 0, len(direction))
	for _, group := range direction {
		groupList = append(groupList, group)
	}

	// Sort the slice by TotalAmount2DigitAfterDot.
	sort.Sort(groupList)

	ts := make([]string, len(direction))
	for i, group := range groupList {
		ts[i] = fmt.Sprintf("\n    %-35s: %7.2f", group.Name, float64(group.TotalAmount2DigitAfterDot/100))
	}
	return strings.Join(ts, "")
}

type MonthStatistics struct {
	MonthStartTimestamp int64
	MonthEndTimestamp   int64
	Income              map[string]Group
	Expense             map[string]Group
}

func (s *MonthStatistics) String() string {
	income := directionToString(s.Income)
	expenses := directionToString(s.Expense)
	return fmt.Sprintf("Statistics for %s..%s:\n  Income:%s\n  Expenses:%s\n",
		time.Unix(s.MonthStartTimestamp, 0).Format(DATE_FORMAT),
		time.Unix(s.MonthEndTimestamp, 0).Format(DATE_FORMAT),
		income,
		expenses,
	)
}
