package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
)

type ApiKey struct {
	ApiKey string `json:"api_key_string"`
}

func main() {

	ApiKey := getApiKey()

	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./templates/index.html"))
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./templates/fragments/results.html"))
		data := map[string][]Stock{
			"Results": SearchTicker(r.URL.Query().Get("key"), ApiKey),
		}
		tmpl.Execute(w, data)
	})

	http.HandleFunc("/stock/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			ticker := r.PostFormValue("ticker")
			stk := SearchTicker(ticker, ApiKey)[0]
			val := GetDailyValues(ticker, ApiKey)
			tmpl := template.Must(template.ParseFiles("./templates/index.html"))
			tmpl.ExecuteTemplate(w, "stock-element",
				Stock{Ticker: stk.Ticker, Name: stk.Name, Price: val.Open})
		}
	})

	log.Println("App running on 8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func getApiKey() string {
	content, err := os.ReadFile("./key.json")
	if err != nil {
		log.Fatal(err)
	}

	var payload ApiKey
	err = json.Unmarshal(content, &payload)
	if err != nil {
		log.Fatal(err)
	}
	return payload.ApiKey
}
