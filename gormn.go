package main

import (
	"github.com/drewwells/utils"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"encoding/json"
	//"time"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"errors"
	"log"
)

type Page struct {
	Title   string
	Body    []byte
	SBody   template.HTML
	Coupons []*Coupon
}

type HttpResponse struct {
	url      string
	body     []byte
	response *http.Response
	err      error
}

type Coupon struct {
	OfferId     int32
	Title       string
	Description string
	ExpiresDate int64 //Golang only reads RFC3339 formated time
	CouponCode  string
	SuccessRate int8
	OutClickUrl string
	NewCoupon   bool
}

var validPath = regexp.MustCompile("^/(view)/(.+)$")
var templates = template.Must(template.ParseGlob("tmpl/*.tmpl"))
var PID string

func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}


func main() {
	err := godotenv.Load()
	if err != nil {
		//log.Fatal("Error loading .env file")
	}
	PID = os.Getenv("PID")
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	search := regexp.MustCompile("\\[" + p.Title + "\\]")
	p.SBody = template.HTML(string(search.ReplaceAllFunc(p.Body, 
		func(s []byte) []byte {
			return []byte("<a href=\"/view/" + p.Title + "\">" + p.Title + "</a>")
		})))

	err := templates.ExecuteTemplate(w, tmpl+".tmpl", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // The title is the second subexpression.
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	uri := "https://api.retailmenot.com/v1/mobile/stores/" + 
		title + "/offers"
	channel := utils.Get(uri, PID)
	req := <-channel
	
	coupons := []*Coupon{}

	err := json.Unmarshal(req.ByteStr, &coupons)
	if err != nil {
		fmt.Println(err)
	}
	p := &Page{Title: title, Coupons: coupons}

	renderTemplate(w, "view", p)

}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
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
