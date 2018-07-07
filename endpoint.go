package endpoint

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type Endpoint struct {
	Url         string
	Method      string
	Headers     map[string]string
	MaxParallel int
	MaxRetries  int
	Retries     int
	Parse       func(b []byte) (interface{}, error)
}

func (ep *Endpoint) GetSequential(ids []string) (results []interface{}) {
	var err error
	var id string
	var result interface{}

	for _, id = range ids {
		result, err = ep.GetEndpoint(id)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			results = append(results, result)
		}
	}
	return
}

func (ep *Endpoint) GetConcurrent(ids []string) (results []interface{}) {
	var err error
	var id string
	var result interface{}
	var i, max int

	// make input and output channels
	inputChan := make(chan string, ep.MaxParallel*2)
	outputChan := make(chan interface{})

	// Load ids into channel
	go func() {
		for _, id = range ids {
			inputChan <- id
		}
	}()

	// Set max parallelism
	max = len(ids) / 4 + 1
	if max > ep.MaxParallel {
		max = ep.MaxParallel
	}

	// Start concurrent requestors
	for i = 1; i <= max; i++ {
		fmt.Printf("Starting requestor #%d\n", i)
		go func() {
			for {
				id = <-inputChan
				result, err = ep.GetEndpoint(id)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					outputChan <- "Error"
				}

				outputChan <- result
			}
		}()
	}

	// Retrieve all results
	for i = 0; i < len(ids); i++ {
		result = <- outputChan
		results = append(results, result)
	}

	close(inputChan)
	close(outputChan)
	return
}

func (ep *Endpoint) GetEndpoint(id string) (result interface{}, err error) {
	var req *http.Request
	var res *http.Response
	var i int
	var b []byte

	// Create request from endpoint
	req, err = http.NewRequest(ep.Method, ep.Url+id, nil)

	// Add headers
	for key, value := range ep.Headers {
		req.Header.Add(key, value)
	}

	// Make request, retry if required
	for i = 0; i < ep.MaxRetries; i++ {
		res, err = http.DefaultClient.Do(req)
		if err == nil {
			b, err = ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err == nil {
				break
			}
		}
		fmt.Printf("Warn: Retrying %s\n", id)
		ep.Retries++
	}
	if err != nil {
		err = fmt.Errorf("http %s request: %v", ep.Method, err)
		return
	}

	result, err = ep.Parse(b)
	if err != nil {
		err = fmt.Errorf("http %s response: %v", ep.Method, err)
		return
	}
	return
}
