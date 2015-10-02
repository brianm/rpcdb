package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/brianm/drdb"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	nlogrus "github.com/meatballhat/negroni-logrus"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		bytes, err := httputil.DumpRequest(req, true)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "%s", err)
			return
		}
		w.Write(bytes)
	})

	n := negroni.New()
	n.Use(nlogrus.NewMiddleware())
	n.Use(drdb.NewMiddleware())
	n.UseHandler(r)
	n.Run(":3030")
}

/*
https://github.com/mholt/binding
*/
