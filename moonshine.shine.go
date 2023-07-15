package main

import (
	"fmt"
)

// newShineSearch creates a ShineRequest object to enable us to query Shine/Warclight.
func newShineSearch(page int, ffb string, sort string, order string) shineRequest {
	//
	// Example Shine requests:
	// `* `https://www.webarchive.org.uk/shine/search?page=1&query=content_ffb:"47494638"&sort=crawl_date&order=asc`
	//  * `https://www.webarchive.org.uk/shine/search?page=1&query=content_ffb:"d0cf11e0"&sort=crawl_date&order=asc`
	//
	// Example warclight requests:
	//
	//  * `http://warclight.archivesunleashed.org/catalog.json?f[content_ffb][]=d0cf11e0`
	//  * `http://warclight.archivesunleashed.org/catalog.json?f[content_ffb][]=baaddeed`
	//  * `http://warclight.archivesunleashed.org/catalog.json?f[content_ffb][]=47494638`
	//
	//
	var newShine shineRequest
	newShine.shineURL = "https://www.webarchive.org.uk/shine/search"
	newShine.page = fmt.Sprintf("%d", page)
	newShine.badDeed = fmt.Sprintf("query=content_ffb:%s", ffb)
	newShine.sort = fmt.Sprintf("sort=%s", sort)
	newShine.order = fmt.Sprintf("order=%s", order)
	return newShine
}

func statShineResults(resp string) (int, int, error) {
	resultsPerPage := 10
	count, err := parseHtmForResults(resp)
	if err != nil {
		return 0, 0, err
	}

	// round up pageCount if remainder isn't zero
	r := count % 10
	pageCount := count / resultsPerPage
	if count > 10 && r > 0 {
		pageCount = pageCount + 1
	}
	return count, pageCount, nil
}
