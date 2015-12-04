package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/brianm/rpcdb"
	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
)

func main() {
	chain := alice.New(rpcdb.NewMiddleware, logger).ThenFunc(handler)

	err := http.ListenAndServe("127.0.0.1:3000", chain)
	if err != nil {
		log.Panicf("unable to start: %s", err)
	}
}

func logger(next http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, next)
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
