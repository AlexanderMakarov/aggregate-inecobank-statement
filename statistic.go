package main

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"log"
)

func (t *InecoTransaction) toTransaction() (trans Transaction, isExpense bool) {
	amount := t.Expense
	isExpense = true
	if amount.int == 0 {
		amount = t.Income
		isExpense = false
	}
	trans = Transaction{
		Date:    t.Date,
		Details: t.Details,
		Amount:  amount,
	}
	return
}

func (t *Transaction) String() string {
	return fmt.Sprintf("Transaction %s %s %s", t.Date.Format(OutputDateFormat), t.Amount, t.Details)
}

func (m MoneyWith2DecimalPlaces) String() string {
	dollars := m.int / 100
	cents := m.int % 100
	dollarString := strconv.Itoa(dollars)
	for i := len(dollarString) - 3; i > 0; i -= 3 {
		dollarString = dollarString[:i] + "," + dollarString[i:]
	}
	return fmt.Sprintf("%9s.%02d", dollarString, cents)
}

// GroupList structure to sort groups by `TotalAmount2DigitAfterDot` descending.
type GroupList []*Group

func (g GroupList) Len() int {
	return len(g)
}

func (g GroupList) Less(i, j int) bool {
	return g[i].Total.int > g[j].Total.int
}

func (g GroupList) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

// InecoTransactionList structure to sort transaction by `Date` ascending.
type InecoTransactionList []InecoTransaction

func (g InecoTransactionList) Len() int {
	return len(g)
}

func (g InecoTransactionList) Less(i, j int) bool {
	return g[i].Date.Before(g[j].Date)
}

func (g InecoTransactionList) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

// MapOfGroupsToStringFull converts map of `Group`-s to human readable string.
// `withTransactions` parameter allows to output all transactions for the each group.
func MapOfGroupsToStringFull(mapOfGroups map[string]*Group, withTransactions bool) []string {
	groupList := make(GroupList, 0, len(mapOfGroups))
	for _, group := range mapOfGroups {
		groupList = append(groupList, group)
	}

	// Sort the slice by TotalAmount2DigitAfterDot.
	sort.Sort(groupList)

	groupStrings := make([]string, len(mapOfGroups))
	for i, group := range groupList {
		groupStrings[i] = fmt.Sprintf("\n    %-35s: %s", group.Name, &group.Total)
		if withTransactions {
			transStrings := make([]string, 0, len(group.Transactions))
			for _, t := range group.Transactions {
				transStrings = append(transStrings, t.String())
			}
			groupStrings[i] = fmt.Sprintf("%s, from %d transaction(s):\n      %s", groupStrings[i],
				len(transStrings), strings.Join(transStrings, "\n      "))
		}
	}
	return groupStrings
}

// MapOfGroupsToStringFull converts map of `Group`-s to human readable string.
func MapOfGroupsToString(mapOfGroups map[string]*Group) []string {
	return MapOfGroupsToStringFull(mapOfGroups, false)
}

func (s *IntervalStatistic) String() string {
	income := MapOfGroupsToStringFull(s.Income, true)
	expense := MapOfGroupsToStringFull(s.Expense, true)
	return fmt.Sprintf("Statistics for %s..%s:\n  Income (%d, sum=%s):%s\n  Expenses (%d, sum=%s):%s\n",
		s.Start.Format(OutputDateFormat),
		s.End.Format(OutputDateFormat),
		len(income),
		MapOfGroupsSum(s.Income),
		strings.Join(income, ""),
		len(s.Expense),
		MapOfGroupsSum(s.Expense),
		strings.Join(expense, ""),
	)
}

// MapOfGroupsSum returns sum from all groups.
func MapOfGroupsSum(mapOfGroups map[string]*Group) MoneyWith2DecimalPlaces {
	sum := MoneyWith2DecimalPlaces{}
	for _, group := range mapOfGroups {
		sum.int += group.Total.int
	}
	return sum
}

// IntervalStatisticsBuilder builds `IntervalStatistic` from `InecoTransaction-s`.
type IntervalStatisticsBuilder interface {

	// HandleTransaction updates inner `MonthStatistics` object with transaction details.
	// The main purpose is to choose right `Group` instance to add data into.
	HandleTransaction(trans *InecoTransaction) error

	// GetIntervalStatistic returns `IntervalStatistic` assembled so far.
	GetIntervalStatistic() *IntervalStatistic
}

const UnknownGroupName = "unknown"

// groupExtractorByDetailsSubstrings is [main.IntervalStatisticsBuilder] which uses
// `InecoTransaction.Details` field to choose right group. Logic is following:
//  1. Find is group for expenses of incomes.
//  2. Search group in `substringsToGroupName` field. If there are such then update it.
//  3. Otherwise check isGroupAllUnknown value:
//  4. If `false` then create new group with name equal to `InecoTransaction.Details` field
//  5. If `true` then add into single group with name from `UnknownGroupName` constant.
type groupExtractorByDetailsSubstrings struct {
	intervalStats          *IntervalStatistic
	groupNamesToSubstrings map[string][]string
	substringsToGroupName  map[string]string
	isGroupAllUnknown      bool
}

