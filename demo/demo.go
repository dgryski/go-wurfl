package main

import (
	"encoding/json"
	"flag"
	"github.com/dgryski/go-wurfl"
	"log"
	"net/http"
	"os"
	"strings"
)

var wurfldb *wurfl.Wurfl

func lookupHandler(w http.ResponseWriter, r *http.Request) {

	ua := strings.TrimPrefix("/lookup/", r.RequestURI)

	m := wurfldb.Lookup(ua)

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.Encode(m)
}

func main() {

	wxml := flag.String("wxml", "wurfl.xml", "path to wurfl.xml")
	flag.Parse()

	wurfldb = wurfl.New(*wxml)

	http.HandleFunc("/lookup/", lookupHandler)

	port := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		port = ":" + p
	}
	log.Println("listening on port", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
