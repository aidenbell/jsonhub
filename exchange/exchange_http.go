package exchange

import "net/http"
import "fmt"
import "io/ioutil"
import "strings"

/*
 * This allows the exchange to be a HTTP handler, creating queues
 * for new requests and reading from the queues.
 */
func (e *Exchange) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var spec string

	// Double check path
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	flusher,ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported!", http.StatusInternalServerError)
		return;
	}

	// Look for the `q` variable in the querystring before reading the body ...
	vals := r.URL.Query()
	qvar := vals.Get("q")
	if qvar != "" {
		spec = qvar
	} else {
			
		body,err := ioutil.ReadAll(r.Body)
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
		e.Send(n)
		http.Error(w, "Sent", http.StatusOK)
		return
	}


	clientChan := make(chan Messager)

	q := NewQueue(e,spec)

	// Set some options on the queue
	if pingOnly := vals.Get("ping_only"); pingOnly == "true" {
		q.SetPingOnly(true);
	}
	
	/*
	if distMethod := vals.Get("dm"), distMethod {
		// TODO: Implement these
	}*/
	
	
	// Handle closing
	closeNotify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<- closeNotify
		q.RemoveClient(clientChan)
	}()

	e.AddQueue(q)

	// Add ourselves as a client
	go q.Run()
	q.AddClient(clientChan)
	
	// Send some headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	// Read from our client channel
	// TODO use CloseNotifier to remove client from list
	// and free those resources.
	for {
		msg := <- clientChan
		out := fmt.Sprintf(
			"event: message\ndata: %s\n\n", 
			strings.Replace(msg.Raw(), "\n", "\\n", -1))
		w.Write([]byte(out))
		flusher.Flush()
	}
	

	q.RemoveClient(clientChan)
}
