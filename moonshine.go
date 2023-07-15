package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	SR "github.com/httpreserve/simplerequest"
)

// Example SHINE uri: https://www.webarchive.org.uk/shine/search?page=2&query=content_ffb:%220baddeed%22&sort=crawl_date&order=asc

type ShineRequest struct {
	shineurl string // https://www.webarchive.org.uk/shine/search?
	page     string // page=2
	baddeed  string // &query=content_ffb:"0baddeed"
	sort     string // sort=crawl_date, title, score, comain, content
	order    string // order=asc
}

const agent string = "moonshine/1.0.0"

// Search result limits to be kind to Shine.
const solrMaxPages int = 1000
const resultsPerPage int = 10

// Consistently limit requests via this app, e.g. download a single page
// or five pages. This also helps us play nice with the service.
const singlePage int = 1
const multiPage int = 5

// FFB for the GIF file format.
const ffbGIF string = "47494638"

// FFB for the first PPT format, which inspired this code.
const ffbBadDeed string = "0baddeed"

var (
	vers   bool
	ffb    string
	gif    bool
	random bool
	page   int
	list   bool
	stat   bool
)

func init() {
	flag.StringVar(&ffb, "ffb", ffbBadDeed, "first four bytes of file to find")
	flag.BoolVar(&gif, "gif", false, "return a single gif")
	flag.BoolVar(&list, "list", false, "list the first five pages from page number")
	flag.IntVar(&page, "page", 1, "specify a page number to return from, [max: 9000]")
	flag.BoolVar(&random, "random", true, "return a random link to a file")
	flag.BoolVar(&stat, "stats", false, "return statistics for the resource")
	flag.BoolVar(&vers, "version", false, "Return version")
}

func minInt(int1 int, int2 int) int {
	if int1 < int2 {
		return int1
	}
	return int2
}

func newSearchString(newshine ShineRequest) string {
	searchstring := fmt.Sprintf("%s?page=%s&%s&%s&%s", newshine.shineurl,
		newshine.page,
		newshine.baddeed,
		newshine.sort,
		newshine.order)
	log.Printf("Created URL: %s\n", searchstring)
	return searchstring
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
	return statShineResults(resp)
}

func ping(baddeedurl string) (string, int, int) {
	log.Printf("Pinging URL: %s", baddeedurl)
	req := newRequest(baddeedurl)
	resp, _ := req.Do()

	if resp.StatusCode != 200 {
		log.Fatalf("Unsuccessful request: %s", resp.StatusText)
	}

	// Stat the results at all times to understand what other processing
	// is needed.
	filecount, pagecount, err := statResults(resp.Data)
	if err != nil {
		log.Println(err)
	}

	// Shine doesn't used a zero-based index.
	if filecount > 0 && pagecount == 0 {
		pagecount = 1
	}

	return resp.Data, filecount, pagecount
}

func concatenateresults(linkslice []string, page string) ([]string, error) {
	var res []string
	var err error
	res, err = parseHtmForLinks(page)
	if err != nil {
		return linkslice, err
	}
	linkslice = append(linkslice, res...)
	return linkslice, nil
}

func getSinglePage(linkslice []string, pageNumber int, baddeedurl ShineRequest) []string {

	var err error
	baddeedurl.page = strconv.Itoa(pageNumber)
	searchstring := newSearchString(baddeedurl)
	sr := newRequest(searchstring)
	resp, _ := sr.Do()

	if resp.StatusCode != 200 {
		log.Fatalf("Unsuccessful request: %s, exiting", resp.StatusText)
	}

	linkslice, err = concatenateresults(linkslice, resp.Data)
	if err != nil {
		log.Fatalf("%s", err)
	}
	return linkslice
}

