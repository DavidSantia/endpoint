package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DavidSantia/endpoint"
	"github.com/PuerkitoBio/goquery"
)

func main() {
	var ids []string
	var definition string
	var result interface{}
	var results []interface{}
	var tStart time.Time

	ids = []string{
		"Fruit",
		"Seed",
		"Angiosperms",
		"Ovary",
		"Flowering",
		"Edible",
		"Plant",
		"Offspring",
		"Union",
		"Flower",
		"Pineapple",
		"Strawberry",
		"Mulberry",
		"Apple",
		"Pear",
		"Seed",
		"Animal",
		"Sugar",
		"Starch",
	}

	ep := endpoint.Endpoint{
		Url:    "https://www.biology-online.org/dictionary/",
		Method: "GET",
		Client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: &http.Transport{TLSHandshakeTimeout: 5 * time.Second},
		},
		MaxParallel: 8,
		MaxRetries:  3,
		Parse:       ParseDefinition,
	}

	ep.Retries = 0
	fmt.Printf("== Calling DoSequential [%d entries] ==\n", len(ids))
	tStart = time.Now()
	results = ep.DoSequential(ids)
	fmt.Printf("Elapsed: %v\n", time.Now().Sub(tStart))
	fmt.Printf("Error Rate: %d retries, %.2f percent\n\n",
		ep.Retries, float32(ep.Retries*100)/float32(len(ids)))

	ep.Retries = 0
	fmt.Printf("== Calling DoConcurrent [%d entries] ==\n", len(ids))
	tStart = time.Now()
	results = ep.DoConcurrent(ids)
	fmt.Printf("Elapsed: %v\n", time.Now().Sub(tStart))
	fmt.Printf("Error Rate: %d retries, %.2f percent\n\n",
		ep.Retries, float32(ep.Retries*100)/float32(len(ids)))

	fmt.Printf("== Results ==\n")
	for _, result = range results {
		definition = result.(string)
		fmt.Printf("* %s\n", definition)
	}
}

func ParseDefinition(b []byte) (result interface{}, err error) {
	var s, text string
	var doc *goquery.Document
	var done bool

	doc, err = goquery.NewDocumentFromReader(bytes.NewReader(b))
	if err != nil {
		return
	}

	doc.Find("div#mw-content-text p").Each(func(i int, row *goquery.Selection) {
		s = strings.TrimSpace(row.Text())
		if s == "Supplement" {
			done = true
		}
		if !done {
			text += s + "\n"
		}
	})

	result = text
	return
}
