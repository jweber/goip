package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
)

type Page struct {
	RemoteAddress string
	Whois         string
}

type JsonResponse map[string]interface{}

func (r JsonResponse) String() (s string) {
	b, err := json.Marshal(r)
	if err != nil {
		s = ""
		return
	}
	s = string(b)
	return
}

func negotiatingHandler(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")
	if accept == "application/json" {
		jsonHandler(w, r)
	} else {
		htmlHandler(w, r)
	}
}

func htmlHandler(w http.ResponseWriter, r *http.Request) {
	addr := getIpAddress(r)
	whoisBody := ""

	p := &Page{RemoteAddress: addr, Whois: whoisBody}

	t, _ := template.ParseFiles("index.html")
	t.Execute(w, p)
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	addr := getIpAddress(r)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, JsonResponse{"ip": addr})
	return
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

func main() {
	http.HandleFunc("/static/", staticHandler)
	http.HandleFunc("/api", jsonHandler)
	http.HandleFunc("/", negotiatingHandler)
	http.ListenAndServe(":8080", nil)
}

func getIpAddress(r *http.Request) string {
	addr := r.Header.Get("X-Real-IP")
	if addr != "" {
		return addr
	}

	addr = r.Header.Get("X-Forwarded-For")
	if addr != "" {
		return addr
	}

	addr = r.RemoteAddr
	idx := strings.LastIndex(addr, ":")
	if idx == -1 {
		return addr
	}

	return addr[:idx]
}

func whois(ip string) string {
	whoisUrl := strings.Join([]string{"http://whois.arin.net/rest/ip/", ip, ".txt"}, "")

	whoisBody := ""
	resp, err := http.Get(whoisUrl)
	if err != nil {
		whoisBody = "error"
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		whoisBody = string(body)
	}

	return whoisBody
}
