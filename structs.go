package main

import (
	"html/template"
	"net/http"
)

type Page struct {
	Title   string
	Body    []byte
	SBody   template.HTML
	Coupons []*Coupon
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
