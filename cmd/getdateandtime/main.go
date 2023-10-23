package main

import (
	tad "github.com/rileys-trash-can/timeanddate"

	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: getdateandtime [location query]")
	}

	query := strings.Join(os.Args[1:], " ")
	r, err := tad.Search(query)
	if err != nil {
		log.Fatalf("Failed to query '%s': %s", query, err)
	}

	if len(r) == 0 {
		log.Fatalf("No results for '%s'", query)
	}

	// limit := 5 // max 5 results printed
	// if len(r) < limit {
	// 	limit = len(r)
	// }

	// for i := 0; i < limit; i++ {
	// 	c := r[i]
	//
	//  log.Printf(" %s (%s/%s) -> %s", c.District, c.Country, c.State, c.Path)
	// }

	result := r[0]

	log.Printf("Getting time for %s which is in %s/%s (%s)",
		result.District, result.Country, result.State, result.CountryCode)

	data, err := tad.Get(result.Path)
	if err != nil {
		log.Fatalf("Failed to get timeanddate: %s", err)
	}

	log.Printf("Country:    %s", data.Country)
	log.Printf("State:      %s", data.State)
	log.Printf("Position:   %s", data.Position)
	log.Printf("Elevation:  %d meters", data.Elevation)
	log.Printf("Currency:   %s", data.Currency)
	log.Printf("Language:   %s", data.Language)
	log.Printf("AccessCode: %s", data.AccessCode)
	log.Printf("---")
	log.Printf("Time:       %s", data.Time)
	log.Printf("Date:       %s", data.Date)
	log.Printf("  =>>       %s", data.TimeTime().Format("Mon Jan 2 15:04:05 MST 2006"))

}
