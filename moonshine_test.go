package main

import (
	"io/ioutil"
	"log"
	"testing"
)

var htm string

type htmCount struct {
	fpath     string
	count     int
	pagecount int
}

var htmCountTests = []htmCount{
	{"baddeed.html.test", 179, 18},
	{"d0cf11e0.html.test", 5501087, 550109},
	{"gif89.html.test", 133576903, 13357691},
}

func getHtm(fname string) string {
	rawhtm, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}
	return string(rawhtm)
}

func TestParseHtmForLinks(t *testing.T) {
	for _, htmTest := range htmCountTests {
		htm := getHtm(htmTest.fpath)
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
		htm := getHtm(htmTest.fpath)
		count, pagecount, _ := statResults(htm)
		if count != htmTest.count {
			t.Errorf("didn't find correct count in %s", htmTest.fpath)
		}
		if pagecount != htmTest.pagecount {
			t.Errorf("didn't find correct pagecount in %s", htmTest.fpath)
		}
	}
}
