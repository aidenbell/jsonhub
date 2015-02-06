package exchange

// A simple client that prints messages it gets. Useful for debugging.
// returns its client channel.
func DebugClient(die chan int) chan Messager {
	// Make the client channel and spawn
	// a goroutine to rec and print messages
	ch := make(chan Messager)
	// Goroutine to listen
	go func() {
		for {
			<-ch
			//fmt.Println("DebugClient() In:", m)
			//die <- 1
		}
	}()
	return ch
}
