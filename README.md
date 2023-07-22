# aggregate-inecobank-statement
[CLI yet] tool to aggregates data from Inecobank "statement" per month or specific interval.

Failed to complete. All Golang libraries I found are based on "encoding/csv" which can't
parse CSV not compliant with RFC 4180. But world is full of "don't care" developers
and CSV with rows like `cell1,"ce,ll2",3` are common.
