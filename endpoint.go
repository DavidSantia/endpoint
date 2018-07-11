package endpoint

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type Endpoint struct {
	Url         string
	Method      string
	Headers     map[string]string
	HeaderFunc  func(*http.Request) error
	Client      *http.Client
	MaxParallel int
	MaxRetries  int
	Retries     int
	ParseFunc   func([]byte, int) (interface{}, error)
}

func (ep *Endpoint) DoSequential(ids []string) (results []interface{}) {
	var err error
	var id string
	var result interface{}

	for _, id = range ids {
		result, err = ep.DoRequest(id, "")
		if err != nil {
			result = err.Error()
		}
		results = append(results, result)
	}
	return
}

func (ep *Endpoint) DoConcurrent(ids []string) (results []interface{}) {
	var err error
	var id string
	var result interface{}
	var i, max int

	// Set max parallelism
	max = len(ids)/4 + 1
	if ep.MaxParallel > 0 && max > ep.MaxParallel {
		max = ep.MaxParallel
	}

	// make input and output channels
	inputChan := make(chan string, max*2)
	outputChan := make(chan interface{}, max)

	// Load ids into channel
	go func() {
		for _, id = range ids {
			inputChan <- id
		}
	}()

	// Start concurrent requestors
	for i = 1; i <= max; i++ {
		fmt.Printf("Starting requestor #%d\n", i)
		go func() {
			for {
				id = <-inputChan
				result, err = ep.DoRequest(id, "")
				if err != nil {
					outputChan <- err.Error()
				} else {
					outputChan <- result
				}
			}
		}()
	}

	// Retrieve all results
	for i = 0; i < len(ids); i++ {
		result = <-outputChan
		results = append(results, result)
	}

	close(inputChan)
	close(outputChan)
	return
}

func (ep *Endpoint) DoRequest(id, data string) (result interface{}, err error) {
	var client *http.Client
	var req *http.Request
	var res *http.Response
	var i int
	var b []byte
	var rdr io.Reader

	// Validate endpoint parameters
	if ep.ParseFunc == nil {
		panic("DoRequest requires Endpoint.ParseFunc")
	}
	if len(ep.Method) == 0 {
		panic("DoRequest requires Endpoint.Method")
	}

	// Create request from endpoint
	if len(data) > 0 {
		rdr = strings.NewReader(data)
	}
	req, err = http.NewRequest(ep.Method, ep.Url+id, rdr)

	// Add headers
	for key, value := range ep.Headers {
		req.Header.Add(key, value)
	}

	// Call custom header function
	if ep.HeaderFunc != nil {
		err = ep.HeaderFunc(req)
		if err != nil {
			err = fmt.Errorf("custom HeaderFunc: %v", err)
			return
		}
	}

	// Configure client
	if ep.Client == nil {
		client = http.DefaultClient
	} else {
		client = ep.Client
	}

	// Make request, retry if required
	for i = 0; i <= ep.MaxRetries; i++ {
		if i > 0 {
			fmt.Printf("Warn: Retry #%d for %q\n", i, id)
			ep.Retries++
		}
		res, err = client.Do(req)
		if err == nil {
			b, err = ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		err = fmt.Errorf("failure for %q %v", id, err)
		return
	}

	result, err = ep.ParseFunc(b, res.StatusCode)
	if err != nil {
		err = fmt.Errorf("failure for %q %v", id, err)
		return
	}
	return
}
