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

TBD
1. Download binary for your operating system from
2. Download "Statement ....xml" file from https://online.inecobank.am.
   Click of account from which you want analyze expenses,
   next put into 'From' and 'To' fields dates you want to analyze,
   press 'Search', scroll page to bottom and here at right corner will be 5 icons to download statement.
   Press XML button and save file as "Statement.xml" near the binary.
3. Download example of configuration. Don't need to update it yet, see step 5.
4. Run binary. It would provide list with a lot of groups where (most probably)
   a lot of names would be taken from "Details" field of transactions but some of them
   would be from example config.
5. Update configuration file with some unique for specific transaction substrings to aggregate transaction into.
   See example in configuration file - you may remove not needed and add your own groups.
6. Run binary again, and repeat configuration changes if need.
   When number of groups with "from Details" names would decrease to small enough number
   set `groupAllUnknownTransactions` to `true` and `detailedOutput` to `false` in configuration file.
   It would cause to put all transactions which not matched by `grouping` substrings into one "unknown" group
   and remove output of transactions. Now binary would provide good and useful information from your statements.
7. Save configuration file for the future use. With time it may require some updates.

# TODO
- [x] Add support of statements from different accounts.
- [x] Parse settings from the file.
- [ ] Provide good start config.
- [ ] Add CI/CD with builds for Unix/Windows/MacOS.
- [ ] Update README.