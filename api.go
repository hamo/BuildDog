package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var handlerMatrix = map[string]map[string]http.HandlerFunc{
	"GET": {
		"/build": getBuild,
		"/task/{id:[0-9]+}": getTaskStatus,
		"/task/{id:[0-9]+}/output": getTaskOutput,
	},
	"POST": {
		"/build": postBuild,
	},
}

func getBuild(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, buildPage)
}

func getTaskStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseUint(vars["id"], 10, 64)
	t := getTaskById(id)
	if t == nil {
		http.NotFound(w, r)
	} else {
		json, _ := json.Marshal(t)
		w.Write(json)
	}
}

func getTaskOutput(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseUint(vars["id"], 10, 64)
	t := getTaskById(id)
	if t == nil {
		http.NotFound(w, r)
	} else {
		w.Write(t.Output.Bytes())
	}
}

func postBuild(w http.ResponseWriter, r *http.Request) {
	args := make(map[string]string)
	r.ParseForm()
	if _, ok := r.Form["repo"]; ok {
		args["repo"] = r.FormValue("repo")
		args["ppa"] = r.FormValue("ppa")
		args["rev"] = r.FormValue("rev")
	} else {
		panic("not implemented")
	}
	t := newTask(args)
	id := t.enqueue()
	fmt.Fprintln(w, id)
}

func newAPIHandler() http.Handler {
	h := mux.NewRouter()

	for method, routes := range handlerMatrix {
		for route, fct := range routes {
			h.Path(route).Methods(method).HandlerFunc(fct)
		}
	}

	return h
}

func newAPIServer() *http.Server {
	return &http.Server{
		Addr:    ":8888",
		Handler: newAPIHandler(),
	}
}
