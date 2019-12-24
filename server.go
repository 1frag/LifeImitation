package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8080", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "./static/index.html")
}

func main() {
	flag.Parse()
	fs := http.FileServer(http.Dir("static"))

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
