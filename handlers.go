package main

import (
	"fmt"
	"html/template"
	"github.com/drewwells/utils"
	"net/http"
	"encoding/json"
	//"regexp"
)

var templates = template.Must(template.ParseGlob("tmpl/*.tmpl"))

func TitleExpand(args ...interface{}) string {
	ok := false
	var s string
	if len(args) == 1 {
		s, ok = args[0].(string)
	}
	if !ok {
		s = fmt.Sprint(args...)
	}

	// find the @ symbol
	/*substrs := strings.Split(s, "@")
	if len(substrs) != 2 {
		return s
	}
	// replace the @ by " at "
	return (substrs[0] + " at " + substrs[1])*/
	return "Title: " + s
}

func ViewHandler(w http.ResponseWriter, r *http.Request, title string) {
	uri := "https://api.retailmenot.com/v1/mobile/stores/" + 
		title + "/offers"
	channel := utils.Get(uri, PID)
	req := <-channel
	
	coupons := []*Coupon{}

	err := json.Unmarshal(req.ByteStr, &coupons)
	utils.CheckError(err)

	p := &Page{Title: title, Coupons: coupons}

	renderTemplate(w, "master", p)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {

	
	funcMap := template.FuncMap{"titleExpand": TitleExpand}
	templates = templates.Funcs(funcMap)

	w.Header().Set("Content-Type", "text/html")

	err := templates.ExecuteTemplate(w, tmpl+".tmpl", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


