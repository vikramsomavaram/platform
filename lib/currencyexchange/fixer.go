/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

/*
	Package currencyexchange provides a simple interface to the
	fixer.io API, a service for currency exchange rates.
*/

package currencyexchange

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

// Request holds the request parameters.
type Request struct {
	base     string
	protocol string
	date     string
	symbols  []string
}

// Response json response object.
type Response struct {
	Base  string `json:"base"`
	Date  string `json:"date"`
	Rates Rates  `json:"rates"`
}

// Rates currency conversion rates map
type Rates map[string]float32

var baseURL = "api.fixer.io"

// New initializes fixerio.
func New() *Request {
	return &Request{
		base:     EUR,
		protocol: "https",
		date:     "",
		symbols:  make([]string, 0),
	}
}

// Base sets base currency.
func (f *Request) Base(currency string) {
	f.base = currency
}

//Secure Make the connection secure or not by setting the
//secures argument to true or false.
func (f *Request) Secure(secure bool) {
	if secure {
		f.protocol = "https"
	} else {
		f.protocol = "http"
	}
}

// Symbols list of currencies that should be returned.
func (f *Request) Symbols(currencies ...string) {
	f.symbols = currencies
}

// Historical specify a date in the past to retrieve historical records.
func (f *Request) Historical(date time.Time) {
	f.date = date.Format("2006-01-02")
}

// GetRates retrieve the exchange rates.
func (f *Request) GetRates() (Rates, error) {
	url := f.GetURL()
	rates, err := f.makeRequest(url)

	if err != nil {
		return Rates{}, err
	}

	return rates, nil
}

// GetURL formats the URL correctly for the API Request.
func (f *Request) GetURL() string {
	var url bytes.Buffer

	url.WriteString(f.protocol)
	url.WriteString("://")
	url.WriteString(baseURL)
	url.WriteString("/")

	if f.date == "" {
		url.WriteString("latest")
	} else {
		url.WriteString(f.date)
	}

	url.WriteString("?base=")
	url.WriteString(string(f.base))

	if len(f.symbols) >= 1 {
		url.WriteString("&symbols=")
		url.WriteString(strings.Join(f.symbols, ","))
	}

	return url.String()
}

func (f *Request) makeRequest(url string) (Rates, error) {
	var response Response
	body, err := http.Get(url)

	if err != nil {
		return Rates{}, errors.New("couldn't connect to server")
	}

	defer body.Body.Close()

	err = json.NewDecoder(body.Body).Decode(&response)

	if err != nil {
		return Rates{}, errors.New("couldn't parse response")
	}

	return response.Rates, nil
}
