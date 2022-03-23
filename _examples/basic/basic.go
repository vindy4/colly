package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

type Buisness struct {
	Name      string
	YelpURL   string
	Timings   string //  all week timings
	ContactNo string
	Address   string
	Claimed   string //  verified vendor or not
	BizURL    string
}

//======= below we define selector constants based on each paage

// on buisnees page
const buisneesMainDiv = `div[data-hypernova-key="yelpfrontend__180__yelpfrontend__GondolaBizDetails__dynamic"]`

// on yelp search page

// GetStringInBetween Returns empty string if no start string found
func GetStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return
	}
	e += s + e - 1
	return str[s:e]
}

func main() {
	// Instantiate default collector
	c := colly.NewCollector(
	// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
	// colly.AllowedDomains("https://www.yelp.com/", "yelp.com/"),
	)

	fName := "yelp" + time.Now().String() + ".json"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}
	defer file.Close()

	allbiz := make([]Buisness, 0)

	//
	c.OnHTML("a[href^='mailto:']", func(e *colly.HTMLElement) {

		fmt.Println(e.Text, " yaha kya milta h in mail ")
	})

	// On every a element which has href attribute call callback
	// this is to find all the buisnees cards and then open that buisness on yelp
	c.OnHTML("span.css-1egxyvc", func(e *colly.HTMLElement) {

		var url string
		for _, v := range e.DOM.Children().Nodes {
			attrs := v.Attr
			for _, k := range attrs {
				if k.Key == "href" {
					url = "https://www.yelp.com" + k.Val
				}
			}

		}
		// pageLenght := e.ChildText(".pagination__09f24__VRjN4.border--top__09f24__exYYb.border--bottom__09f24___mg5X.border-color--default__09f24__NPAKY")
		// Only those links are visited which are in AllowedDomains
		c.Visit(e.Request.AbsoluteURL(url))

	})

	c.OnHTML(buisneesMainDiv, func(e *colly.HTMLElement) {

		bizName := e.ChildText("h1.css-12dgwvn")
		urlText := e.ChildText("div.css-xp8w2v.padding-t2__09f24__Y6duA.padding-r2__09f24__ByXi4.padding-b2__09f24__F0z5y.padding-l2__09f24__kf_t_.border--top__09f24__exYYb.border--right__09f24__X7Tln.border--bottom__09f24___mg5X.border--left__09f24__DMOkM.border-radius--regular__09f24__MLlCO.background-color--white__09f24__ulvSM > div > div > div > p  a[target=_blank]")
		contactNo := e.ChildText("div.css-xp8w2v.padding-t2__09f24__Y6duA.padding-r2__09f24__ByXi4.padding-b2__09f24__F0z5y.padding-l2__09f24__kf_t_.border--top__09f24__exYYb.border--right__09f24__X7Tln.border--bottom__09f24___mg5X.border--left__09f24__DMOkM.border-radius--regular__09f24__MLlCO.background-color--white__09f24__ulvSM > div > div > div > p.css-1p9ibgf")
		address := e.ChildText("div.css-xp8w2v.padding-t2__09f24__Y6duA.padding-r2__09f24__ByXi4.padding-b2__09f24__F0z5y.padding-l2__09f24__kf_t_.border--top__09f24__exYYb.border--right__09f24__X7Tln.border--bottom__09f24___mg5X.border--left__09f24__DMOkM.border-radius--regular__09f24__MLlCO.background-color--white__09f24__ulvSM > div > div > div > p.css-qyp8bo")
		timings := e.ChildText("table.hours-table__09f24__KR8wh.table__09f24__J2OBP.table--simple__09f24__vy16f")
		claimed := e.ChildText("span.claim-text--light__09f24__BSQOJ.css-q8l0re")

		bizurl := e.ChildAttr("div.css-xp8w2v.padding-t2__09f24__Y6duA.padding-r2__09f24__ByXi4.padding-b2__09f24__F0z5y.padding-l2__09f24__kf_t_.border--top__09f24__exYYb.border--right__09f24__X7Tln.border--bottom__09f24___mg5X.border--left__09f24__DMOkM.border-radius--regular__09f24__MLlCO.background-color--white__09f24__ulvSM > div > div > div > p  a[target=_blank]", "href")

		fmt.Println("name : ", bizName)
		fmt.Println("url text : ", urlText)
		fmt.Println("contact : ", contactNo)
		fmt.Println("address : ", address)
		fmt.Println("timings : ", timings)
		fmt.Println("claimed : ", claimed)

		parsedUrl, err := url.QueryUnescape(bizurl)
		if err != nil {
			fmt.Println("boom")
		}

		//  get the biz url
		re := regexp.MustCompile(`url=(.*)&cachebuster`)
		match := re.FindStringSubmatch(parsedUrl)
		if len(match) > 1 {
			// fmt.Println("match found -", match[1])
			bizurl = match[1]
		}
		fmt.Println("biz url  ", bizurl)

		newBiz := Buisness{
			Name:      bizName,
			ContactNo: contactNo,
			Address:   address,
			Timings:   timings,
			Claimed:   claimed,
			BizURL:    bizurl,
			YelpURL:   e.Request.URL.String(),
		}

		allbiz = append(allbiz, newBiz)

		c.Visit(e.Request.AbsoluteURL(bizurl))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping
	c.Visit("https://www.yelp.com/search?find_desc=science+camp&find_loc=Miami%2C+FL&start=10")

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")

	// Dump json to the standard output
	enc.Encode(allbiz)
}
