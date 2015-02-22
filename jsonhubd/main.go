// The jsonhubd command is a server that creates an Exchange and provides a
// simple test client for publishing and subscribing. You can find the test
// client at http://localhost:9977 after running the program.
//
// This command is pretty under-developed at this point. Running it is pretty
// simple
//
//	]$ cd $GOPATH/src/github.com/aidenbell/jsonhub/jsonhubd
//	]$ $GOPATH/bin/jsonhubd
//
// The requirement to change directory to the source location of jsonhubd
// exists because the test client's HTML is located and served from there.
// This requirement will not exist in a release version.
package main

import "log"
import "net/http"
import "github.com/gorilla/mux"
import "github.com/aidenbell/jsonhub/exchange"

// main creates a HTTP server and adds the HTML client handler
// creates an Exchange and binds its handler to the root URL. This
// is a good place to add other handlers in future.
func main() {
	router := mux.NewRouter()

	var namedExchanges = make(map[string]*exchange.HTTPPubSub)

	// Declare an internal exchange
	namedExchanges["_system"] = exchange.NewHTTPPubSub(exchange.NewExchange())

	// For clients polling from the exchange
	router.HandleFunc("/client/", HtmlClientHandler)

	// For POSTing messages to the exchange
	router.HandleFunc("/exchanges/{name}/", func(w http.ResponseWriter, r *http.Request) {
		// Lookup the right exchange
		vars := mux.Vars(r)
		requestedEx := vars["name"]
		e, found := namedExchanges[requestedEx]

		if !found {
			// Create an exchange if one wasn't found
			log.Println("Creating new exchange", requestedEx)
			newEx := exchange.NewHTTPPubSub(exchange.NewExchange())
			namedExchanges[requestedEx] = newEx
			newEx.ServeHTTP(w, r)
		} else {
			e.ServeHTTP(w, r)
		}
	})

	http.Handle("/", router)
	log.Println("Running JSONHub server on http://localhost:9977")
	http.ListenAndServe(":9977", nil)
}