func (s groupExtractorByDetailsSubstrings) HandleTransaction(trans *InecoTransaction) error {
	statTransaction, isExpense := trans.toTransaction()

	// Choose map of groups to operate on.
	var mapOfGroups map[string]*Group
	if isExpense {
		mapOfGroups = s.intervalStats.Expense
	} else {
		mapOfGroups = s.intervalStats.Income
	}

	// Try to find group in configuration.
	found := false
	for substring, groupName := range s.substringsToGroupName {
		if strings.Contains(trans.Details, substring) {
			// Refresh group from the map! Because golang makes copy of `group` which leads to flacky behavior.
			existingGroup, exists := mapOfGroups[groupName]
			if !exists {
				// If the group doesn't exist in the map, create a new one.
				existingGroup = &Group{
					Name:         groupName,
					Total:        MoneyWith2DecimalPlaces{},
					Transactions: []*Transaction{},
				}
				mapOfGroups[groupName] = existingGroup
			}
			existingGroup.Transactions = append(existingGroup.Transactions, &statTransaction)
			existingGroup.Total.int += statTransaction.Amount.int
			found = true
			break
		}
	}

	// Check group is found in configuration.
	if !found {
		// Choose name of custom group to search.
		var groupName string
		if s.isGroupAllUnknown {
			groupName = UnknownGroupName
		} else {
			groupName = trans.Details
		}

		if group, exists := mapOfGroups[groupName]; exists {
			// If exists then update group with current transaction.
			group.Total.int += statTransaction.Amount.int
			group.Transactions = append(group.Transactions, &statTransaction)
		} else {
			// If not exists, create new and add to group.
			newGroup := Group{
				Name:         groupName,
				Total:        statTransaction.Amount,
				Transactions: []*Transaction{&statTransaction},
			}
			mapOfGroups[newGroup.Name] = &newGroup
		}
	}

	return nil
}

func (s groupExtractorByDetailsSubstrings) GetIntervalStatistic() *IntervalStatistic {
	return s.intervalStats
}

type GroupExtractorBuilder func(start, end time.Time) IntervalStatisticsBuilder

// NewGroupExtractorByDetailsSubstrings returns
// [github.com/AlexanderMakarov/aggregate-inecobank-statement.main.GroupExtractorBuilder] which builds
// [github.com/AlexanderMakarov/aggregate-inecobank-statement.main.groupExtractorByDetailsSubstrings] in a safe way.
func NewGroupExtractorByDetailsSubstrings(
	groupNamesToSubstrings map[string][]string,
	isGroupAllUnknownTransactions bool,
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
	log.Printf("Going to separate transactions by %d named groups from %d substrings",
		len(groupNamesToSubstrings), len(substringsToGroupName))

	return func(start, end time.Time) IntervalStatisticsBuilder {

		// Return new groupExtractorByDetailsSubstrings.
		return groupExtractorByDetailsSubstrings{
			intervalStats: &IntervalStatistic{
				Start:   start,
				End:     end,
				Income:  make(map[string]*Group),
				Expense: make(map[string]*Group),
			},
			groupNamesToSubstrings: groupNamesToSubstrings,
			substringsToGroupName:  substringsToGroupName,
			isGroupAllUnknown:      isGroupAllUnknownTransactions,
		}
	}, nil
}

// BuildStatisticFromInecoTransactions builds list of
// [github.com/AlexanderMakarov/aggregate-inecobank-statement.main.IntervalStatistic] from provided transactions
// with specified configuration.
func BuildStatisticFromInecoTransactions(
	transactions []InecoTransaction,
	groupExtractorBuilder GroupExtractorBuilder,
	monthStart uint,
	timeLocation *time.Location,
) ([]*IntervalStatistic, error) {

	// Sort transactions.
	sort.Sort(InecoTransactionList(transactions))

	var stats []*IntervalStatistic
	var current IntervalStatisticsBuilder

	// Get first month boundaries from the first transaction. Build first month statistics.
	monthStartDate := time.Date(transactions[0].Date.Year(), transactions[0].Date.Month(),
		int(monthStart), 0, 0, 0, 0, timeLocation)
	monthEndDate := monthStartDate.AddDate(0, 1, 0).Add(-1 * time.Nanosecond)
	current = groupExtractorBuilder(monthStartDate, monthEndDate)

	// Iterate through all the transactions.
	for _, trans := range transactions {

		// Check if this transaction is part of the new month.
		if trans.Date.After(monthEndDate) {

			// Save previous month statistic if there is one.
			stats = append(stats, current.GetIntervalStatistic())

			// Calculate start and end of the next month.
			monthStartDate = time.Date(trans.Date.Year(), trans.Date.Month(), int(monthStart), 0, 0, 0, 0, timeLocation)
			monthEndDate = monthStartDate.AddDate(0, 1, 0).Add(-1 * time.Nanosecond)
			current = groupExtractorBuilder(monthStartDate, monthEndDate)
		}

		// Handle transaction.
		if err := current.HandleTransaction(&trans); err != nil {
			return nil, err
		}
	}

	// Add last MonthStatistics.
	if current != nil {
		stats = append(stats, current.GetIntervalStatistic())
	}
	return stats, nil
}
