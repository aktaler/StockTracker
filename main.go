package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

var ApiKeyString string
var sessions = make(map[string][]Stock)

type ApiKey struct {
	ApiKey string `json:"api_key_string"`
}

func main() {

	getApiKey(&ApiKeyString)

	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", sessionHandler)

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./templates/fragments/results.html"))
		data := map[string][]Stock{
			"Results": SearchTicker(r.URL.Query().Get("key"), ApiKeyString),
		}
		tmpl.Execute(w, data)
	})

	http.HandleFunc("/stock/", func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "Session not found", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case "POST":
			ticker := r.PostFormValue("ticker")
			stk := SearchTicker(ticker, ApiKeyString)[0]
			val := GetDailyValues(ticker, ApiKeyString)

			// Add the stock to the session
			sessionData := sessions[sessionID.Value]
			sessionData = append(sessionData, Stock{Ticker: stk.Ticker, Name: stk.Name, Price: val.Open})
			sessions[sessionID.Value] = sessionData

			tmpl := template.Must(template.ParseFiles("./templates/index.html"))
			tmpl.ExecuteTemplate(w, "stock-element", sessionData[len(sessionData)-1])
		}
	})

	log.Println("App running on 8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func getApiKey(key *string) {
	content, err := os.ReadFile("./key.json")
	if err != nil {
		log.Fatal(err)
	}

	var payload ApiKey
	err = json.Unmarshal(content, &payload)
	if err != nil {
		log.Fatal(err)
	}
	*key = payload.ApiKey
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		sessionID = &http.Cookie{
			Name:  "session_id",
			Value: uuid.NewString(),
			Path:  "/",
		}
		http.SetCookie(w, sessionID)
	}

	sessionData, exists := sessions[sessionID.Value]
	// Check if the session exists in the sessions map
	if !exists {
		sessionData = make([]Stock, 0)
		sessions[sessionID.Value] = sessionData
	}

	tmpl := template.Must(template.ParseFiles("./templates/index.html"))
	// better use a fitting struct to pass the sessionData so different elements
	// in the html can be targeted
	tmpl.Execute(w, sessionData)
}
