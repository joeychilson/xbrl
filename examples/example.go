package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joeychilson/xbrl"
)

func main() {
	file, err := os.ReadFile("msft.xml")
	if err != nil {
		log.Fatal(err)
	}

	startTime := time.Now()

	xbrl, err := xbrl.Parse(string(file))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Parsed XBRL in %s\n", time.Since(startTime))

	bytes, err := json.Marshal(xbrl)
	if err != nil {
		log.Fatal(err)
	}

	jsonFile, err := os.Create("msft.json")
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	jsonFile.Write(bytes)
}
