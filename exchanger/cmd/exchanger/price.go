package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type BPI struct {
	url string
	http.Client
}

func NewBPI(url string, timeout time.Duration) *BPI {
	return &BPI{
		url:    url,
		Client: http.Client{Timeout: 5 * time.Second},
	}
}

func (bpi *BPI) GetPrice() (Price, error) {
	var resp, err = bpi.Get(bpi.url)
	if err != nil {
		return Price{}, err
	}

	if resp.StatusCode != 200 {
		return Price{}, fmt.Errorf("%s", resp.Status)
	}

	defer resp.Body.Close()
	var price Price
	return price, json.NewDecoder(resp.Body).Decode(&price)
}

type Price struct {
	Time       UpdateTimestamp     `json:"time"`
	Disclaimer string              `json:"disclaimer"`
	ChartName  string              `json:"chartName"`
	Bpi        map[string]Currency `json:"bpi"`
}

type UpdateTimestamp struct {
	Updated    string    `json:"updated"`
	UpdatedISO time.Time `json:"updatedISO"`
	Updateduk  string    `json:"updateduk"`
}

type Currency struct {
	Сode        string  `json:"code"`
	Symbol      string  `json:"symbol"`
	Rate        string  `json:"rate"`
	Description string  `json:"description"`
	RateFloat   float64 `json:"rate_float"`
}
