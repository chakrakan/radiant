package main

import (
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
	profileID := "adorn-1625"

	if profileID == "" {
		log.Println("Profile post id required")
		os.Exit(1)
	}

	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains: blitz.gg, www.blitz.gg
		colly.AllowedDomains("blitz.gg", "www.blitz.gg"),

		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./radiant_cache"),
	)

	detailCollector := c.Clone()

	profiles := make([]Profile, 0, 200)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		log.Println(e.Attr("href"))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	// Extract details of the profile
	detailCollector.OnHTML(`div[id=blitz-app]`, func(e *colly.HTMLElement) {
		log.Println("Profile found", e.Request.URL)
		name := e.ChildText(".profile-info")
		if name == "" {
			log.Println("No profile found", e.Request.URL)
		}
		profile := Profile{
			Name:       name,
			LastPlayed:         e.Request.URL.String(),
		}
		profiles = append(profiles, profile)
	})

	c.Visit("https://blitz.gg/valorant/profile/" + profileID)

}
