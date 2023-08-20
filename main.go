package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
)

type Args struct {
	FilePath         string `arg:"positional" help:"'Statement #############' CSV (not supported yet) or XML downloaded from https://online.inecobank.am"`
	MonthStart       uint   `arg:"-s" default:"1" help:"Day of month to treat as a month start. By default is 1."`
	IsDetailedOutput bool   `arg:"-v" default:"false" help:"Print detailed statistic."`
	TimeZone         string `arg:"-t" default:"Local" help:"Timezone name to use; like 'UTC', 'America/Los_Angeles'. By default is system."`
}

type FileParser interface {
	ParseRawTransactionsFromFile(args Args) ([]InecoTransaction, error)
}

func main() {
	var args Args
	argsParser := arg.MustParse(&args)

	// Check if the file path argument is provided.
	if args.FilePath == "" {
		argsParser.WriteHelp(os.Stdout)
		fmt.Println("Please provide the path to a file with 'Statement #############.xml' downloaded" +
			"from https://online.inecobank.am. First open account from which you want analyze expenses," +
			"next put into 'From' and 'To' fields dates you want to analyze, press 'Search', scroll" +
			"page to bottom and here at right corner will be 5 icons to download statement." +
			"Press XML button and save file. Next specify path to this file to the script.")
		os.Exit(1)
	}
	_, err := os.Stat(args.FilePath)
	if os.IsNotExist(err) {
		log.Fatalf("File '%s' does not exist.\n", args.FilePath)
	}

	// Parse timezone or set system.
	loc, err := time.LoadLocation(args.TimeZone)
	if err != nil {
		log.Fatalf("Unknown timzone name is specified '%s'.\n", args.TimeZone)
	}

	// Validate month start.
	if args.MonthStart < 1 || args.MonthStart > 31 {
		argsParser.WriteHelp(os.Stdout)
		fmt.Println("Error: Month start must be between 1 and 31.")
		os.Exit(1)
	}

	dotItems := strings.Split(args.FilePath, ".")
	fileExtension := dotItems[len(dotItems)-1]
	var parser FileParser
	switch fileExtension {
	case "xml":
		parser = XmlParser{}
	case "csv":
		parser = CSVParser{}
	}
	log.Printf("Going to parse with %v parser with settings %+v", parser, args)
	rawTransactions, err := parser.ParseRawTransactionsFromFile(args)
	if err != nil {
		fmt.Println("Can't parse transactions:", err)
		os.Exit(2)
	}
	if len(rawTransactions) < 1 {
		log.Fatal("Can't find transactions.")
	}
	log.Printf("Found %d transactions.", len(rawTransactions))

	// Create statistics builder.
	ge, err := NewGroupExtractorByDetailsSubstrings(
		map[string][]string{
			"Yandex Taxi":    {"YANDEX"},
			"Vika's health":  {"ARABKIR JMC"},
			"Sasha's health": {"CRYSTAL DENTAL CLINIC", "CHKA\\10 LEPSUS STR."},
			"Olya's health":  {"GEGHAMA\\ABOVYAN 34 A", "VARDANANTS"},
			"Common health":  {"DIALAB", "36.6", "NATALI FARM", "THEOPHARMA", "GEDEON RICHTER", "PHARM"},
			"Groceries":      {"CHEESE MARKET", "YEREVAN  CITY", "EVRIKA", "MARKET", "FIESTA\\19", "FIX PRICE", "MAQUR TUN", "GRAND CANDY", "MARINE GRIGORYAN", "VOSKE GAGAT", "GAYANE HAKOBYAN", "NARINE VOSKANJAN", "MIKAN GREEN", "KNARIK MKHITARYAN"},
			"Other account":  {"Account replenishment"},
			"Wildberries":    {"WILDBERRIES"},
			"Cash":           {"INECO ATM", "H.HOVHANNISYAN 24/7, ԿԱՆԽԻԿԱՑՈՒՄ"},
			"Hotels":         {"SANATORIUM"},
			"Entertainment":  {"AQUATEK", "EATERY", "TASHIR PIZZA", "KARAS", "PLAY CITY", "VICTORY\\2 AZATUTYAN AVE", "NEW CITY DIL 1\\76 MYASNIK", "INSTITUTE OF BOTANY"},
			"Subscriptions":  {"GOOGLE", "SUBSCRIPTION", "AWS EMEA", "CLOUD"},
			"Main salary":    {"ամսվա աշխատավարձ"},
			"Bonuses":        {"մԴԵՎԱՐՏՄ ՍՊԸ Արդշինբանկ  ՓԲԸ/ՊարգՅատրում"},
			"Vacation pay":   {"մԴԵՎԱՐՏՄ ՍՊԸ/Արձակուրդային վճար"},
		},
		args.IsDetailedOutput, // Make "group per uknown transaction" only if "verbose" output requested.
	)
	if err != nil {
		fmt.Println("Can't create statistic builder:", err)
		os.Exit(2)
	}

	// Build statistic.
	statistics, err := BuildStatisticFromInecoTransactions(rawTransactions, ge, args.MonthStart, loc)
	if err != nil {
		fmt.Println("Can't build statistic:", err)
		os.Exit(2)
	}

	// Process received statistics.
	for _, s := range statistics {
		if args.IsDetailedOutput {
			log.Println(s)
			continue
		}

		// Note that this logic is intentionally separate from `func (s *IntervalStatistic) String()`.
		income := MapOfGroupsToString(s.Income)
		expense := MapOfGroupsToString(s.Expense)
		log.Printf(
			"\n%s..%s:\n  Income (%d, sum=%s):%s\n  Expenses (%d, sum=%s):%s",
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
	log.Printf("Total %d month.", len(statistics))
}
