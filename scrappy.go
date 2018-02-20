package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type LastTradedPrice struct {
	scrip string
	ltd   time.Time
	ltp   float64
}

type LastTradedPrices []LastTradedPrice

var link string = "https://finance.google.com/finance?q=NASDAQ:"

var (
	scrips string
)

func main() {
	var ltps LastTradedPrices

	flag.StringVar(&scrips, "scrips", "", "Enter scrips comma separated. (Example: $scrappy -scrips=\"AAPL,GOOG\")")
	flag.Parse()

	if scrips == "" {
		flag.PrintDefaults()
	} else {

		s := strings.Split(scrips, ",")

		for _, sc := range s {
			t := LastTradedPrice{scrip: sc}
			ltps = append(ltps, t)
		}

		for i, scrip := range ltps {
			var scrapLink string = link + scrip.scrip
			fmt.Printf("Fetching LTP for %s ...", scrip.scrip)
			s := scrape(scrapLink)
			if s != "" {
				ltp, ltd := extractValues(s)
				const shortForm = "2006-01-02"
				t, _ := time.Parse(shortForm, ltd)
				ltps[i].ltp = ltp
				ltps[i].ltd = t
				fmt.Printf("done\n")
			}
		}

		for _, s := range ltps {
			fmt.Printf("%-10s\t%s\t%f\n", s.scrip, s.ltd.Local(), s.ltp)
		}
	}
}

func scrape(link string) string {

	//create the client, this will allow us to control http headers, redirect policy and other settings
	client := &http.Client{
		CheckRedirect: nil,
	}

	// we now create the request
	// since we are going to simply grab the html from the page,
	// we use the "GET"  method and pass the url of the page and nil as the body.
	req, err := http.NewRequest("GET", link, nil)

	// we now add the necessary headers
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/35.0.1916.114 Safari/537.36")

	// we now send the HTTP request. This will return the response in the resp variable
	// and if any errors are encountered they will be stored in the err variable
	resp, err := client.Do(req)
	if err != nil { // in case of any error we log it and return blank from the function
		log.Fatal(err)
		return ""
	}
	defer resp.Body.Close()                // we use defer here to close the response body when the function returns
	body, err := ioutil.ReadAll(resp.Body) // grab the entire body of the response
	s := string(body)                      // convert the body to string
	return s
}

func extractValues(html string) (float64, string) {

	// regular expression to grab the price
	// we will use a named capture group "ltp"
	var re = regexp.MustCompile(`(<meta itemprop="price"\s*content=")(?P<ltp>(\d*,*\d*\.\d*))("\s*\/>)`)

	match := re.FindStringSubmatch(html)
	// creating a map to hold capture group names and their values
	subMatchMap := make(map[string]string)
	// populating the map
	for i, name := range re.SubexpNames() {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}

	// assigning the ltp value to the variable ltp
	// also removing any comma before we convert the value to float64.
	ltp, _ := strconv.ParseFloat(strings.Replace(subMatchMap["ltp"], ",", "", -1), 64)

	// regular expression to grab the last traded date
	// we will use a named capture group "ltd"
	re = regexp.MustCompile(`(<meta itemprop="quoteTime"\s*content=")(?P<ltd>(\S+)T(\S+))(" \/>)`)

	match = re.FindStringSubmatch(html)
	subMatchMap = make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}

	ltd := string(subMatchMap["ltd"])
	ltd = ltd[:10] // taking the first 10 chars to discard the time part

	return ltp, ltd
}
