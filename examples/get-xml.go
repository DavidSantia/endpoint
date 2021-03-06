package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/DavidSantia/endpoint"
)

type Category struct {
	Id   int    `xml:"id"`
	Name string `xml:"name"`
}
type CategoryList struct {
	XMLName    xml.Name   `xml:"response"`
	Categories []Category `xml:"data>categories>category"`
}

func main() {
	var err error
	var id string
	var result interface{}
	var tStart time.Time

	id = "list"

	ep := endpoint.Endpoint{
		Url:         "http://thecatapi.com/api/categories/",
		Method:      "GET",
		Headers:     map[string]string{"Content-Type": "text/xml"},
		Client:      &http.Client{Timeout: 10 * time.Second},
		MaxParallel: 8,
		MaxRetries:  3,
		ParseFunc:   ParseCategoryList,
	}

	ep.Retries = 0
	fmt.Printf("== Calling DoRequest ==\n")
	tStart = time.Now()
	result, err = ep.DoRequest(id, "")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Elapsed: %v\n", time.Now().Sub(tStart))
	fmt.Printf("Error Rate: %d retries\n\n", ep.Retries)

	if result != nil {
		clist := result.(CategoryList)
		for _, category := range clist.Categories {
			fmt.Printf("* Id: %2d, Name: %s\n", category.Id, category.Name)
		}
	}

	fmt.Printf("\n== Error test ==\n")
	result, err = ep.DoRequest("foo", "")
	fmt.Printf("Error: %v\n", err)
}

func ParseCategoryList(b []byte, code int) (result interface{}, err error) {
	var clist CategoryList

	if code != 200 {
		err = fmt.Errorf("status %d %s", code, http.StatusText(code))
		return
	}

	err = xml.Unmarshal(b, &clist)
	result = clist
	return
}
