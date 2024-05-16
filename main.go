package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func handleHome(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fmt.Println(r.Form)
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)

	for k, v := range r.Form {
		fmt.Println("key: ", k)
		fmt.Println("value: ", strings.Join(v, ""))
	}

	fmt.Fprintf(w, "Hello go web")

}

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Method: ", r.Method)

	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.gtpl")

		log.Println(t.Execute(w, nil))
	} else {
		fmt.Println("username: ", r.FormValue("username"))
		fmt.Println("password: ", r.FormValue("password"))
	}
}

func main() {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login", login)

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatalln("Failed listen 8080 port", err)
	}
}
