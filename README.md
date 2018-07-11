# endpoint
Simplified API access in Go. Configure an endpoint struct, then make a batch of sequential or concurrent requests.

## Overview
This package is intended to provide a streamlined way to make repeated requests to an API.

Each request takes the form:
* **Url** + **ID**

Url will be fixed for an endpoint, while ID will change for each request.

### Weather Offices, JSON API example

The sample program [examples/get-json.go](https://github.com/DavidSantia/endpoint/blob/master/examples/get-json.go) uses the National Weather Service API. For example, you can use this API to locate the San Diego weather office as follows:
* https://api.weather.gov/offices/SGX

So requests take the form
* **Url**: `https://api.weather.gov/offices/` + **ID**: `SGX`

By splitting the link into Url and ID, we can call DoConcurrent with a list of office codes, i.e.:
* AKQ
* FWD
* SGX

The program configures the endpoint as shown:
```go
ep := endpoint.Endpoint{
	Url:         "https://api.weather.gov/offices/",
	Method:      "GET",
	Headers:     map[string]string{"Content-Type": "application/json", "Accept": "*"},
	MaxParallel: 8,
	MaxRetries:  3,
	ParseFunc:   ParseOffice,
}
```

This specifies the Url, the request Method, any headers, some additional parameters, and the function
needed to parse the API response body.

This parse function is simply:
```go
func ParseOffice(b []byte, code int) (result interface{}, err error) {
    var office Office
    if code != 200 {
    	err = fmt.Errorf("status %d %s", code, http.StatusText(code))
    	return
    }
    err = json.Unmarshal(b, &office)
    result = office
    return
}
```
It parses the response body (using *json.Unmarshal*) into a struct with the expected fields and format.  By specifying
a function in the endpoint, this package can be used for any kind of data.

### Biology Definitions, HTML example

Another sample program, [examples/get-html.go](https://github.com/DavidSantia/endpoint/blob/master/examples/get-html.go),
parses an HTML page, locating a specific paragraph containing the definition of a given term.

## How to use the package

Start with a sample response from the API you are going to access.  Then create a parse function for that response.

Next, configure an endpoint.
* If an API Key is required, make sure you specify it in the Headers
* Include any Headers, such as "Accept" or "Content-Type" if required
* Add a Header function if you need dynamically calculated headers
* Make sure you specify the right Method (GET, POST, PUT, etc.)
* Add a Client to customize a request timeout

For example, a POST to an API requiring a Basic auth key and a timestamp might look like this:
```go
ep := endpoint.Endpoint{
	Url:         "https://example.com/api/", 
	Method:      "POST", 
	Headers:     map[string]string{"Authorization": "Basic ***** api key *****"}, 
	HeaderFunc:  SetDate,
	Client:      &http.Client{Timeout: 10 * time.Second},
}

func SetDate(r *http.Request) (err error) {
	date := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 -0700")
	r.Header.Add("Date", date)
	return
}
```
The timestamp is added by setting the HeaderFunc in the endpoint.

Once you have configured an endpoint, create a parse function as shown above. Then you can perform a single request
using the following function:
#### func (ep *Endpoint) DoRequest(id, data string) (result interface{}, err error)

This function will retry the request up to *ep.MaxRetries* times, before giving up.

Once you have it working for a single request, you can then call either the DoSequential() or DoConcurrent()
functions for an array of ID's.

## Sequential versus Concurrent
We can compare the easily obtained performance gain by making the requests concurrently:

```sh
$ go run examples/get-json.go
== Calling DoSequential [16 entries] ==
Elapsed: 4.493768998s
Error Rate: 0 retries, 0.00 percent

== Calling DoConcurrent [16 entries] ==
Starting requestor #1
Starting requestor #2
Starting requestor #3
Starting requestor #4
Starting requestor #5
Elapsed: 91.076842ms
Error Rate: 0 retries, 0.00 percent
```

DoConcurrent() uses the setting *ep.MaxParallel* to configure the maximum concurrency to use.
The sample programs set this to 8. This is a pretty good setting depending on the bandwith of your
internet connection, and the bandwidth of the API you are accessing.

If you set it too high, you may see the error rate of retries go up.
```
Error Rate: 10 retries, 0.14 percent
```

The DoConcurrent() function considers that there is some initial cost to launching multiple requestors.
So it calculates max as follows:
```go
max = len(ids) / 4 + 1
if max > ep.MaxParallel {
	max = ep.MaxParallel
}
```
This is why you see only 5 requestors when given 16 ID's, instead of the max possible setting.

## Notes

Although the DoConcurrent() function is efficient in terms of general parallelism, the examples have not been
optimized for connection persistence.  There is additional savings by reducing TLS handshakes.
