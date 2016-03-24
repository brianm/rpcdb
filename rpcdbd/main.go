package main

import (
	"net/http"
	"log"
	"fmt"
	"github.com/codegangsta/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "rpcdbd"
	app.Usage = "run rpc debugger server"
	app.Flags = []cli.Flag {
		cli.IntFlag{
			Name: "port, p",
			Value: 8000,
			Usage: "port to listen on",
			EnvVar: "RPCDB_PORT",
		},
	}
	app.Action = server

	app.Run(os.Args)
}

func server(c *cli.Context) {
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", c.Int("port")),
		Handler:        &DebugHandler{},
	}
	log.Fatal(s.ListenAndServe())
}

type DebugHandler struct {

}

func (d *DebugHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("hello world"))
}
