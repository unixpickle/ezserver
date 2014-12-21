// This is like http_server.go, but it makes sure the Wait() function works.

package main

import (
	"fmt"
	"github.com/unixpickle/ezserver"
	"net/http"
	"strconv"
	"sync"
)

func main() {
	server := ezserver.NewHTTP(http.HandlerFunc(handler))
	for {
		var portStr string
		fmt.Print("Enter port: ")
		fmt.Scanln(&portStr)
		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Println("Invalid number.")
			continue
		}
		err = server.Start(port)
		if err != nil {
			fmt.Println("Error starting server:", err)
			continue
		}
		wg := sync.WaitGroup{}
		wg.Add(3)
		for i := 0; i < 3; i++ {
			go func() {
				server.Wait()
				wg.Done()
			}()
		}
		fmt.Print("Hit enter to stop server...")
		fmt.Scanln()
		server.Stop()
		// Make sure all the Goroutines stopped.
		wg.Wait()
		// Make sure Wait() on a stopped server returns immediately.
		server.Wait()
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world!"))
}
