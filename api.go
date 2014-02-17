package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type HttpApiFunc func(w http.ResponseWriter, r *http.Request, vars map[string]string)

func makeHttpHandler(localMethod string, localRoute string, handlerFunc HttpApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check User-Agent
		if r.Header.Get("User-Agent") != "BuildDog-Client" {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		handlerFunc(w, r, mux.Vars(r))
		return

	}
}

var handlerMatrix = map[string]map[string]HttpApiFunc{
	"GET": {
		"/task/{id:[0-9]+}":        getTaskStatus,
		"/task/{id:[0-9]+}/output": getTaskOutput,
	},
	"POST": {
		"/build": postBuild,
	},
}

func getTaskStatus(w http.ResponseWriter, r *http.Request, vars map[string]string) {
	id, _ := strconv.ParseUint(vars["id"], 10, 64)
	t := getTaskById(id)
	if t == nil {
		http.NotFound(w, r)
	} else {
		json, _ := json.Marshal(t)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(json)
	}
}

func getTaskOutput(w http.ResponseWriter, r *http.Request, vars map[string]string) {
	id, _ := strconv.ParseUint(vars["id"], 10, 64)
	t := getTaskById(id)
	if t == nil {
		http.NotFound(w, r)
	} else {
		w.Write(t.Output.Bytes())
	}
}

func postBuild(w http.ResponseWriter, r *http.Request, vars map[string]string) {
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
			f := makeHttpHandler(method, route, fct)
			h.Path(route).Methods(method).HandlerFunc(f)
		}
	}

	return h
}

func newAPIServer() *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", flPort),
		Handler: newAPIHandler(),
	}
}
