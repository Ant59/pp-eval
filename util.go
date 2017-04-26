package ppeval

import (
	"log"
	"strconv"
	"strings"
)

func parseFloat(f string) float64 {
	g, err := strconv.ParseFloat(strings.TrimSpace(f), 64)
	if err != nil {
		log.Fatal(err)
	}
	return g
}

func parseInt(f string) int64 {
	g, err := strconv.ParseInt(strings.TrimSpace(f), 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return g
}
