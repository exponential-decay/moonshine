package main

import (
	"fmt"
)

// newShineSearch creates a ShineRequest object to enable us to query Shine/Warclight.
func newShineSearch(page int, ffb string, sort string, order string) ShineRequest {
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
	var newshine ShineRequest
	newshine.shineurl = "https://www.webarchive.org.uk/shine/search"
	newshine.page = fmt.Sprintf("%d", page)
	newshine.baddeed = fmt.Sprintf("query=content_ffb:%s", ffb)
	newshine.sort = fmt.Sprintf("sort=%s", sort)
	newshine.order = fmt.Sprintf("order=%s", order)
	return newshine
}

func statShineResults(resp string) (int, int, error) {
	resultsperpage := 10
	count, err := parseHtmForResults(resp)
	if err != nil {
		return 0, 0, err
	}

	// round up pagecount if remainder isn't zero
	r := count % 10
	pagecount := count / resultsperpage
	if count > 10 && r > 0 {
		pagecount = pagecount + 1
	}
	return count, pagecount, nil
}
