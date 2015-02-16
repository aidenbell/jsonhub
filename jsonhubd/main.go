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

import "github.com/aidenbell/jsonhub/exchange"

// main creates a HTTP server and adds the HTML client handler
// creates an Exchange and binds its handler to the root URL. This
// is a good place to add other handlers in future.
func main() {
	log.Println("jsonhub!")

	// Make an exchange
	ex := exchange.NewExchange()

	log.Println("Making HTTPPubSub")
	// Make a HTTP handler for the exchange
	httpBinding := exchange.NewHTTPPubSub(ex)

	// For clients polling from the exchange
	http.Handle("/client/", http.HandlerFunc(HtmlClientHandler))

	// For POSTing messages to the exchange
	http.Handle("/", httpBinding)
	log.Println("Running server http://localhost:9977")
	http.ListenAndServe(":9977", nil)
}
