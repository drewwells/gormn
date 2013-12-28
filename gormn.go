package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"errors"
	"reflect"
)

type Page struct {
	Title string
	Body []byte
	SBody template.HTML
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var templates = template.Must(template.ParseGlob("tmpl/*.tmpl"))

func enumerate(x interface{}) {
	val := reflect.ValueOf(x).Elem()

	i := 0
	for {
		if(i >= val.NumField()){
			break
		}
		//valueField := val.Field(i)
		typeField := val.Type().Field(i)

		fmt.Printf("Field Name: %s,\t Field Value: ,\t \n",
			typeField.Name)
		i++
	}
}

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

var PID string

func main() {
	err := godotenv.Load()
	if err != nil {
		//log.Fatal("Error loading .env file")
	}
	PID = os.Getenv("PID")

	http.HandleFunc("/",func(w http.ResponseWriter, r *http.Request){
		http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
	})
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	http.HandleFunc("/rmn/", func(w http.ResponseWriter, r *http.Request){
		
		w.Header().Set("pid",PID)
		w.Header().Set("fp","gormn")
		resp, err := http.Get("https://api.retailmenot.com/v1/services")

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			
		}
		fmt.Println(string(body))
	})

	body := ajax("https://api.retailmenot.com/" +
		"v1/mobile/offers/mobile/featured")

	fmt.Println(string(body))
	
}

func ajax(url string) ([]byte) {

	client := &http.Client{
	}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println(err)
	}

	req.Header.Set("pid",PID)
	req.Header.Set("fp","gormn")

	resp, errr := client.Do(req)
	if errr != nil {
		fmt.Println(errr)
	}
	defer resp.Body.Close()
	bs, _ := ioutil.ReadAll(resp.Body)
	return bs
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
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
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
