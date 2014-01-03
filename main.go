package main

import (
	"fmt"
	"os"

	//"encoding/json"
	//"html/template"
	"errors"
	"net/http"
	"regexp"
)

var validPath = regexp.MustCompile("^/(view)/(.+)$")
var PID string

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
	//err := godotenv.Load()
	//if err != nil {
	//log.Fatal("Error loading .env file")
	//}
	fmt.Println("listening...")
	PID = os.Getenv("PID")
	http.HandleFunc("/", root)
	http.HandleFunc("/view/", makeHandler(ViewHandler))

	//log.Fatal(http.ListenAndServe(":8080", nil))
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // The title is the second subexpression.
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Try some stores")
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//Here we will extract the page title from the Request,
		// and call the provided handler 'fn'
		m := validPath.FindStringSubmatch(r.URL.Path)

		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}
