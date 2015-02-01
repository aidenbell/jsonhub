package main

import "log"
import "net/http"

import "github.com/aidenbell/jsonhub/exchange"

func main() {
	log.Println("JSON Message Queue")
	
	// Make an exchange
	ex := exchange.NewExchange()
	ex.Run()

	// For clients polling from the exchange
	http.Handle("/client/", http.HandlerFunc(HtmlClientHandler))

	// For POSTing messages to the exchange
	http.Handle("/", ex)
	log.Println("Running server on port 9977")
	http.ListenAndServe(":9977", nil)
}
