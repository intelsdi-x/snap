package server

import (
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"net/http"
	"fmt"
)

type Server struct {
	Port int
}

func (s *Server) Start() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)

	n := negroni.New()
	n.UseHandler(r)
	http.ListenAndServe(":" + fmt.Sprintf("%d", s.Port), n)
}

func HomeHandler(h http.ResponseWriter, r *http.Request) {
	fmt.Println("HOME")
}
