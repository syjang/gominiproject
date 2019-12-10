package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

func resultHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {

	b := []byte("hello world")
	w.Write(b)
}

func main() {
	router := httprouter.New()

	router.GET("/", resultHandler)

	// create negroni for recovery, logger, static resource
	n := negroni.Classic()
	n.UseHandler(router)

	s := fmt.Sprintf(":%d", 8989)
	// Start Server
	log.Printf("Listen ... %d\n", 8989)
	http.ListenAndServe(s, n)
}
