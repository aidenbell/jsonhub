package main

import "net/http"
import "os"
import "io"

// HTTP handler for the test client. Just returns a simple HTML/Javascript
// combo that can send messages to the queue and listen for messages.
func HtmlClientHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("test_client.html")
	if err != nil {
		http.Error(w, "Error reading client file", http.StatusInternalServerError)
		return
	}
	io.Copy(w, file)
	file.Close()
	return
}
