# aggregate-inecobank-statement
Local tool to aggregates data from Armenian's [Inecobank](https://online.inecobank.am)
"statements" from multiple accounts monthly, into groups which allows to get insights into your budget.

Example of output (numbers are made up, sum may not match):
```
2023-08-01..2023-08-31:
  Income (2, sum=1,493,878.00):
    Main salary                        : 1,345,343.00
  Expenses (13, sum= 920,636.38):
    Rent                               :   300,000.00
    Tom's health                       :   240,000.00
    Cash withdrowal                    :   178,000.00
    Groceries                          :   112,831.00
    Kindergarten                       :    90,000.00
    Kate's health                      :    61,000.00
    Taxi                               :    17,600.00
    Entertainment                      :    14,000.00
    Subscriptions                      :     7,787.78
    Pharmacies                         :     5,957.60
    Online stores                      :     3,090.00
2023-09-01..2023-09-30:
  Income (2, sum=1,516,629.00):
...
```

# How to use

1. Download binary ("aggregate-inecobank-statements-\*-\*" file) for your operating system from
   [Releases](https://github.com/AlexanderMakarov/aggregate-inecobank-statement/releases) page.
2. Download "Statement ....xml" files from https://online.inecobank.am for interesting period and
   put them near the "aggregate-inecobank-statements-\*-\*" file.
   In details, on [main page](https://online.inecobank.am) click on the chosen account,
   specify into 'From' and 'To' fields dates you want to analyze,
   press 'Search', scroll page to bottom and here at right corner will be 5 icons to download statement.
   Press XML button and save near "aggregate-inecobank-statements-\*-\*" file.
3. Download example of configuration
   [config.yaml](https://raw.githubusercontent.com/AlexanderMakarov/aggregate-inecobank-statement/master/config.yaml).
   Don't need to update it yet, see step 5.
4. Run binary ("aggregate-inecobank-statements-\*-\*" file).
   It would open text file with the list from a lot of groups where (most probably)
   a lot of names would be taken from "Details" field of transactions but some of them
   would be from the example config groups.
5. Investigate your personal transaction information and update configuration file groups with unique
   for specific transaction substrings to aggregate transaction into these groups.
   See examples in configuration file - you may remove not needed and add your own groups.
   Be careful about syntax and indentations, in case of error resulting file would contain error description.
6. Run binary again, and repeat configuration changes if need.
   When number of transactions in "unknown" group would decrease to small enough number
   set `detailedOutput` to `false` in configuration file to hide detalization by transactions.
   If you still want to see all these "unknown" transactions then consider to set
   `groupAllUnknownTransactions` to `false` - it will cause to put new groups with name equal to "Details" field value.
7. Run binary again - it should provide clean report for manual investigation, comparing months, etc.
   On new month it is enough to downloan "Statements" with new transactions and run binary again.

# TODO
- [x] Add support of statements from different accounts.
- [x] Parse settings from the file.
- [x] Configure way to skip transactions from other file.
- [x] Provide good start config.
- [x] Add CI/CD with builds for Unix/Windows/MacOS.
- [x] Clean extra tags from the repo.
- [x] Update README.
- [ ] Create video "How to use".