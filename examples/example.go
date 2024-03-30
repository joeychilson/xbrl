package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joeychilson/xbrl"
)

func main() {
	file, err := os.Open("msft.xml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	startTime := time.Now()

	var xbrl xbrl.XBRL
	if err := xml.NewDecoder(file).Decode(&xbrl); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Parsed in %s\n", time.Since(startTime))

	fmt.Println("Facts:", len(xbrl.Facts))

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
