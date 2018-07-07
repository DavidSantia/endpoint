package main

import (
	"fmt"
	"time"

	"github.com/DavidSantia/endpoint"
	"encoding/json"
)

type Office struct {
	Id                          string   `json:"id"`
	Name                        string   `json:"name"`
	Address                     Address  `json:"address"`
	Telephone                   string   `json:"telephone"`
	FaxNumber                   string   `json:"faxNumber"`
	Email                       string   `json:"email"`
	SameAs                      string   `json:"sameAs"`
	NwsRegion                   string   `json:"nwsRegion"`
	ParentOrganization          string   `json:"parentOrganization"`
	ResponsibleCounties         []string `json:"responsibleCounties"`
	ResponsibleForecastZones    []string `json:"responsibleForecastZones"`
	ResponsibleFireZones        []string `json:"responsibleFireZones"`
	ApprovedObservationStations []string `json:"approvedObservationStations"`
}

type Address struct {
	StreetAddress   string `json:"streetAddress"`
	AddressLocality string `json:"addressLocality"`
	AddressRegion   string `json:"addressRegion"`
	PostalCode      string `json:"postalCode"`
}

func main() {
	var offices []string
	var office Office
	var result interface{}
	var results []interface{}
	var tStart time.Time

	offices = []string{
		"AKQ",
		"FWD",
		"SGX",
		"BGM",
		"JKL",
		"RLX",
		"PQR",
		"FGZ",
		"GSP",
		"LBF",
		"FSD",
		"MTR",
		"LOX",
		"PBZ",
		"GRR",
		"CTP",
	}

	ep := endpoint.Endpoint{
		Url:         "https://api.weather.gov/offices/",
		Method:      "GET",
		Headers:     map[string]string{"Content-Type": "application/json", "Accept": "*"},
		MaxParallel: 8,
		MaxRetries:  3,
		Parse:       ParseOffice,
	}

	ep.Retries = 0
	fmt.Printf("== Calling GetSequential [%d entries] ==\n", len(offices))
	tStart = time.Now()
	results = ep.GetSequential(offices)
	fmt.Printf("Elapsed: %v\n", time.Now().Sub(tStart))
	fmt.Printf("Error Rate: %d retries, %.2f percent\n\n",
		ep.Retries, float32(ep.Retries)/float32(len(offices)))

	ep.Retries = 0
	fmt.Printf("== Calling GetConcurrent [%d entries] ==\n", len(offices))
	tStart = time.Now()
	results = ep.GetConcurrent(offices)
	fmt.Printf("Elapsed: %v\n", time.Now().Sub(tStart))
	fmt.Printf("Error Rate: %d retries, %.2f percent\n\n",
		ep.Retries, float32(ep.Retries)/float32(len(offices)))

	fmt.Printf("== Results ==\n")
	for _, result = range results {
		office = result.(Office)
		fmt.Printf("* %s\n", office.Name)
		fmt.Printf("  %v\n", office.Address)
		fmt.Printf("  %s\n", office.Telephone)
		fmt.Printf("  %d Counties, %d ForecastZones, %d FireZones\n\n",
			len(office.ResponsibleCounties), len(office.ResponsibleForecastZones), len(office.ResponsibleFireZones))
	}
}

func ParseOffice(b []byte) (result interface{}, err error) {
	var office Office

	err = json.Unmarshal(b, &office)

	result = office
	return
}