// listResults returns a slice of all results for a given page.
func listResults(baddeedurl ShineRequest, pagecontent string, pageNumber int,
	numberOfPages int) []string {

	var linkslice []string
	var err error

	if numberOfPages == 1 && pageNumber == 1 {
		log.Println("First result already in memory")
		linkslice, err = concatenateresults(linkslice, pagecontent)
		if err != nil {
			log.Fatalf("%s", err)
		}
		return linkslice
	}

	if numberOfPages == 1 {
		return getSinglePage(linkslice, pageNumber, baddeedurl)
	}

	for pages := 0; pages < numberOfPages; pages++ {
		if pageNumber+pages == 1 {
			log.Println("First result already in memory")
			linkslice, err = concatenateresults(linkslice, pagecontent)
			continue
		}
		linkslice = getSinglePage(linkslice, pageNumber+pages, baddeedurl)
		time.Sleep(500 * time.Millisecond)
	}

	return linkslice
}

func validateHex(magic string) error {

	/*hex errors to return*/
	const NOTHEX string = "Contains invalid hexadecimal characters."
	const UNEVEN string = "Contains uneven character filecount."
	const LENGTH string = "FFB must be four bytes."

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

func returnRandomFile(pagecount int, baddeedurl ShineRequest, pagecontent string) {
	// Shine doesn't use a zero-based index.
	randomPageNumber := getRandom(pagecount) + 1
	linkslice := listResults(baddeedurl, pagecontent, randomPageNumber, singlePage)
	if len(linkslice) == 0 {
		log.Fatalf("Returned zero attempting to get random result. Exiting.")
	}
	randomFileNumber := getRandom(len(linkslice))
	// Out slice uses a zero-based index so we don't need to increment.
	log.Printf("Returning file: %d from page: %d", randomFileNumber+1, randomPageNumber)
	fmt.Println(linkslice[randomFileNumber])
}

// getFile is the primary runner of this app,
func getFile() {
	// Override the ffb and enter GIF mode...
	if gif == true {
		log.Println("Searching in GIF mode")
		ffb = ffbGIF
	}

	// Convert the ffb to lowercase for Shine then validate
	ffb = strings.ToLower(ffb)
	err := validateHex(ffb)
	if err != nil {
		log.Fatal("Invalid hexadecimal string: ", err)
	}

	// Ping the first page of the shine service to configure the search.
	var baddeedurl ShineRequest
	var pageContent string
	var filecount, pagecount int
	log.Println("Searching Shine@UKWA")
	baddeedurl = newShineSearch(1, ffb, "crawl_date", "asc")
	pageContent, filecount, pagecount = ping(newSearchString(baddeedurl))
	log.Printf("%d files discovered\n", filecount)
	log.Printf("%d pages available\n", pagecount)

	// if this, our work is done...
	if stat || filecount == 0 {
		// No files to return.
		return
	}

	if filecount > 0 && pagecount == 0 {
		// Non-zero based indexing.
		pagecount = singlePage
	}

	// Shine's SOLR has a issue deep paging beyond 10,000 results. This eats
	// RAM and CPU. To be kind to Shine we will keep the limits lower than that.
	if pagecount >= solrMaxPages {
		log.Printf("Setting pagecount ('%d') max to: %d (solrMaxPages)", pagecount, solrMaxPages)
		pagecount = solrMaxPages
		filecount = solrMaxPages * resultsPerPage
	}

	if random && !list {
		if page > 0 {
			log.Printf("Argument `-page %d` has no effect when random (default) is selected", page)
		}
		// Return a random file and then exit.
		returnRandomFile(pagecount, baddeedurl, pageContent)
		return
	}

	// Else, list five pages of files from a given offset.
	listSize := minInt((pagecount-page), multiPage) + 1
	if page > pagecount {
		log.Printf("Page number: '%d' too high, setting to max: '%d' (list size: %d)", page, pagecount, listSize)
		page = pagecount
		listSize = 1
	}

	if page == 0 {
		log.Println("Page can't be zero, setting to 1")
		page = 1
	}

	linkslice := listResults(baddeedurl, pageContent, page, listSize)

	log.Printf("Returning %d results\n", len(linkslice))
	for _, value := range linkslice {
		fmt.Println(value)
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
