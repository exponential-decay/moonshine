// Warclight Pieces
//
// NB. OBSOLETE and should be removed when there is an opportunity. I
// want to query the status of this and understand more where the
// resource went.
package main

import (
	"encoding/json"
	"fmt"
)

// newWarclightSearch creates a ShineRequest object to enable us to query Shine/Warclight.
func newWarclightSearch(page int, ffb string, sort string, order string) shineRequest {
	//
	// Example warclight requests:
	//  * `http://warclight.archivesunleashed.org/catalog.json?f[content_ffb][]=47494638&page=2`
	//
	//
	var newShine shineRequest
	newShine.shineURL = "http://warclight.archivesunleashed.org/catalog.json"
	newShine.page = fmt.Sprintf("%d", page)
	newShine.badDeed = fmt.Sprintf("f[content_ffb][]=%s", ffb)
	newShine.sort = fmt.Sprintf("sort=%s", sort)
	newShine.order = fmt.Sprintf("order=%s", order)
	return newShine
}

func statWarclightResults(resp string) (int, int, error) {
	wl, err := parseWarclight(resp)
	if err != nil {
		return 0, 0, err
	}
	return wl.Meta.Pages.TotalCount,
		wl.Meta.Pages.TotalPages,
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
	CurrentPage int  `json:"Current_Page"`
	FirstPage   bool `json:"First_Page"`
	LastPage    bool `json:"Last_Page"`
	LimitValue  int  `json:"Limit_Value"`
	TotalCount  int  `json:"Total_Count"`
	TotalPages  int  `json:"Total_Pages"`
}

func parseWarclight(data string) (WarclightResult, error) {
	var js WarclightResult
	json.Unmarshal([]byte(data), &js)
	if js.Meta.Pages.CurrentPage < 1 {
		return WarclightResult{}, fmt.Errorf("Unable to read JSON result")
	}
	return js, nil
}

func parseJSONForLinks(js string) ([]string, error) {

	var httpSlice []string

	wl, err := parseWarclight(js)
	if err != nil {
		return httpSlice, err
	}

	for index := range wl.Data {
		httpSlice = append(httpSlice, wl.Data[index].Attributes.URL)
	}

	return httpSlice, nil
}
