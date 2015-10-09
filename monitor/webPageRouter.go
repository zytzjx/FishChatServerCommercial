package main

import (
	"net/http"
	"text/template"
)

//index
func PageIndex(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.Error(res, "Not found", 404)
		return
	}
	if req.Method != "GET" {
		http.Error(res, "Method not allowed", 405)
		return
	}

	var homeTempl = template.Must(template.ParseFiles("static/index.html"))

	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl.Execute(res, req.Host)
}
