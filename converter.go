package main

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
)

type Meta struct {
	LastUpdate string `json:"last_updated_at"`
}

type Currency struct {
	Code  string  `json:"code"`
	Value float64 `json:"value"`
}

type Currencies struct {
	Meta Meta                `json:"meta"`
	Data map[string]Currency `json:"data"`
}

type ResultData struct {
	Currency     string
	Value        float64
	InputFrom    string
	CurrencyFrom string
	CurrencyTo   string
}

var resultData ResultData

const ApiKey = ""

const Url = "https://api.currencyapi.com/v3/latest"

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func viewHandler(writer http.ResponseWriter, request *http.Request) {
	htmlFile, err := template.ParseFiles("view.html")
	check(err)

	err = htmlFile.Execute(writer, resultData)
	check(err)
}

func convertHandler(writer http.ResponseWriter, request *http.Request) {
	currencyFrom := request.FormValue("currencyFrom")
	currencyTo := request.FormValue("currencyTo")
	inputFrom := request.FormValue("imputFrom")

	newUrl := Url + "?base_currency=" + currencyFrom + "&currencies=" + currencyTo

	client := &http.Client{}
	req, err := http.NewRequest("GET", newUrl, nil)
	check(err)
	req.Header.Add("apikey", ApiKey)

	res, err := client.Do(req)
	check(err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	check(err)

	var currencies Currencies
	err = json.Unmarshal([]byte(body), &currencies)
	check(err)

	var value float64
	var currencyCode string

	for _, currency := range currencies.Data {
		currencyCode = currency.Code
		if inputFrom != "" {
			value, err = strconv.ParseFloat(inputFrom, 64)
			check(err)
			value *= currency.Value
		}
	}

	resultData = ResultData{
		Currency:     currencyCode,
		Value:        value,
		InputFrom:    inputFrom,
		CurrencyFrom: currencyFrom,
		CurrencyTo:   currencyTo,
	}

	http.Redirect(writer, request, "/", http.StatusFound)
}

func main() {
	http.HandleFunc("/", viewHandler)
	http.HandleFunc("/convert", convertHandler)

	err := http.ListenAndServe("localhost:8080", nil)
	log.Fatal(err)
}
