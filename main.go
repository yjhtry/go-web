package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/yjhtry/go-web/session"
	_ "github.com/yjhtry/go-web/session/memory"
)

var globalSessions *session.Manager

// 然后在init函数中初始化
func init() {
	globalSessions, _ = session.NewSessionManager("memory", "goSessionId", 3600)
	go globalSessions.GC()
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	session := globalSessions.SessionStart(w, r)

	sessionValue := session.Get("username")

	fmt.Println("sessionValue: ", sessionValue)

	fmt.Fprintf(w, "Hello go web")

}

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Method: ", r.Method)

	session := globalSessions.SessionStart(w, r)

	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.gtpl")

		log.Println(t.Execute(w, nil))
	} else {
		print("session set username: ", r.Form["username"])
		session.Set("username", r.FormValue("username"))
		fmt.Println("username: ", r.FormValue("username"))
		fmt.Println("password: ", r.FormValue("password"))
		http.Redirect(w, r, "/", http.StatusFound)
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
