package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	SR "github.com/httpreserve/simplerequest"
)

// Example SHINE uri: https://www.webarchive.org.uk/shine/search?page=2&query=content_ffb:%220baddeed%22&sort=crawl_date&order=asc

// shineRequest holds the information needed to make a request to Shine.
type shineRequest struct {
	shineURL string // https://www.webarchive.org.uk/shine/search?
	page     string // page=2
	badDeed  string // &query=content_ffb:"0baddeed"
	sort     string // sort=crawl_date, title, score, comain, content
	order    string // order=asc
}

var (
	version = "dev-0.0.0"
	commit  = "000000000000000000000000000000000baddeed"
	date    = "1970-01-01T00:00:01Z"
)

var agent string = fmt.Sprintf("moonshine/%s", version)

// Search result limits to be kind to Shine.
const solrMaxPages int = 1000
const resultsPerPage int = 10

// Consistently limit requests via this app, e.g. download a single page
// or five pages. This also helps us play nice with the service.
const singlePage int = 1
const multiPage int = 5

// Limit the number of sampled results to return in sample mode.
const maxSample int = 20

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
	sample bool
)

func init() {
	flag.StringVar(&ffb, "ffb", ffbBadDeed, "first four bytes of file to find")
	flag.BoolVar(&gif, "gif", false, fmt.Sprintf("return a single gif (hex \"%s\")", ffbGIF))
	flag.BoolVar(&list, "list", false, "list the first five pages from page number")
	flag.BoolVar(&sample, "sample", false, "return a sampled list of up to 20 files across the maximum no. results")
	flag.IntVar(&page, "page", 1, fmt.Sprintf("specify a page number to return from, [max: %d]", solrMaxPages))
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

func newSearchString(newShine shineRequest) string {
	searchString := fmt.Sprintf("%s?page=%s&%s&%s&%s", newShine.shineURL,
		newShine.page,
		newShine.badDeed,
		newShine.sort,
		newShine.order)
	log.Printf("created URL: %s\n", searchString)
	return searchString
}

func newRequest(badDeedURL string) SR.SimpleRequest {
	sr, err := SR.Create("GET", badDeedURL)
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
			log.Fatal("exiting: UKWA server is currently experiencing technical difficultues")
		}
	}

	// if we arrive here, we've no results
	return 0, fmt.Errorf("no results string in htm")
}

func parseHtmForLinks(htm string) ([]string, error) {

	const httpIndex int = 1 // location in the split where the URL will be
	var httpSlice []string

	// Splits on newlines by default.
	scanner := bufio.NewScanner(strings.NewReader(htm))

	found := false
	for scanner.Scan() {
		if found {
			lnk := strings.Split(strings.TrimSpace(scanner.Text()), "\"")[httpIndex]
			if !strings.Contains(lnk, "http") &&
				!strings.Contains(lnk, "https") {
				return nil, fmt.Errorf("no http")
			}
			httpSlice = append(httpSlice, lnk)
			found = false
		}
		if strings.Contains(scanner.Text(), "<h4 class=\"list-group-item-heading\">") {
			found = true
		}
		if strings.Contains(scanner.Text(), "message:Service Temporarily Unavailable") {
			log.Fatal("exiting: UKWA server is currently experiencing technical difficultues")
		}
	}

	if len(httpSlice) == 0 {
		return nil, fmt.Errorf("no results")

	}

	return httpSlice, nil
}

func statResults(resp string) (int, int, error) {
	return statShineResults(resp)
}

func ping(badDeedURL string) (string, int, int) {
	log.Printf("pinging URL: %s", badDeedURL)
	req := newRequest(badDeedURL)
	resp, _ := req.Do()

	if resp.StatusCode != 200 {
		log.Fatalf("unsuccessful request: %s", resp.StatusText)
	}

	// Stat the results at all times to understand what other processing
	// is needed.
	fileCount, pageCount, err := statResults(resp.Data)
	if err != nil {
		log.Println(err)
	}

	// Shine doesn't used a zero-based index.
	if fileCount > 0 && pageCount == 0 {
		pageCount = 1
	}

	return resp.Data, fileCount, pageCount
}

