package main

import "log"
import "net/http"

import "github.com/aidenbell/jsonhub/exchange"

// main creates a HTTP server and adds the HTML client handler
// creates an Exchange and binds its handler to the root URL. This
// is a good place to add other handlers in future.
func main() {
	log.Println("jsonhub!")

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
