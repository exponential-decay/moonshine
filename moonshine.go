package main

import (
	"bufio"
	"flag"
	"fmt"
	SR "github.com/httpreserve/simplerequest"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// Example SHINE uri: https://www.webarchive.org.uk/shine/search?page=2&query=content_ffb:%220baddeed%22&sort=crawl_date&order=asc

type shinerequest struct {
	shineurl string // https://www.webarchive.org.uk/shine/search?
	page     string // page=2
	baddeed  string // &query=content_ffb:"0baddeed"
	sort     string // sort=crawl_date, title, score, comain, content
	order    string // order=asc
}

const agent string = "moonshine-api-wrapper/0.0.1"
const maxpages int = 5

var (
	vers   bool
	ffb    string
	random bool
	//page    int
	list bool
	//listall bool
	randno int
	stat   bool
	rgen   *rand.Rand
)

func init() {
	flag.StringVar(&ffb, "ffb", "0baddeed", "first four bytes of file to find")
	flag.BoolVar(&list, "list", false, "list up to the first five pages results")
	//flag.BoolVar(&listall, "list-all", false, "list all the results")
	//flag.IntVar(&page, "page", 1, "specify a page number to return from")
	flag.BoolVar(&random, "random", true, "return a random link to a file")
	flag.BoolVar(&stat, "stat", false, "stat the resource")
	flag.BoolVar(&vers, "version", false, "Return version.")

	seed := rand.NewSource(time.Now().UnixNano())
	rgen = rand.New(seed)
}

func minint(x int, y int) int {
	if x < y {
		return x
	}
	return y
}

func newSearchString(newshine shinerequest) string {
	return fmt.Sprintf("%s?page=%s&%s&%s&%s", newshine.shineurl,
		newshine.page,
		newshine.baddeed,
		newshine.sort,
		newshine.order)
}

func newSearch(page int, ffb string, sort string, order string) shinerequest {
	var newshine shinerequest
	newshine.shineurl = "https://www.webarchive.org.uk/shine/search"
	newshine.page = fmt.Sprintf("%d", page)
	newshine.baddeed = fmt.Sprintf("&query=content_ffb:\"%s\"", ffb)
	newshine.sort = fmt.Sprintf("sort=%s", sort)
	newshine.order = fmt.Sprintf("order=%s", order)
	return newshine
}

func newRequest(baddeedurl string) SR.SimpleRequest {
	sr, err := SR.Create("GET", baddeedurl)
	if err != nil {
		log.Fatal("create request failed: %s\n", err)
	}
	sr.Agent(agent)
	return sr
}

func parseHtmForResults(htm string) (int, error) {
	// Look for: Results <span id="displayingXOfY">1 to 10 of 179</span>

	var res string

	// Splits on newlines by default.
	scanner := bufio.NewScanner(strings.NewReader(htm))
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "<span id=\"displayingXOfY\">") {
			res = strings.TrimSpace(scanner.Text())
			res = strings.Replace(res, "Results <span id=\"displayingXOfY\">", "", 1)
			res = strings.Replace(res, "</span>", "", 1)
			res = strings.Replace(res, ",", "", -1)
			resList := strings.Split(res, "of")
			res, _ := strconv.ParseInt(strings.TrimSpace(resList[len(resList)-1]), 10, 32)
			resInt := int(res)
			return resInt, nil
		}
	}

	if err := scanner.Err(); err != nil {
		// Handle the error
	}

	// if we arrive here, we've no results
	return 0, fmt.Errorf("no results string")
}

func parseHtmForLinks(htm string) ([]string, error) {

	const httpindex int = 1 // location in the split where the URL will be
	var httpslice []string

	// Splits on newlines by default.
	scanner := bufio.NewScanner(strings.NewReader(htm))

	f := false
	for scanner.Scan() {
		if f == true {
			lnk := strings.Split(strings.TrimSpace(scanner.Text()), "\"")[httpindex]
			if !strings.Contains(lnk, "http") &&
				!strings.Contains(lnk, "https") {
				return nil, fmt.Errorf("no http")
			}
			httpslice = append(httpslice, lnk)
			f = false
		}
		if strings.Contains(scanner.Text(), "<h4 class=\"list-group-item-heading\">") {
			f = true
		}
	}

	if len(httpslice) == 0 {
		return nil, fmt.Errorf("no results")
	}

	if err := scanner.Err(); err != nil {
		// Handle the error
	}

	return httpslice, nil
}

func statResults(resp string) (int, int, error) {
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

func ping(baddeedurl string) (string, int, int) {
	req := newRequest(baddeedurl)
	resp, _ := req.Do()

	// stat the results at all times to understand what other processing
	// is needed.
	count, pagecount, err := statResults(resp.Data)
	if err != nil {
		log.Println(err)
	}

	log.Printf("%d files discovered\n", count)
	log.Printf("%d pages available\n", pagecount)
	return resp.Data, count, pagecount
}

func augmentresults(linkslice []string, page string) ([]string, error) {
	res, err := parseHtmForLinks(page)
	if err != nil {
		return linkslice, err
	}
	linkslice = append(linkslice, res...)
	return linkslice, nil
}

func listresults(baddeedurl shinerequest, page string, pagecount int) []string {

	var linkslice []string
	var err error

	for x := 1; x <= minint(maxpages, pagecount); x++ {
		// don't redo work we've already done
		// use the current page in memory
		if page != "" {
			linkslice, err = augmentresults(linkslice, page)
			if err != nil {
				log.Println(err)
			}
			page = ""
			continue
		}
		baddeedurl.page = strconv.Itoa(x)
		sr := newRequest(newSearchString(baddeedurl))
		resp, _ := sr.Do()
		linkslice, err = augmentresults(linkslice, resp.Data)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(200 * time.Millisecond)
	}
	return linkslice
}

func getFile() {

	baddeedurl := newSearch(1, "0baddeed", "crawl_date", "asc")
	if ffb != "" {
		baddeedurl = newSearch(1, ffb, "crawl_date", "asc")
	}

	log.Printf("URL: %s\n", baddeedurl)
	page, count, pagecount := ping(newSearchString(baddeedurl))

	// if this, our work is done...
	if stat || count == 0 {
		return
	}

	linkslice := listresults(baddeedurl, page, pagecount)

	if random && !list {
		randno = rgen.Intn(len(linkslice))
		fmt.Println(linkslice[randno])
		return
	}

	log.Printf("Returning %d results\n", len(linkslice))
	for _, l := range linkslice {
		fmt.Println(l)
	}

}

func main() {
	flag.Parse()
	if vers {
		fmt.Fprintf(os.Stderr, "%s \n", agent)
		os.Exit(0)
	} else if flag.NFlag() < 0 { // can access args w/ len(os.Args[1:]) too
		fmt.Fprintln(os.Stderr, "Usage:  baddeed")
		fmt.Fprintln(os.Stderr, "        OPTIONAL: [-ffb] ... OPTIONAL: [-list] ...")
		fmt.Fprintln(os.Stderr, "        OPTIONAL: [-stat]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Output: [STRING] {url}")
		fmt.Fprintf(os.Stderr, "Output: [STRING] '%s ...'\n\n", agent)
		flag.Usage()
		os.Exit(0)
	} else {
		getFile()
	}
}
