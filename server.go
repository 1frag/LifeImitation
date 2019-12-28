package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

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
	var addr string
	if os.Getenv("PORT") != "" {
		addr = os.Getenv("PORT")
	} else {
		addr = "8888"
	}
	fmt.Print(os.Getenv("PORT"))
	flag.Parse()
	fs := http.FileServer(http.Dir("static"))

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	http.HandleFunc("/_debug", ServeDebug)
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	err := http.ListenAndServe(":" + addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
