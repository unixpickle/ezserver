package main

import (
	"fmt"
	"github.com/unixpickle/ezserver"
	"net/http"
	"strconv"
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
		fmt.Print("Hit enter to stop server...")
		fmt.Scanln()
		server.Stop()
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world!"))
}
