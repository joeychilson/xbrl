# xbrl

A simple parser for XBRL files.


## Usage
```go
func main() {
	file, err := os.Open("msft.xml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var xbrl xbrl.XBRL
	if err := xml.NewDecoder(file).Decode(&xbrl); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Facts:", len(xbrl.Facts))
}
```

## Example fact from parsed XBRL
```json
{
  "context": {
    "entity": "0000789019",
    "segments": [
      {
        "dimension": "DerivativeInstrumentRisk",
        "member": "InterestRateContract"
      },
      {
        "dimension": "DerivativeInstrumentsGainLossByHedgingRelationship",
        "member": "FairValueHedging"
      },
      {
        "dimension": "IncomeStatementLocation",
        "member": "NonoperatingIncomeExpense"
      }
    ],
    "period": {
      "startDate": "2023-10-01",
      "endDate": "2023-12-31"
    }
  },
  "concept": "ChangeInUnrealizedGainLossOnHedgedItemInFairValueHedge1",
  "value": -34000000,
  "decimals": "-6",
  "unit": "USD"
},
```