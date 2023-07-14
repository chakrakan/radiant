package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

const (
	baseURL string = "https://tracker.gg/valorant/profile/riot/%s/overview"
	gistName string = "Valorant Stats"
)

// Profile stores information about a valorant profile from tracker.gg
type Profile struct {
	Name        string
	KdRatio     float64
	HsPercent   float64
	Wins 		int
	Losses 		int
	AvgCombatScore float64
	CurrentRank string
	PeakRank string
	CurrentRankPngURL string
	PeakRankPngURL  string
}

func (p *Profile) setPeakAndCurrentRankInfo(e *colly.HTMLElement) {
	if strings.Contains(e.Text, "Peak") {
		splitRankAndRating := strings.Split(e.Text, "    ")[0]
		p.PeakRank = strings.Split(splitRankAndRating, "Rating")[1]
		p.PeakRankPngURL = e.ChildAttr("img", "src")
	} else {
		p.CurrentRank = strings.Split(e.Text, "Rating")[1]
		p.CurrentRankPngURL = e.ChildAttr("img", "src")
	}
}

func (p *Profile) setWinsAndLosses(e *colly.HTMLElement) {
	cleanUp := strings.TrimRight(e.Text, "WL")
	trimSpace := strings.TrimSpace(cleanUp)
	splitText := strings.Split(trimSpace, "  ")
	wins, err := strconv.Atoi(splitText[0])
	if err != nil {
		log.Printf("Error converting wins: %s", err.Error())
	}
	p.Wins = wins

	losses, err := strconv.Atoi(splitText[1])
	if err != nil {
		log.Printf("Error converting losses: %s", err.Error())
	}
	p.Losses = losses
}

func (p *Profile) setNumericalInfo(e *colly.HTMLElement) {
	switch {
		case strings.Contains(e.Text, "ACS"):
			cleanupLeft := strings.TrimLeft(e.Text, "ACS")
			acsValue, err := strconv.ParseFloat(strings.Split(cleanupLeft, "Top")[0], 64)
			if err != nil {
				log.Println("Unable to parse ACS")
			}
			p.AvgCombatScore = acsValue
		case strings.Contains(e.Text, "K/D Ratio"):
			cleanupLeft := strings.TrimLeft(e.Text, "K/D Ratio")
			kdRatio, err := strconv.ParseFloat(strings.Split(cleanupLeft, "Top")[0], 64)
			if err != nil {
				log.Println("Unable to parse K/D Ratio")
			}
			p.KdRatio = kdRatio
		case strings.Contains(e.Text, "Headshot"):
			cleanupLeft := strings.TrimLeft(e.Text, "Headshot%")
			hsPercentage, err := strconv.ParseFloat(strings.Split(cleanupLeft, "%")[0], 64)
			if err != nil {
				log.Println("Unable to parse Headshot Percentage")
			}
			p.HsPercent = hsPercentage
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var (
		githubToken = os.Getenv("GITHUB_TOKEN")
		profileID = os.Getenv("TRACKER_PROFILE_ID")
	)

	if profileID == "" || githubToken == "" {
		log.Fatal("Please ensure you have the correct env variables set")
	}

	targetURL := fmt.Sprintf(baseURL, url.QueryEscape(profileID))

	// setup Profile struct
	profile := Profile{Name: profileID}

	// Instantiate default collector and scrape
	c := colly.NewCollector(
		// Visit only domains: blitz.gg, www.blitz.gg
		colly.AllowedDomains("tracker.gg", "www.tracker.gg"),
		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./radiant_cache"),
	)

	// Get ranks
	c.OnHTML("div.rating-summary__content", func(e *colly.HTMLElement) {
		profile.setPeakAndCurrentRankInfo(e)
	})

	// Get Numerical data other than wins/losses
	c.OnHTML("div.numbers", func(e *colly.HTMLElement) {
		profile.setNumericalInfo(e)
	})

	// Get Wins/Losses
	c.OnHTML("div.trn-profile-highlighted-content__ratio", func(e *colly.HTMLElement) {
		profile.setWinsAndLosses(e)
	})

	c.OnRequest(func(r *colly.Request) {
		log.Printf("Visiting %s\n", r.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error while scraping: %s\n", err.Error())
	})

	c.Visit(targetURL)

	fmt.Println(profile)

	// setup Gist
	// content := &github.Gist{Files: map[github.GistFilename]github.GistFile{
	// 	github.GistFilename(gistName): {Content: github.String(title)},
	// }}

}
