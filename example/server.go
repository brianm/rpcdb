package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/brianm/rpcdb"
	"github.com/justinas/alice"
)

func main() {
	chain := alice.New(rpcdb.NewMiddleware).ThenFunc(handler)

	http.ListenAndServe(":3000", chain)
}

func handler(w http.ResponseWriter, req *http.Request) {
	bytes, err := httputil.DumpRequest(req, true)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "%s", err)
		return
	}
	w.Write(bytes)
}
