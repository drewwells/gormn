package main

import (
	"fmt"
	"html/template"
	"github.com/drewwells/utils"
	"net/http"
	"encoding/json"
	//"regexp"
)

type Page struct {
	Title   string
	Body    []byte
	SBody   template.HTML
	Coupons *[]Coupon
	Store   *Store
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

type Store struct {
	StoreId     int32
	Title       string
	Domain      string
	Description string
	MobileInStoreEnabled bool
}

var funcMap = template.FuncMap{
	"titleExpand": TitleExpand,
}

var templates = template.Must(template.New("").Funcs(funcMap).ParseGlob("tmpl/*.tmpl"))

func TitleExpand(args ...interface{}) string {
	ok := false
	var s string
	if len(args) == 1 {
		s, ok = args[0].(string)
	}
	if !ok {
		s = fmt.Sprint(args...)
	}

	return "Title: " + s
}

func ViewHandler(w http.ResponseWriter, r *http.Request, domain string) {
	coupons, store := ViewData(domain)
	renderTemplate(w, "master", &Page{
		Coupons: coupons,
		Store: store,
	})
}

func ViewData(domain string) (*[]Coupon, *Store){
	uri := "https://api.retailmenot.com/v1/mobile/stores/" + 
		domain + "/offers"

	storeURI :=  "https://api.retailmenot.com/v1/mobile/stores/" + 
		domain

	coupons := &[]Coupon{}
	channel := utils.Get(uri, PID)
	//defer close(channel)

	store := &Store{}
	storeChannel := utils.Get(storeURI, PID)
	//defer close(channel)

	//Retrieve and Unmarshal JSON
	req      := <-channel
	storeReq := <-storeChannel

	err := json.Unmarshal(req.ByteStr, &coupons)
	utils.CheckError(err)

	err = json.Unmarshal(storeReq.ByteStr, &store)
	utils.CheckError(err)

	return coupons, store
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {

	
	w.Header().Set("Content-Type", "text/html")

	err := templates.ExecuteTemplate(w, tmpl+".tmpl", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


