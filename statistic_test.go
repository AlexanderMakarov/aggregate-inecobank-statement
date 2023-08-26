package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func Test_NewGroupExtractorByDetailsSubstrings(t *testing.T) {

	// Cases
	type args struct {
		groupNamesToSubstrings        map[string][]string
		isGroupAllUnknownTransactions bool
	}
	tests := []struct {
		name                          string
		groupNamesToSubstrings        map[string][]string
		isGroupAllUnknownTransactions bool
		expectedSubstringsToGroupName map[string]string
	}{
		{
			"no_groups_all_unknown",
			map[string][]string{},
			false,
			map[string]string{},
		},
		{
			"no_groups_1_unknown",
			map[string][]string{},
			true,
			map[string]string{},
		},
		{
			"many_groups_all_unknown",
			groups1,
			false,
			func() map[string]string {
				m := map[string]string{}
				m["a"] = "g1"
				m["b"] = "g2"
				m["c"] = "g2"
				m["d"] = "g3"
				return m
			}(),
		},
	}
	const testName = "NewGroupExtractorByDetailsSubstrings()"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Act
			builder, err := NewStatisticBuilderByDetailsSubstrings(tt.groupNamesToSubstrings,
				tt.isGroupAllUnknownTransactions, []string{})
			actualGE := builder(now, nowPlusMonth)

			// Assert
			if err != nil {
				t.Errorf("%s failed: %#v", testName, err)
			}
			if actualGE == nil {
				t.Errorf("%s builder returned null", testName)
			}
			groupNamesToSubstrings := actualGE.(groupExtractorByDetailsSubstrings).groupNamesToSubstrings
			if !reflect.DeepEqual(groupNamesToSubstrings, tt.groupNamesToSubstrings) {
				t.Errorf("%s builder set wrong groupNamesToSubstrings: expected=%+v, actual=%+v", testName,
					tt.groupNamesToSubstrings, groupNamesToSubstrings)
			}
			substringsToGroupName := actualGE.(groupExtractorByDetailsSubstrings).substringsToGroupName
			if !reflect.DeepEqual(substringsToGroupName, tt.expectedSubstringsToGroupName) {
				t.Errorf("%s builder set wrong substringsToGroupName: expected=%+v, actual=%+v", testName,
					tt.expectedSubstringsToGroupName, substringsToGroupName)
			}
		})
	}
}

