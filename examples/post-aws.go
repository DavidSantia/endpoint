package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DavidSantia/endpoint"
)

const (
	// AWS info
	AWS_SES_ENDPOINT  = "https://email.us-east-1.amazonaws.com"
	AWS_ACCESS_KEY_ID = "********************"
	AWS_SECRET_KEY    = "****************************************"

	// Test Email
	FROM_EMAIL = "\"endpoint package\" <logging@example.com>"
)

type EmailSuccess struct {
	XMLName   xml.Name `xml:"SendEmailResponse"`
	MessageId string   `xml:"SendEmailResult>MessageId"`
	RequestId string   `xml:"ResponseMetadata>RequestId"`
}

type EmailError struct {
	XMLName   xml.Name `xml:"ErrorResponse"`
	Type      string   `xml:"Error>Type"`
	Code      string   `xml:"Error>Code"`
	Message   string   `xml:"Error>Message"`
	RequestId string   `xml:"RequestId"`
}

func main() {
	var err error
	var tStart time.Time

	ep := endpoint.Endpoint{
		Url:    AWS_SES_ENDPOINT,
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		HeaderFunc: HeaderSES,
		Client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: &http.Transport{TLSHandshakeTimeout: 5 * time.Second},
		},
		ParseFunc: ParseSES,
	}

	var to, cc, subject, body string
	to = "testuser@example.com"
	subject = "Test email using SES"
	body = "This is a test of the endpoint package for posting to SES."

	ep.Retries = 0
	fmt.Printf("== Calling SendEmail ==\n")
	tStart = time.Now()
	err = SendEmail(ep, to, cc, subject, body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Elapsed: %v\n", time.Now().Sub(tStart))
	fmt.Printf("Error Rate: %d retries\n\n", ep.Retries)
}

func SendEmail(ep endpoint.Endpoint, to, cc, subject, body string) (err error) {
	var result interface{}

	data := make(url.Values)
	data.Add("Action", "SendEmail")
	data.Add("Source", FROM_EMAIL)
	data.Add("Destination.ToAddresses.member.1", to)
	if len(cc) > 0 {
		data.Add("Destination.CcAddresses.member.1", cc)
	}
	data.Add("Message.Subject.Data", subject)
	data.Add("Message.Body.Text.Data", body)
	data.Add("AWSAccessKeyId", AWS_ACCESS_KEY_ID)

	result, err = ep.DoRequest("", data.Encode())
	if err != nil {
		return
	}

	// If success, log recipient and MessageId
	log.Printf("Sent email to %s, MessageId %s\n", to, result)
	return
}

func HeaderSES(r *http.Request) (err error) {

	// Add date
	date := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 -0700")
	r.Header.Add("Date", date)

	// Add signature
	h := hmac.New(sha256.New, []uint8(AWS_SECRET_KEY))
	h.Write([]uint8(date))
	r.Header.Add("X-Amzn-Authorization",
		fmt.Sprintf("AWS3-HTTPS AWSAccessKeyId=%s, Algorithm=HmacSHA256, Signature=%s",
			AWS_ACCESS_KEY_ID, base64.StdEncoding.EncodeToString(h.Sum(nil))))
	return
}

func ParseSES(b []byte, code int) (result interface{}, err error) {
	response := string(b)

	// Interpret response
	if code != 200 {
		if len(response) < 15 {
			// Not enough data to parse response
			err = fmt.Errorf("failure AWS SES status %d: %s", code, http.StatusText(code))
			return
		} else if response[1:14] == "ErrorResponse" {
			// Show AWS Error Type, Code and Message
			errResp := EmailError{}
			err = xml.Unmarshal(b, &errResp)
			if err != nil {
				err = fmt.Errorf("decoding AWS SES response: %v", err)
				return
			}
			err = fmt.Errorf("failure AWS SES %s %s: %s", strings.TrimSpace(errResp.Type),
				strings.TrimSpace(errResp.Code), strings.TrimSpace(errResp.Message))
			return
		} else if response[1:14] != "SendEmailResp" {
			err = fmt.Errorf("unrecognized AWS SES response: %s", response)
			return
		}
	}

	success := EmailSuccess{}
	err = xml.Unmarshal(b, &success)
	if err != nil {
		err = fmt.Errorf("decoding AWS SES response: %v", err)
		return
	}
	result = success.MessageId
	return
}
