package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"log"
)

// GroupList structure to sort groups by `TotalAmount2DigitAfterDot` descending.
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

func MapOfGroupsToString(mapOfGroups map[string]Group) string {
	groupList := make(GroupList, 0, len(mapOfGroups))
	for _, group := range mapOfGroups {
		groupList = append(groupList, group)
	}

	// Sort the slice by TotalAmount2DigitAfterDot.
	sort.Sort(groupList)

	ts := make([]string, len(mapOfGroups))
	for i, group := range groupList {
		ts[i] = fmt.Sprintf("\n    %-35s: %7.2f", group.Name, float64(group.TotalAmount2DigitAfterDot/100))
	}
	return strings.Join(ts, "")
}

func (t *InecoTransaction) toTransaction() (trans Transaction, isExpense bool) {
	amount := t.Expense
	isExpense = true
	if amount.int == 0 {
		amount = t.Income
		isExpense = false
	}
	return Transaction{
		Date:                 t.Date.Format(OutputDateFormat),
		Details:              t.Details,
		Amount2DigitAfterDot: uint(amount.int),
	}, isExpense
}

func (s *MonthStatistics) String() string {
	income := MapOfGroupsToString(s.Income)
	expenses := MapOfGroupsToString(s.Expense)
	return fmt.Sprintf("Statistics for %s..%s:\n  Income:%s\n  Expenses:%s\n",
		s.MonthStartTimestamp.Format(OutputDateFormat),
		s.MonthEndTimestamp.Format(OutputDateFormat),
		income,
		expenses,
	)
}

type MonthStatisticsBuilder interface {

	// HandleTransaction updates inner `MonthStatistics` object with transaction details.
	// The main purpose is to choose right `Group` instance to add data into.
	HandleTransaction(trans InecoTransaction) error

	// GetMonthStatistics returns `MonthStatistic` assembled so far.
	GetMonthStatistics() *MonthStatistics
}

// groupExtractorByDetailsSubstrings is `MonthStatisticsBuilder` which uses
// `InecoTransaction.Details` field to choose right group. Logic is following:
//  1. Find is group for expenses of incomes.
//  2. Search group in `substringsToGroup` field. If there are such then update it.
//  3. Otherwise check groupUnknown value:
//  4. If `nil` then create new group with name equal to `InecoTransaction.Details` field
//  5. If some group then add into it.
type groupExtractorByDetailsSubstrings struct {
	monthStats             *MonthStatistics
	groupNamesToSubstrings map[string][]string
	substringsToGroup      map[string]Group
	groupUnknown           *Group
}

func (s groupExtractorByDetailsSubstrings) HandleTransaction(trans InecoTransaction) error {
	statTransaction, isExpense := trans.toTransaction()
	var groupMap map[string]Group
	if isExpense {
		groupMap = s.monthStats.Expense
	} else {
		groupMap = s.monthStats.Income
	}

	// Try to find group.
	found := false
	for substring, group := range s.substringsToGroup {
		if strings.Contains(trans.Details, substring) {
			found = true
			updatedGroup := group
			updatedGroup.Transactions = append(updatedGroup.Transactions, &statTransaction)
			updatedGroup.TotalAmount2DigitAfterDot += statTransaction.Amount2DigitAfterDot
			groupMap[group.Name] = updatedGroup
			break
		}
	}

	// If group is not found then create new with name = InecoTransaction.Details
	if !found {
		if s.groupUnknown != nil {
			s.groupUnknown.Transactions = append(s.groupUnknown.Transactions, &statTransaction)
			s.groupUnknown.TotalAmount2DigitAfterDot += statTransaction.Amount2DigitAfterDot
		} else {
			newGroup := Group{
				Name:                      trans.Details,
				Transactions:              []*Transaction{&statTransaction},
				TotalAmount2DigitAfterDot: statTransaction.Amount2DigitAfterDot,
			}
			groupMap[trans.Details] = newGroup
			s.substringsToGroup[trans.Details] = newGroup
		}
	}

	return nil
}

func (s groupExtractorByDetailsSubstrings) GetMonthStatistics() *MonthStatistics {
	return s.monthStats
}

type GroupExtractorBuilder func(start, end time.Time) MonthStatisticsBuilder

func NewGroupExtractorByDetailsSubstrings(
	groupNamesToSubstrings map[string][]string,
	isGroupUnknown bool,
) (GroupExtractorBuilder, error) {

	// Invert groupNamesToSubstrings and check for duplicates.
	substringsToGroupName := map[string]string{}
	for name, substrings := range groupNamesToSubstrings {
		for _, substring := range substrings {
			if group, exist := substringsToGroupName[substring]; exist {
				return nil, errors.New(fmt.Sprintf("'%s' is duplicated in '%s' and in previous '%s'",
					substring, name, group))
			}
			substringsToGroupName[substring] = name
		}
	}

	// Prepare uknownGroup if need.
	var unknownGroup *Group
	if isGroupUnknown {
		unknownGroup = &Group{Name: "Unknown", Transactions: []*Transaction{}}
	}
	log.Printf("Going to separate transactions by %d named groups from %d substrings",
		len(groupNamesToSubstrings), len(substringsToGroupName))

	return func(start, end time.Time) MonthStatisticsBuilder {

		// Create map of new Group-s for each MonthStatistic.
		substringsToGroup := map[string]Group{}
		for substring, name := range substringsToGroupName {
			substringsToGroup[substring] = Group{Name: name}
		}

		// Return new groupExtractorByDetailsSubstrings.
		return groupExtractorByDetailsSubstrings{
			monthStats: &MonthStatistics{
				MonthStartTimestamp: start,
				MonthEndTimestamp:   end,
				Income:              make(map[string]Group),
				Expense:             make(map[string]Group),
			},
			groupNamesToSubstrings: groupNamesToSubstrings,
			substringsToGroup:      substringsToGroup,
			groupUnknown:           unknownGroup,
		}
	}, nil
}

// BuildStatisticFromInecoTransactions builds a MonthStatistics from provided transactions
// with specified month start day. It uses given groupExtractorBuilder to make “
func BuildStatisticFromInecoTransactions(
	transactions []InecoTransaction,
	groupExtractorBuilder GroupExtractorBuilder,
	monthStart uint,
) ([]*MonthStatistics, error) {

	var stats []*MonthStatistics
	var monthEndDate time.Time
	var current MonthStatisticsBuilder
	for _, trans := range transactions {

		// Calculate month start and end.
		monthStartDate := time.Date(trans.Date.Year(), trans.Date.Month(), int(monthStart), 0, 0, 0, 0, time.UTC)

		// Check if this transaction is part of the current month or a new month.
		if current == nil || trans.Date.After(monthEndDate) {
			if current != nil {
				stats = append(stats, current.GetMonthStatistics())
			}
			monthEndDate = monthStartDate.AddDate(0, 1, 0).Add(-1 * time.Nanosecond)
			current = groupExtractorBuilder(monthStartDate, monthEndDate)
		}

		// Handle transaction.
		if err := current.HandleTransaction(trans); err != nil {
			return nil, err
		}
	}

	// Add last MonthStatistics.
	if current != nil {
		stats = append(stats, current.GetMonthStatistics())
	}
	return stats, nil
}
