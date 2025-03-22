package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const _addr = ":8080"

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /dice", func(w http.ResponseWriter, r *http.Request) {
		dice := rand.Intn(6)
		fmt.Fprintf(w, "%d", dice+1)
	})
	mux.HandleFunc("POST /sleep", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	})
	log.Println("running on", _addr)
	log.Fatal("server stopped:", http.ListenAndServe(_addr, mux).Error())
}