func Test_groupExtractorByDetailsSubstrings_HandleTransaction(t *testing.T) {
	tI1a := newIT(1, false, "a")
	tE1b := newIT(1, true, "b")
	tI2c := newIT(2, false, "c")
	tI3d := newIT(3, false, "d")
	tI4c := newIT(4, false, "c")
	tE2b := newIT(2, true, "b")
	tE3b := newIT(3, true, "b")
	tE4c := newIT(4, true, "c")
	tI5b := newIT(5, false, "b")
	tI6e := newIT(6, false, "e")
	tE6e := newIT(6, true, "e")
	tI7f := newIT(7, false, "f")
	tE7f := newIT(7, true, "f")
	transactions1 := []*InecoTransaction{tI1a, tE1b, tI2c, tI3d, tI4c, tE2b, tE3b, tE4c, tI5b, tI6e, tE6e, tI7f, tE7f}
	// For groups1 with content:
	// "g1": {"a"},
	// "g2": {"b", "c"},
	// "g3": {"d"},
	// we should get:
	// Income: {g1 1 [tI1a], g2 11 [tI2c,tI4c,tI5b], g3 3 [tI3d], e 6 [tI6e], f 7 [tI7f]}
	// Expenses: {g1 0 [], g2 10 [tE1b,tE2b,tE3b,tE4c], g3 0 [], e 6 [tE6e], f 7 [tE7f]}

	type fields struct {
		intervalStats          *IntervalStatistic
		groupNamesToSubstrings map[string][]string
		substringsToGroupName  map[string]string
		isGroupAllUnknown      bool
	}
	newFields := func(groupNamesToSubstrings map[string][]string, isGroupAllUnknown bool) fields {
		substringsToGroupName := map[string]string{}
		for name, substrings := range groupNamesToSubstrings {
			for _, substring := range substrings {
				substringsToGroupName[substring] = name
			}
		}
		return fields{
			newIntervalStatistic(),
			groupNamesToSubstrings,
			substringsToGroupName,
			isGroupAllUnknown,
		}
	}
	tests := []struct {
		name         string
		fields       fields
		transactions []*InecoTransaction
		expected     *IntervalStatistic
	}{
		{
			name:         "no_groups_1_unknown",
			fields:       newFields(map[string][]string{}, true),
			transactions: transactions1,
			expected: func() *IntervalStatistic {
				r := newIntervalStatistic()
				r.Income = map[string]*Group{
					UnknownGroupName: groupFromITs(UnknownGroupName,
						[]*InecoTransaction{tI1a, tI2c, tI3d, tI4c, tI5b, tI6e, tI7f}),
				}
				r.Expense = map[string]*Group{
					UnknownGroupName: groupFromITs(UnknownGroupName,
						[]*InecoTransaction{tE1b, tE2b, tE3b, tE4c, tE6e, tE7f}),
				}
				return r
			}(),
		},
		{
			name:         "no_groups_all_unknown",
			fields:       newFields(map[string][]string{}, false),
			transactions: transactions1,
			expected: func() *IntervalStatistic {
				r := newIntervalStatistic()
				r.Income = map[string]*Group{
					"+1a": groupFromITs("+1a", []*InecoTransaction{tI1a}),
					"+2c": groupFromITs("+2c", []*InecoTransaction{tI2c}),
					"+3d": groupFromITs("+3d", []*InecoTransaction{tI3d}),
					"+4c": groupFromITs("+4c", []*InecoTransaction{tI4c}),
					"+5b": groupFromITs("+5b", []*InecoTransaction{tI5b}),
					"+6e": groupFromITs("+6e", []*InecoTransaction{tI6e}),
					"+7f": groupFromITs("+7f", []*InecoTransaction{tI7f}),
				}
				r.Expense = map[string]*Group{
					"-1b": groupFromITs("-1b", []*InecoTransaction{tE1b}),
					"-2b": groupFromITs("-2b", []*InecoTransaction{tE2b}),
					"-3b": groupFromITs("-3b", []*InecoTransaction{tE3b}),
					"-4c": groupFromITs("-4c", []*InecoTransaction{tE4c}),
					"-6e": groupFromITs("-6e", []*InecoTransaction{tE6e}),
					"-7f": groupFromITs("-7f", []*InecoTransaction{tE7f}),
				}
				return r
			}(),
		},
		{
			name:         "many_groups_1_unknown",
			fields:       newFields(groups1, true),
			transactions: transactions1,
			expected: func() *IntervalStatistic {
				r := newIntervalStatistic()
				r.Income = map[string]*Group{
					"g1":             groupFromITs("g1", []*InecoTransaction{tI1a}),
					"g2":             groupFromITs("g2", []*InecoTransaction{tI2c, tI4c, tI5b}),
					"g3":             groupFromITs("g3", []*InecoTransaction{tI3d}),
					UnknownGroupName: groupFromITs(UnknownGroupName, []*InecoTransaction{tI6e, tI7f}),
				}
				r.Expense = map[string]*Group{
					"g2":             groupFromITs("g2", []*InecoTransaction{tE1b, tE2b, tE3b, tE4c}),
					UnknownGroupName: groupFromITs(UnknownGroupName, []*InecoTransaction{tE6e, tE7f}),
				}
				return r
			}(),
		},
		{
			name:         "many_groups_all_unknown",
			fields:       newFields(groups1, false),
			transactions: transactions1,
			expected: func() *IntervalStatistic {
				r := newIntervalStatistic()
				r.Income = map[string]*Group{
					"g1":  groupFromITs("g1", []*InecoTransaction{tI1a}),
					"g2":  groupFromITs("g2", []*InecoTransaction{tI2c, tI4c, tI5b}),
					"g3":  groupFromITs("g3", []*InecoTransaction{tI3d}),
					"+6e": groupFromITs("+6e", []*InecoTransaction{tI6e}),
					"+7f": groupFromITs("+7f", []*InecoTransaction{tI7f}),
				}
				r.Expense = map[string]*Group{
					"g2":  groupFromITs("g2", []*InecoTransaction{tE1b, tE2b, tE3b, tE4c}),
					"-6e": groupFromITs("-6e", []*InecoTransaction{tE6e}),
					"-7f": groupFromITs("-7f", []*InecoTransaction{tE7f}),
				}
				return r
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Arrange
			handler := groupExtractorByDetailsSubstrings{
				intervalStats:          tt.fields.intervalStats,
				groupNamesToSubstrings: tt.fields.groupNamesToSubstrings,
				substringsToGroupName:  tt.fields.substringsToGroupName,
				isGroupAllUnknown:      tt.fields.isGroupAllUnknown,
			}

			// Act
			for _, trans := range tt.transactions {
				if err := handler.HandleTransaction(trans); err != nil {
					t.Errorf("groupExtractorByDetailsSubstrings.HandleTransaction() failed on %s with %#v", trans, err)
				}
			}

			// Assert
			assertIntervalStatisticEqual(tt.expected, tt.fields.intervalStats, t)
		})
	}
}

