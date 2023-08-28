# aggregate-inecobank-statement
[CLI yet] tool to aggregates data from Armenian's Inecobank "statement" per month
into groups which allows to get insights into your budget.

Example of output (numbers are made up, sum may not match):
```
2023-08-01..2023-08-31:
  Income (2, sum=1,493,878.00):
    Main salary                        : 1,345,343.00
    unknown                            :   148,535.00
  Expenses (13, sum=1,020,636.38):
    Other account                      :   456,000.00
    Tom's health                       :   240,000.00
    Cash                               :   178,000.00
    Groceries                          :   112,831.00
    Hotels                             :    90,000.00
    Kate's health                      :    61,000.00
    unknown                            :    19,370.00
    Taxi                               :    17,600.00
    Entertainment                      :    14,000.00
    Subscriptions                      :     7,787.78
    Pharmacies                         :     5,957.60
    Online stores                      :     3,090.00
2023-09-01..2023-09-30:
  Income (2, sum=1,516,629.00):
...
```
Where "unknown" is group-s of "not classified yet" transactions.

Works with XML statements only, for CSV outputs failed to complete.
All Golang libraries I found are based on "encoding/csv" stdlib package which can't
parse CSV not compliant with [RFC 4180](https://www.ietf.org/rfc/rfc4180.txt).
But world is full of "don't care/know" developers and CSV-s with rows formatted as `cell1,"ce,ll2",3`.

# How to use

1. Download binary ("aggregate-inecobank-statements-\*-\*" file) for your operating system from
   [Releases](https://github.com/AlexanderMakarov/aggregate-inecobank-statement/releases) page.
2. Download "Statement ....xml" file from https://online.inecobank.am.
   Click of account from which you want analyze expenses,
   next put into 'From' and 'To' fields dates you want to analyze,
   press 'Search', scroll page to bottom and here at right corner will be 5 icons to download statement.
   Press XML button and save file as "Statement.xml" near the binary.
3. Download example of configuration
   [config.yaml](https://github.com/AlexanderMakarov/aggregate-inecobank-statement/raw/master/config.yaml).
   Don't need to update it yet, see step 5.
4. Run binary ("aggregate-inecobank-statements-\*-\*" file).
   It would open text file with list from a lot of groups where (most probably)
   a lot of names would be taken from "Details" field of transactions but some of them
   would be from the example config groups.
5. Investigate your personal transaction information and update configuration file groups with some unique
   for specific transaction substrings to aggregate transaction into these groups.
   See example in configuration file - you may remove not needed and add your own groups.
   Be careful about syntax and indentations.
6. Run binary again, and repeat configuration changes if need.
   When number of transactions in "unknown" group would decrease to small enough number
   set `detailedOutput` to `false` in configuration file to hide detalization by transactions.
   If you still want to see all these "unknown" transactions then consider to set
   `groupAllUnknownTransactions` to `false` - it will cause to put new groups with name equal to "Details" field value.
7. Run binary again - it should provide clean report for manual investigation, comparing months, etc.

# TODO
- [x] Add support of statements from different accounts.
- [x] Parse settings from the file.
- [x] Configure way to skip transactions from other file.
- [x] Provide good start config.
- [x] Add CI/CD with builds for Unix/Windows/MacOS.
- [ ] Clean extra tags from the repo.
- [ ] Update README.