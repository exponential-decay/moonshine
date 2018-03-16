package main

import (
	"bufio"
	"flag"
	"fmt"
	SR "github.com/httpreserve/simplerequest"
	"log"
	"math/rand"
	"os"
	"regexp"
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

const agent string = "moonshine-api-wrapper/0.0.2"
const maxpages int = 5
const solrmax int = 10
const resultsPerPage int = 10
const solrlimit int = 9
const ffbgif string = "47494638"

var (
	vers   bool
	ffb    string
	gif    bool
	random bool
	page   int
	list   bool
	stat   bool
	rgen   *rand.Rand
)

func init() {
	flag.StringVar(&ffb, "ffb", "0baddeed", "first four bytes of file to find")
	flag.BoolVar(&gif, "gif", false, "return a single gif from the UKWA")
	flag.BoolVar(&list, "list", false, "list the first five pages from page number")
	flag.IntVar(&page, "page", 1, "specify a page number to return from, [max: 9000]")
	flag.BoolVar(&random, "random", true, "return a random link to a file")
	flag.BoolVar(&stat, "stat", false, "stat the resource")
	flag.BoolVar(&vers, "version", false, "Return version")

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
	newshine.baddeed = fmt.Sprintf("query=content_ffb:\"%s\"", ffb)
	newshine.sort = fmt.Sprintf("sort=%s", sort)
	newshine.order = fmt.Sprintf("order=%s", order)
	return newshine
}

func newRequest(baddeedurl string) SR.SimpleRequest {
	sr, err := SR.Create("GET", baddeedurl)
	if err != nil {
		log.Fatalf("create request failed: %s\n", err)
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

		if strings.Contains(scanner.Text(), "message:Service Temporarily Unavailable") {
			log.Fatal("Exiting: UKWA server is currently experiencing technical difficultues")
		}
	}

	if err := scanner.Err(); err != nil {
		// Handle the error
	}

	// if we arrive here, we've no results
	return 0, fmt.Errorf("no results string in htm")
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
		if strings.Contains(scanner.Text(), "message:Service Temporarily Unavailable") {
			log.Fatal("Exiting: UKWA server is currently experiencing technical difficultues")
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
	log.Printf("URL: %s", baddeedurl)
	req := newRequest(baddeedurl)
	resp, _ := req.Do()

	if resp.StatusCode != 200 {
		log.Fatalf("Unsuccessful request: %s", resp.StatusText, 1)
	}

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

func concatenateresults(linkslice []string, page string) ([]string, error) {
	res, err := parseHtmForLinks(page)
	if err != nil {
		return linkslice, err
	}
	linkslice = append(linkslice, res...)
	return linkslice, nil
}

func listresults(baddeedurl shinerequest,
	pagecontent string,
	pageno int,
	pagecount int,
	numberOfPages int) []string {

	var linkslice []string
	var err error

	if numberOfPages == 0 || numberOfPages > maxpages {
		numberOfPages = maxpages
	}

	for x := pageno; x < pageno+minint(numberOfPages, pagecount); x++ {
		// don't redo work we've already done in PING if we're just
		// looking for the first page.
		if pagecontent != "" && pageno == 1 {
			log.Println("First result already in memory")
			linkslice, err = concatenateresults(linkslice, pagecontent)
			if err != nil {
				log.Println(err)
			}
			pagecontent = ""
			continue
		}

		baddeedurl.page = strconv.Itoa(x)
		log.Println(newSearchString(baddeedurl))
		sr := newRequest(newSearchString(baddeedurl))
		resp, _ := sr.Do()

		if resp.StatusCode != 200 {
			log.Printf("Unsuccessful request: %s\n", resp.StatusText)
			if len(linkslice) > 0 {
				log.Println("Dumping results received so far:")
				for _ , x := range linkslice {
					fmt.Println(x)
				}
			}
			log.Fatal("exiting")
		}

		linkslice, err = concatenateresults(linkslice, resp.Data)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(500 * time.Millisecond)
	}
	return linkslice
}

func validateHex(magic string) error {

	/*hex errors to return*/
	const NOTHEX string = "Contains invalid hexadecimal characters."
	const UNEVEN string = "Contains uneven character count."
	const LENGTH string = "FFB in UKWA must be four bytes."

	var regexString = "^[A-Fa-f\\d]+$"
	res, _ := regexp.MatchString(regexString, magic)
	if res == false {
		return fmt.Errorf(NOTHEX)
	}
	if len(magic)%2 != 0 {
		return fmt.Errorf(UNEVEN)
	}
	if len(magic) < 8 || len(magic) > 8 {
		return fmt.Errorf(LENGTH)
	}
	return nil
}

func getRandom(count int, pagecount int) (int, int) {

	randno := rgen.Intn(count)
	pageno := (randno / resultsPerPage) + 1
	if pageno > pagecount {
		log.Println("finding the page number for the random file failed, resetting")
		getRandom(count, pagecount)
	}
	log.Printf("returning result %d", randno)
	if randno%resultsPerPage == 0 {
		return resultsPerPage, pageno
	}
	return randno % resultsPerPage, pageno
}

func getFile() {

	// Override ffb and enter GIF mode...
	if gif == true {
		log.Println("Searching UKWA in GIF mode")
		ffb = ffbgif
	}

	err := validateHex(ffb)
	if err != nil {
		log.Fatal("Invalid hexadecimal string: ", err)
	}

	// Ping the first page to configure our search...
	baddeedurl := newSearch(1, ffb, "crawl_date", "asc")
	pagecontent, count, pagecount := ping(newSearchString(baddeedurl))

	// if this, our work is done...
	if stat || count == 0 {
		return
	}

	// SOLR has a issue deep paging beyond 10,000 results. This eats RAM and
	// CPU. Be kind to the UKWA Shine Project and don't make the pagecouny
	// any higher than that.
	if pagecount >= solrmax {
		pagecount = solrlimit
		count = solrlimit * resultsPerPage
	}

	if random && !list {
		randno, pageno := getRandom(count, pagecount)
		linkslice := listresults(baddeedurl, pagecontent, pageno, pagecount, 1)
		fmt.Println(linkslice[randno])
		return
	}

	// List results from specific page no.
	numberOfPages := maxpages
	if pagecount < page {
		page = pagecount
		numberOfPages = 1
		log.Println("page number too high, returning last page of results")
	}

	linkslice := listresults(baddeedurl, pagecontent, page, pagecount, numberOfPages)

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
		fmt.Fprintln(os.Stderr, "        OPTIONAL: [-gif] ...")
		fmt.Fprintln(os.Stderr, "        OPTIONAL: [-stat]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Output: [STRING] {url}")
		fmt.Fprintln(os.Stderr, "Output: [LIST]   {url}")
		fmt.Fprintln(os.Stderr, "                 {url}")
		fmt.Fprintln(os.Stderr, "                  ... ")
		fmt.Fprintln(os.Stderr, "                  ... ")
		fmt.Fprintf(os.Stderr, "Output: [STRING] '%s ...'\n\n", agent)
		flag.Usage()
		os.Exit(0)
	} else {
		getFile()
	}
}
