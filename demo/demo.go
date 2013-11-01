package main

import (
	"encoding/json"
	"flag"
	"github.com/dgryski/go-wurfl"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var wurfldb *wurfl.Wurfl

func lookupHandler(w http.ResponseWriter, r *http.Request) {

	ua := strings.TrimPrefix(r.RequestURI, "/lookup/")

	ua, err := url.QueryUnescape(ua)
	if err != nil {
		http.Error(w, "bad query", http.StatusBadRequest)
	}

	m := wurfldb.Lookup(ua)

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.Encode(m)
}

func main() {

	port := flag.String("p", ":8080", "port to listen on")
	wxml := flag.String("wxml", "wurfl.xml", "path to wurfl.xml")
	flag.Parse()

	var err error
	wurfldb, err = wurfl.New(*wxml)

	if err != nil {
		log.Fatalln("error loading wurfl:", err)
	}

	http.HandleFunc("/lookup/", lookupHandler)

	log.Println("listening on port", *port)
	log.Fatal(http.ListenAndServe(*port, nil))
}
