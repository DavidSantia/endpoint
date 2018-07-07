# endpoint
Simplified API access in Go. Configure an enpoint struct, then make a batch of sequential or concurrent requests.

## Overview
This package is intended to provide a streamlined way to make repeated requests to an API.

Each request takes the form:
* **Url** + **ID**

URL will be fixed for an endpoint, while ID will change for each request.

For example, the National Weather Service has a JSON API. You can locate the San Diego weather office as follows:
* https://api.weather.gov/offices/SGX

The sample program [examples/get-json.go](https://github.com/DavidSantia/endpoint/blob/master/examples/get-json.go) has
a list of office codes as the IDs:
* AKQ
* FWD
* SGX

It configures the base URL in the endpoint as shown:
```go
ep := endpoint.Endpoint{
	Url:         "https://api.weather.gov/offices/",
	Method:      "GET",
	Headers:     map[string]string{"Content-Type": "application/json", "Accept": "*"},
	MaxParallel: 8,
	MaxRetries:  3,
	Parse:       ParseOffice,
}
```

Notice we are specifying the Url, the request Method, any headers, some additional parameters, and the function
needed to parse the API response body.

In our JSON weather office example, the parse function is simply:
```go
func ParseOffice(b []byte) (result interface{}, err error) {
	var office Office

	err = json.Unmarshal(b, &office)

	result = office
	return
}
```
It parses the response body using json.Unmarshal, into a struct with the expected fields and format.  By specifying
a function in the endpoint, it can be used for any kind of data.

Another sample program, [examples/get-html.go](https://github.com/DavidSantia/endpoint/blob/master/examples/get-html.go),
parses an HTML page, locating a specific paragraph containing the definition of a given term.

## How to use the package

Start with a sample response from the API you are going to access.  Then create a parse function for that response.

Next, configure an endpoint.
* If an API Key is required, make sure you specify it in the Headers
* Include any other headers, such as "Accept" or "Content-Type" if required
* Make sure you specify the right Method (GET, POST, PUT, etc.)

For example, a POST to an api requiring a Basic auth key might look like this:

```go
ep := endpoint.Endpoint{
	Url:         "https://example.com/api/",
	Method:      "POST",
	Headers:     map[string]string{"Authorization": "Basic ***** api key *****"},
}
```

Once you have configured an endpoint, create a parse function as shown above. Then you can perform a single request
using the following function:
#### func (ep *Endpoint) GetEndpoint(id string) (result interface{}, err error)

This function will retry the request up to *ep.MaxRetries* times, before giving up.

Once you have it working for a single request, you can then call either the GetSequential() or GetConcurrent()
functions for an array of ID's.

## Sequential versus Concurrent
Go language is designed for concurrency. We can compare the easily obtained performance gain by using concurrency
to make the API requests in parallel.

```sh
$ go run examples/get-json.go
== Calling GetSequential [16 entries] ==
Elapsed: 4.493768998s
Error Rate: 0 retries, 0.00 percent

== Calling GetConcurrent [16 entries] ==
Starting requestor #1
Starting requestor #2
Starting requestor #3
Starting requestor #4
Starting requestor #5
Elapsed: 91.076842ms
Error Rate: 0 retries, 0.00 percent
```

GetConcurrent() uses the setting *ep.MaxParallel* to configure the maximum concurrency to use.
The sample programs set this to 8. This is a pretty good setting depending on the bandwith of your
internet connection, and the bandwidth of the API you are accessing.

If you set it too high, you may see the error rate of retries go up.
```
Error Rate: 10 retries, 0.14 percent
```

The GetConcurrent() function considers that there is some initial cost to launching multiple requestors.
So it calculates max as follows:
```go
max = len(ids) / 4 + 1
if max > ep.MaxParallel {
	max = ep.MaxParallel
}
```
This is why you see only 5 requestors when given 16 ID's, instead of the max possible setting.