func concatenateResults(linkSlice []string, page string) ([]string, error) {
	var res []string
	var err error
	res, err = parseHtmForLinks(page)
	if err != nil {
		return linkSlice, err
	}
	linkSlice = append(linkSlice, res...)
	return linkSlice, nil
}

func getSinglePage(linkSlice []string, pageNumber int, badDeedRequest shineRequest) []string {

	var err error
	badDeedRequest.page = strconv.Itoa(pageNumber)
	searchString := newSearchString(badDeedRequest)
	sr := newRequest(searchString)
	resp, _ := sr.Do()

	if resp.StatusCode != 200 {
		log.Fatalf("unsuccessful request: %s, exiting", resp.StatusText)
	}

	linkSlice, err = concatenateResults(linkSlice, resp.Data)
	if err != nil {
		log.Fatalf("%s", err)
	}
	return linkSlice
}

// listResults returns a slice of all results for a given page.
func listResults(badDeedRequest shineRequest, pageContent string, pageNumber int,
	numberOfPages int) []string {

	var linkSlice []string
	var err error

	if numberOfPages == 1 && pageNumber == 1 {
		log.Println("first result already in memory")
		linkSlice, err = concatenateResults(linkSlice, pageContent)
		if err != nil {
			log.Fatalf("%s", err)
		}
		return linkSlice
	}

	if numberOfPages == 1 {
		return getSinglePage(linkSlice, pageNumber, badDeedRequest)
	}

	for pages := 0; pages < numberOfPages; pages++ {
		if pageNumber+pages == 1 {
			log.Println("first result already in memory")
			linkSlice, _ = concatenateResults(linkSlice, pageContent)
			continue
		}
		linkSlice = getSinglePage(linkSlice, pageNumber+pages, badDeedRequest)
		time.Sleep(500 * time.Millisecond)
	}

	return linkSlice
}

