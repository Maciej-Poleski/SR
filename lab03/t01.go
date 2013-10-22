package main

import (
    "fmt"
    "net/http"
)

type String string

type Struct struct {
    Greeting string
    Punct    string
    Who      string
}

func (s String) ServeHTTP(r http.ResponseWriter,
                          q *http.Request){
	fmt.Fprintf(r,string(s))
}

func (s *Struct) ServeHTTP(r http.ResponseWriter,
                           q *http.Request){
	fmt.Fprintf(r,s.Greeting)
    fmt.Fprintf(r,s.Punct)
    fmt.Fprintf(r,s.Who)
    
}

func main() {
    // your http.Handle calls here
    http.Handle("/string", String("I'm a frayed knot."))
	http.Handle("/struct", &Struct{"Hello", ":", "Gophers!"})
    http.ListenAndServe("localhost:4000", nil)
}
