package main

import (
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/kdar/factorlog"
	"github.com/kdar/factorlog-contrib/gelf"
)

func main() {
	log := factorlog.New(os.Stdout, gelf.NewGELFFormatter())
	log.Print("GELF formatter")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print("GELF: client connected", gelf.Extra{"_hello": "there"}, r)
	}))

	http.Get(server.URL)
}
