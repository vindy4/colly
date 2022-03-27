package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
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
	BizEmail  string
}

//======= below we define selector constants based on each paage

// on buisnees page
const buisneesMainDiv = `div[data-hypernova-key="yelpfrontend__180__yelpfrontend__GondolaBizDetails__dynamic"]`

// on yelp search page

func main() {
	// Instantiate default collector
	c := colly.NewCollector(
	// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
	// colly.AllowedDomains("https://www.yelp.com/", "yelp.com/"),
	)

	allBiz := make([]Buisness, 0)

	fName := "yelp" + time.Now().String() + ".csv"

	file, err := os.Create(fName)
	if err != nil {
		log.Fatalln("failed to open file", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	//  try to find email on buisness url
	c.OnHTML("a[href^='mailto:']", func(e *colly.HTMLElement) {
		fmt.Println(e.Text, " yaha kya milta h in mail ", e.Request.URL.String())

		for _, b := range allBiz {
			if b.BizURL == e.Request.URL.String() {
				b.BizEmail = e.Text
			}
		}

	})

	// On every a element which has href attribute call callback
	// this is to find all the buisnees cards and then open that buisness on yelp
	c.OnHTML("span.css-1egxyvc", func(e *colly.HTMLElement) {

		var buisnessUrl string
		for _, v := range e.DOM.Children().Nodes {
			attrs := v.Attr
			for _, k := range attrs {
				if k.Key == "href" {
					buisnessUrl = "https://www.yelp.com" + k.Val
				}
			}

		}
		c.Visit(e.Request.AbsoluteURL(buisnessUrl))

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

		allBiz = append(allBiz, Buisness{
			Name:      bizName,
			ContactNo: contactNo,
			Address:   address,
			Timings:   timings,
			Claimed:   claimed,
			BizURL:    bizurl,
			YelpURL:   e.Request.URL.String(),
		})

		c.Visit(e.Request.AbsoluteURL(bizurl))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping
	c.Visit("https://www.yelp.com/search?find_desc=Math+Tutoring+Center&find_loc=Miami%2C+FL&start=10")

	// add heading
	headerRow := []string{"Name",
		"YelpURL",
		"Timings ",
		"ContactNo",
		"Address",
		"Claimed",
		"BizURL"}

	err = w.Write(headerRow)
	if err != nil {
		fmt.Println(" error in writing header")
	}

	// Using Write
	for _, b := range allBiz {
		row := []string{b.Name, b.YelpURL, b.Timings, b.Claimed, b.Address, b.Claimed, b.BizURL}
		if err := w.Write(row); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}

}
