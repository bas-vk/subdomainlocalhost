package main

import (
	"fmt"
	"net/http"
)

func main() {
	store, err := NewDAppStore("/home/bas/tmp")
	mux, err := NewDAppMux(store)

	if err != nil {
		fmt.Printf("err: %v, %dapps = %d\n", err, len(store.dapps))
	} else {
		fmt.Printf("loaded %d dapps\n", len(store.dapps))
	}

	http.ListenAndServe(":4545", mux)
}
