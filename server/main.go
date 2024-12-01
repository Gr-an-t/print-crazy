package main

import (
	"net/http"

	"github.com/phin1x/go-ipp"
)

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello!"))
	})
	http.ListenAndServe(":8676", nil)

	client := ipp.NewIPPClient("printserver", 631, "user", "password", true)
	// print file
	client.PrintFile("/path/to/file", "my-printer", map[string]interface{}{})
}
