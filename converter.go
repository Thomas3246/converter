package main

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
)

type Meta struct {
	LastUpdate string `json:"last_updated_at"` //Types for JSON encoding
}

type Currency struct {
	Code  string  `json:"code"` //Types for JSON encoding
	Value float64 `json:"value"`
}

type Currencies struct {
	Meta Meta                `json:"meta"` //Types for JSON encoding
	Data map[string]Currency `json:"data"`
}

type ResultData struct { //Type for exchanging data to HTML file
	Currency     string
	Value        float64
	InputFrom    string
	CurrencyFrom string
	CurrencyTo   string
}

var resultData ResultData //Global var to exchange data to HTML

const ApiKey = "" //Here must be API key from currency API

const Url = "https://api.currencyapi.com/v3/latest"

func check(err error) { //Function to check errors
	if err != nil {
		log.Fatal(err)
	}
}

func viewHandler(writer http.ResponseWriter, request *http.Request) { //Main handler function
	htmlFile, err := template.ParseFiles("assets/view.html")
	check(err)

	err = htmlFile.Execute(writer, resultData) //Sending ResulData type to the HTML with received values
	check(err)
}

func convertHandler(writer http.ResponseWriter, request *http.Request) { //Function-converter that makes calculations
	currencyFrom := request.FormValue("currencyFrom")
	currencyTo := request.FormValue("currencyTo") //Here we're getting input values
	inputFrom := request.FormValue("imputFrom")

	newUrl := Url + "?base_currency=" + currencyFrom + "&currencies=" + currencyTo //making new URL to the currency api site from one currency to other

	client := &http.Client{} //Making request to currency api site with our API key
	req, err := http.NewRequest("GET", newUrl, nil)
	check(err)
	req.Header.Add("apikey", ApiKey)

	res, err := client.Do(req)
	check(err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	check(err)

	var currencies Currencies
	err = json.Unmarshal([]byte(body), &currencies) //Reading JSON answer with value
	check(err)

	var value float64
	var currencyCode string

	for _, currency := range currencies.Data { //Had to use for range because we don't know exact currency
		currencyCode = currency.Code
		if inputFrom != "" { //If user didn't enter value to input field, we do NOT parsing this value
			value, err = strconv.ParseFloat(inputFrom, 64) //Parsing string to float64
			check(err)
			value *= currency.Value //Getting new value that equals input value * exchange rate
		}
	}

	resultData = ResultData{ //Setting received values to the struct to send it to the HTML file
		Currency:     currencyCode,
		Value:        math.Round(value*100) / 100,
		InputFrom:    inputFrom,
		CurrencyFrom: currencyFrom,
		CurrencyTo:   currencyTo,
	}

	http.Redirect(writer, request, "/", http.StatusFound) //Redirecting back to the main page
}

func main() {
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	http.HandleFunc("/", viewHandler)
	http.HandleFunc("/convert", convertHandler)

	err := http.ListenAndServe("localhost:8080", nil)
	log.Fatal(err)
}
