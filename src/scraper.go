package main

import (
	"flag"
	"log"
	"os"

	"github.com/gocolly/colly"
)

// Profile stores information about a valorant profile from blitz.gg
type Profile struct {
	Name        string
	LastPlayed  string
	KdRatio     string
	HsPercent   string
	DmgPerRound string
	CombatScore string
	RankSvgURL  string
}

func main() {
	var profileID string
	flag.StringVar(&profileID, "Your blitz.gg profile id")
	flag.Parse()

	if itemID == "" {
		log.Println("Profile post id required")
		os.Exit(1)
	}

	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains: blitz.gg, www.blitz.gg
		colly.AllowedDomains("blitz.gg", "www.blitz.gg"),

		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./rank_cache"),
	)

	profileInfo := make([]Profile, 0, 200)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		log.Println(e.Attr("href"))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	c.Visit("https://blitz.gg/valorant/profile/" + itemID)
}
