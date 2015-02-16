package exchange

import (
	"log"
	"net/http"
)
import "fmt"
import "io/ioutil"
import "strings"

// HTTPPubSub is our Subscriber type that passes received messages
// to a http.ResponseWriter using the Server Sent Events protocol
type HTTPPubSub struct {
	parent Publisher // Where to send new messages to and read messages from
}

// NewHTTPPubSub creates a new HTTP interface for PubSuber types
func NewHTTPPubSub(p Publisher) *HTTPPubSub {
	return &HTTPPubSub{p}
}

// httpChanClient is some binding between a simple message channel and
// the HTTP connection. We have to give the client pool a Subscriber
// compatible interface, so we provide this. Messages passed to receive
// are just dumped to a chan that we read from in a loop.
type httpChanClient struct {
	Incoming chan Messager
}

func (c *httpChanClient) Receive(m Messager) {
	log.Printf("httpChanClient.Receive(%T)", m)
	c.Incoming <- m
}

// ServeHTTP allows the http server to interface with a PubSuber. HTTP clients
// can publish via POST and subscribe via Server-Sent Events
func (s *HTTPPubSub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var spec string

	// Double check path
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported!", http.StatusInternalServerError)
		return
	}

	// Look for the `q` variable in the querystring before reading the body ...
	vals := r.URL.Query()
	qvar := vals.Get("q")
	if qvar != "" {
		spec = qvar
	} else {

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading MatchSpec", http.StatusInternalServerError)
			return
		}
		spec = string(body)
	}

	if r.Method == "POST" {
		// Just send the message and return
		n := NewJsonMessage(spec)
		fmt.Printf("HTTP/POST got %T -> %s (%s)\n", n, n.Raw(), spec)
		s.parent.Publish(n)
		http.Error(w, "Sent", http.StatusOK)
		return
	}

	log.Printf("HTTP: (O_O)")
	// Make a new client pool. We don't match clients joining with the same
	// spec to an existing pool (yet). This is a todo
	pool, err := NewClientPool(s.parent, spec)
	if err != nil {
		http.Error(w, "Error creating new ClientPool", http.StatusInternalServerError)
		return
	}
	s.parent.Subscribe(pool)
	client := &httpChanClient{
		make(chan Messager),
	}
	// Handle closing
	closeNotify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-closeNotify
		log.Printf("HTTP: (X_X)")
		s.parent.Unsubscribe(pool)
		pool.Unsubscribe(client)
	}()

	// Add our client binding to the bool
	log.Printf("HTTP: Subscribing %p(%T) to %p(%T)", client, client, pool, pool)
	pool.Subscribe(client)
	log.Printf(".. done")
	// Set some options on the queue
	if pingOnly := vals.Get("ping_only"); pingOnly == "true" {
		pool.SetPingOnly(true)
	}

	// Send some headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Read from our client channel
	for {
		msg := <-client.Incoming
		log.Printf("HTTP Chan got data\n")
		out := fmt.Sprintf(
			"event: message\ndata: %s\n\n",
			strings.Replace(msg.Raw(), "\n", "\\n", -1))
		w.Write([]byte(out))
		flusher.Flush()
	}
}