func validateHex(magic string) error {

	/*hex errors to return*/
	const NOTHEX string = "contains invalid hexadecimal characters"
	const UNEVEN string = "contains uneven character filecount"
	const LENGTH string = "ffb must be four bytes"

	var regexString = "^[A-Fa-f\\d]+$"
	res, _ := regexp.MatchString(regexString, magic)
	if !res {
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

func returnRandomFile(pageCount int, badDeedRequest shineRequest, pageContent string) {
	// Shine doesn't use a zero-based index.
	randomPageNumber := getRandom(pageCount) + 1
	linkSlice := listResults(badDeedRequest, pageContent, randomPageNumber, singlePage)
	if len(linkSlice) == 0 {
		log.Fatalf("returned zero attempting to get random result. Exiting.")
	}
	randomFileNumber := getRandom(len(linkSlice))
	// Out slice uses a zero-based index so we don't need to increment.
	log.Printf("returning file: %d from page: %d", randomFileNumber+1, randomPageNumber)
	fmt.Println(linkSlice[randomFileNumber])
}

func checkExistence(needle int, slice []int) bool {
	for _, value := range slice {
		if value == needle {
			return true
		}
	}
	return false
}

// Return a bar distribution for the pages files are selected from.
func getDistribution(dist []int) {
	sort.Ints(dist)
	var distString string
	columns := 100
	for i := 0; i < columns; i++ {
		for _, v := range dist {
			x := float64(v)
			y := float64(solrMaxPages)
			percentage := float64((x / y) * 100)
			calculated := int(math.Ceil(percentage))
			if calculated == i {
				distString = fmt.Sprintf("%s|", distString)
				continue
			}
		}
		distString = fmt.Sprintf("%s‖", distString)
	}
	log.Printf("distribution (pages): %s", distString)
}

func returnSampledList(badDeedRequest shineRequest, pageCount int, fileCount int) {

	var randomFiles []int
	var returnSlice []string

	for i := 0; i < maxSample; i++ {
		possibleFile := getRandom(fileCount)
		// If performance is terrible we can rely on the map to de-dupe and return
		// less than 20.
		if checkExistence(possibleFile, randomFiles) {
			// Decrement counter and find another value.
			i--
			continue
		}
		randomFiles = append(randomFiles, possibleFile)
	}

	// Map file number to a page number, map([file]pageNumber)
	indexPageMap := make(map[int]int)
	var pageDist []int
	var keys []int
	for _, value := range randomFiles {
		page := int(value/10) + 1
		indexPageMap[value] = page
		pageDist = append(pageDist, page)
		keys = append(keys, value)
	}

	getDistribution(pageDist)
	sort.Ints(keys)
	pageCounter := 0
	var newLinkSlice []string
	for _, value := range keys {
		_page := indexPageMap[value]
		if _page != pageCounter {
			pageCounter = _page
			log.Println("get page", pageCounter)
			newLinkSlice = nil
			newLinkSlice = getSinglePage(newLinkSlice, pageCounter, badDeedRequest)
		} else {
			log.Println("page already in memory:", pageCounter)
		}
		index := value % 10
		returnSlice = append(returnSlice, newLinkSlice[index])
	}

	for _, value := range returnSlice {
		// Print our random sample out to the console.
		fmt.Println(value)
	}
}

// getFile is the primary runner of this app,
func getFile() {
	// Override the ffb and enter GIF mode...
	if gif {
		log.Println("searching in GIF mode")
		ffb = ffbGIF
	}

	// Convert the ffb to lowercase for Shine then validate
	ffb = strings.ToLower(ffb)
	err := validateHex(ffb)
	if err != nil {
		log.Fatal("invalid hexadecimal string: ", err)
	}

	// Ping the first page of the shine service to configure the search.
	var badDeedRequest shineRequest
	var pageContent string
	var fileCount, pageCount int
	log.Println("searching Shine@UKWA")
	badDeedRequest = newShineSearch(1, ffb, "crawl_date", "asc")
	pageContent, fileCount, pageCount = ping(newSearchString(badDeedRequest))
	log.Printf("files discovered: '%d'", fileCount)
	log.Printf("pages available: '%d'", pageCount)

	// if this, our work is done...
	if stat || fileCount == 0 {
		// No files to return.
		return
	}

	if fileCount > 0 && pageCount == 0 {
		// Non-zero based indexing.
		pageCount = singlePage
	}

	// Shine's SOLR has a issue deep paging beyond 10,000 results. This eats
	// RAM and CPU. To be kind to Shine we will keep the limits lower than that.
	if pageCount >= solrMaxPages {
		log.Printf("setting pageCount ('%d') max to: %d (solrMaxPages)", pageCount, solrMaxPages)
		pageCount = solrMaxPages
		fileCount = solrMaxPages * resultsPerPage
	}

	if random && !list && !sample {
		if page > 0 {
			log.Printf("argument `-page %d` has no effect when random (default) is selected", page)
		}
		// Return a random file and then exit.
		returnRandomFile(pageCount, badDeedRequest, pageContent)
		return
	}

	if sample && fileCount > maxSample {
		log.Println("returning a sampled list...")
		returnSampledList(badDeedRequest, pageCount, fileCount)
		return
	}

	// Else, list five pages of files from a given offset.
	listSize := minInt((pageCount-page), multiPage) + 1
	if page > pageCount {
		log.Printf("page number: '%d' too high, setting to max: '%d' (list size: %d)", page, pageCount, listSize)
		page = pageCount
		listSize = 1
	}

	if page == 0 {
		log.Println("page can't be zero, setting to 1")
		page = 1
	}

	linkSlice := listResults(badDeedRequest, pageContent, page, listSize)

	log.Printf("returning %d results\n", len(linkSlice))
	for _, value := range linkSlice {
		fmt.Println(value)
	}
}

func main() {
	flag.Parse()
	if vers {
		fmt.Fprintf(os.Stderr, "%s (%s) commit: %s date: %s\n", agent, version, commit, date)
		os.Exit(0)
	} else if flag.NFlag() < 0 { // can access args w/ len(os.Args[1:]) too
		fmt.Fprintln(os.Stderr, "Usage:  baddeed")
		fmt.Fprintln(os.Stderr, "        OPTIONAL: [-ffb] ... OPTIONAL: [-list] ...")
		fmt.Fprintln(os.Stderr, "        OPTIONAL: [-ffb] ... OPTIONAL: [-sample] ...")
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
