package main

import (
	"io/ioutil"
	"log"
	"testing"
)

var htm string

type dataCount struct {
	fpath     string
	count     int
	pagecount int
}

var htmCountTests = []dataCount{
	{"baddeed.shine.html.test", 179, 18},
	{"d0cf11e0.shine.html.test", 5501087, 550109},
	{"gif89.shine.html.test", 133576903, 13357691},
}

var jsonCountTests = []dataCount{
	{"baadeed.warclight.json.test", 0, 0},
	{"d0cf11e0.warclight.json.test", 10, 1},
	{"gif89.warclight.json.test", 379, 38},
}

func getData(fname string) string {
	rawhtm, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}
	return string(rawhtm)
}

func TestParseHtmForLinks(t *testing.T) {
	for _, htmTest := range htmCountTests {
		htm := getData(htmTest.fpath)
		h, err := parseHtmForLinks(htm)
		if err != nil {
			t.Error("didn't work, error returned")
		}
		if len(h) != 10 {
			t.Error("didn't work")
		}
	}
}

func TestParseHtmForResults(t *testing.T) {
	for _, htmTest := range htmCountTests {
		htm := getData(htmTest.fpath)
		count, pagecount, _ := statResults(htm)
		if count != htmTest.count {
			t.Errorf("didn't find correct count in %s", htmTest.fpath)
		}
		if pagecount != htmTest.pagecount {
			t.Errorf("didn't find correct pagecount in %s", htmTest.fpath)
		}
	}
}

func TestParseWarclight(t *testing.T) {
	for _, jsonTest := range jsonCountTests {
		js := getData(jsonTest.fpath)
		res, err := parseWarclight(js)
		if err != nil {
			t.Errorf("Unexpected error in parsing JSON in %s", jsonTest.fpath)
		}
		if res.Meta.Pages.TotalCount != jsonTest.count {
			t.Errorf("didn't find correct res count in %s", jsonTest.fpath)
		}
	}
}
