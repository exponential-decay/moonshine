// Warclight Pieces
package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func newWarclightSearch(page int, ffb string, sort string, order string) shinerequest {
	// Example warclight requests:
	// http://warclight.archivesunleashed.org/catalog.json?f[content_ffb][]=47494638&page=2
	var newshine shinerequest
	newshine.shineurl = "http://warclight.archivesunleashed.org/catalog.json"
	newshine.page = fmt.Sprintf("%d", page)
	newshine.baddeed = fmt.Sprintf("f[content_ffb][]=%s", ffb)
	newshine.sort = fmt.Sprintf("sort=%s", sort)
	newshine.order = fmt.Sprintf("order=%s", order)
	return newshine
}

func statWarclightResults(resp string) (int, int, error) {
	wl, err := parseWarclight(resp)
	if err != nil {
		return 0, 0, err
	}
	return wl.Meta.Pages.Total_Count,
		wl.Meta.Pages.Total_Pages,
		nil
}

// WarclightResult holds the Warclight return data
type WarclightResult struct {
	Data []WarclightAttrs
	Meta WarclightMeta
}

// WarclightAttrs data contains information about the web pages returned
type WarclightAttrs struct {
	Attributes WarclightAttrDetails
}

// WarclightAttrDetails stores web page attributes including the URL we want
type WarclightAttrDetails struct {
	URL string
}

// WarclightMeta contains metadata about the result returned
type WarclightMeta struct {
	Pages WarclightPages
}

// WarclightPages is our result metadata
type WarclightPages struct {
	Current_Page int
	First_Page   bool
	Last_Page    bool
	Limit_Value  int
	Total_Count  int
	Total_Pages  int
}

func parseWarclight(data string) (WarclightResult, error) {
	log.Println("Received JSON data for Warclight")
	var js WarclightResult
	json.Unmarshal([]byte(data), &js)
	if js.Meta.Pages.Current_Page < 1 {
		return WarclightResult{}, fmt.Errorf("Unable to read JSON result")
	}
	return js, nil
}

func parseJSONForLinks(js string) ([]string, error) {

	var httpslice []string

	wl, err := parseWarclight(js)
	if err != nil {
		return httpslice, err
	}

	for index := range wl.Data {
		httpslice = append(httpslice, wl.Data[index].Attributes.URL)
	}

	return httpslice, nil
}
