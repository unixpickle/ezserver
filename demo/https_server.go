package main

import (
	"fmt"
	"github.com/unixpickle/ezserver"
	"net/http"
	"strconv"
)

var SERVER_KEY = `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCn/nB2th1O/M/5qwSelG4Uk7AM/lJWIBBz3nw7Ga1+lI4nNAo8
F/5fCmy7OK10Zt58l94zTe3ZiHvfhvq2jYa2oFqG6Bm5mjO6BmI8FFNJN2ovelUV
0SBzhuFiEAa3ZuQljuj9I5vICtau9LZRzaRjpGNAMh5XHZzqgzJOe2+mxQIDAQAB
AoGAPAIFRkJTQc3ItJREOCkpESyYLGwEGUOm3NzSX4ISmS6TgKl0Jncjo+tjX5Ul
UHkWbEcLViQ2HAhGx1e94su3HJzWjCMrYcNNtz3k5B6y8j3AhLhPiWQ4+oY5x0yh
NzgPn8BEo7tGZvNuEGDWa/qZMjZAfsBYYBq+DgzUfM78DsECQQDbBWNpITRdzrU5
PZlCnAcsxRObbXb6Kkx48yEoR/ud+QBHjz4/iSt3eN0ryD4P2glrRre/CZU/8tvM
f88VhoL9AkEAxFuJQ8sKW29sYvrXI5xvEPqvVnh/SFCaUsFhI32w23S3WzSFb84W
Wu2flMJITt6pZvIBO93KYiKaC0nTt2+xaQJAH105LC/mGNzmFMlebix71oxuT16w
oAh4pQVkJSmRvcCPqq+3oU+aWuSC/6cQRCLcIHGjFIdhySOVGEbhN9roXQJAH7Cq
QZ+2RzV/Z6YWLLAlmLbsr2b5G+GuVmbRV5oEfhajNPwQARBguUIafDay1s/GxU+P
dWsBK79r3yCGI9fJ6QJBAIZckHyxv+uCpjDW4b8D5WKT63Xqq7sC+mrTUW6KCQgf
k0pKcsPW67oVB93bjCaZBAx8BMfqLVsR6pj1nxf6E7I=
-----END RSA PRIVATE KEY-----`

var SERVER_CERT = `-----BEGIN CERTIFICATE-----
MIICXzCCAcgCCQDwoffrCN89uzANBgkqhkiG9w0BAQUFADB0MQswCQYDVQQGEwJV
UzEVMBMGA1UECBMMUGVubnN5bHZhbmlhMRUwEwYDVQQHEwxQaGlsYWRlbHBoaWEx
DTALBgNVBAoTBERlbW8xFDASBgNVBAsTC1Byb2dyYW1taW5nMRIwEAYDVQQDEwls
b2NhbGhvc3QwHhcNMTQxMjIxMTYyODQ0WhcNMTcwOTE2MTYyODQ0WjB0MQswCQYD
VQQGEwJVUzEVMBMGA1UECBMMUGVubnN5bHZhbmlhMRUwEwYDVQQHEwxQaGlsYWRl
bHBoaWExDTALBgNVBAoTBERlbW8xFDASBgNVBAsTC1Byb2dyYW1taW5nMRIwEAYD
VQQDEwlsb2NhbGhvc3QwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAKf+cHa2
HU78z/mrBJ6UbhSTsAz+UlYgEHPefDsZrX6Ujic0CjwX/l8KbLs4rXRm3nyX3jNN
7dmIe9+G+raNhragWoboGbmaM7oGYjwUU0k3ai96VRXRIHOG4WIQBrdm5CWO6P0j
m8gK1q70tlHNpGOkY0AyHlcdnOqDMk57b6bFAgMBAAEwDQYJKoZIhvcNAQEFBQAD
gYEAUGKiESUiBAmhJvxveX1zHc2w+5p8oqjHhaXnXyk+g4YDh4G7TKULAOagQpEQ
Tu2WDupr3MOnNJy+Ir8WxJA+C0Pol/Ifg4iTG7YWxuO7UZKkLm1MhrumCYJfTWx6
ZaqigGnpcK95++/VkLxI8sDZgKRfCUBF5X1rVh9YC1LDzrE=
-----END CERTIFICATE-----`

func main() {
	config := new(ezserver.TLSConfig)
	config.Default.Key = SERVER_KEY
	config.Default.Certificate = SERVER_CERT
	server := ezserver.NewHTTPS(http.HandlerFunc(handler), config)
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