func assertIntervalStatisticEqual(expected, actual *IntervalStatistic, t *testing.T) {
	if !expected.Start.Equal(actual.Start) {
		t.Errorf("Start time does not match. Expected: %v, got: %v", expected.Start, actual.Start)
		return
	}
	if !expected.End.Equal(actual.End) {
		t.Errorf("End time does not match. Expected: %v, got: %v", expected.End, actual.End)
		return
	}
	compareGroupMap("Income", expected.Income, actual.Income, t)
	compareGroupMap("Expense", expected.Expense, actual.Expense, t)
}

func compareGroupMap(name string, expected, actual map[string]*Group, t *testing.T) {
	if len(expected) != len(actual) {
		t.Errorf("%s maps have different lengths - %d instead of %d. Expected keys: %v, Actual keys: %v",
			name, len(actual), len(expected), getMapKeys(expected), getMapKeys(actual))
		return
	}
	for k, v := range expected {
		v2, ok := actual[k]
		if !ok {
			t.Errorf("%s map missing key: %s", name, k)
			continue
		}
		if !equalGroup(v, v2) {
			t.Errorf("%s group not equal for key %s. Expected:\n%+v\nActual:\n%+v", name, k, v, v2)
		}
	}
	for k, group := range actual {
		if _, ok := expected[k]; !ok {
			t.Errorf("%s map has extra key '%s' with group %+v", name, k, group)
		}
	}
}

func getMapKeys(m map[string]*Group) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func equalGroup(a, b *Group) bool {
	if a.Name != b.Name {
		return false
	}

	// Compare totals
	if a.Total.int != b.Total.int {
		return false
	}

	// Compare transactions slice lengths
	if len(a.Transactions) != len(b.Transactions) {
		return false
	}

	// Compare each transaction
	for i, transA := range a.Transactions {
		transB := b.Transactions[i]
		if !equalTransaction(transA, transB) {
			return false
		}
	}

	return true
}

func equalTransaction(a, b *Transaction) bool {
	return a.Date == b.Date &&
		a.Details == b.Details &&
		a.Amount.int == b.Amount.int
}

var (
	now          = time.Now()
	nowPlusMonth = now.AddDate(0, 1, 0).Add(-1 * time.Nanosecond)
	groups1      = map[string][]string{
		"g1": {"a"},
		"g2": {"b", "c"},
		"g3": {"d"},
	}
)

func newIntervalStatistic() *IntervalStatistic {
	return &IntervalStatistic{
		Start:   now,
		End:     nowPlusMonth,
		Income:  make(map[string]*Group),
		Expense: make(map[string]*Group),
	}
}

func newIT(id int, isExpense bool, details string) *InecoTransaction {
	var sign, name string
	var income, expense MoneyWith2DecimalPlaces
	if isExpense {
		sign = "-"
		name = fmt.Sprintf("Expense %d", id)
		income = MoneyWith2DecimalPlaces{0}
		expense = MoneyWith2DecimalPlaces{id}
	} else {
		sign = "+"
		name = fmt.Sprintf("Income %d", id)
		income = MoneyWith2DecimalPlaces{id}
		expense = MoneyWith2DecimalPlaces{0}
	}
	return &InecoTransaction{
		Nn:                     name,
		Number:                 name,
		Date:                   now,
		Currency:               "C",
		Income:                 income,
		Expense:                expense,
		RecieverOrPayerAccount: name,
		RecieverOrPayer:        name,
		Details:                fmt.Sprintf("%s%d%s", sign, id, details),
	}
}

func groupEmpty(gn string) *Group {
	return &Group{
		Name:         gn,
		Total:        MoneyWith2DecimalPlaces{},
		Transactions: []*Transaction{},
	}
}

func groupFromITs(name string, its []*InecoTransaction) *Group {
	sum := 0
	transactions := make([]*Transaction, len(its))
	for i, t := range its {
		trans, _ := t.toTransaction()
		sum += trans.Amount.int
		transactions[i] = &trans
	}
	return &Group{
		Name:         name,
		Total:        MoneyWith2DecimalPlaces{sum},
		Transactions: transactions,
	}
}
