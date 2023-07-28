package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"github.com/google/go-github/v53/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

const (
	baseURL  string = "https://tracker.gg/valorant/profile/riot/%s/overview"
	gistName string = "🎮 Valorant Stats"
)

// Profile stores information about a valorant profile from tracker.gg
type Profile struct {
	Name              string
	KdRatio           float64
	HsPercent         float64
	Wins              int
	Losses            int
	AvgCombatScore    float64
	CurrentRank       string
	PeakRank          string
	CurrentRankPngURL string
	PeakRankPngURL    string
	EpisodeAct		  string
	PlayTime		  string
	Matches			  int64
}

func (p *Profile) setPeakAndCurrentRankInfo(e *colly.HTMLElement) {
	if strings.Contains(e.Text, "Peak") {
		splitRankAndRating := strings.Split(e.Text, "    ")[0]
		peakRank := strings.Split(splitRankAndRating, "Rating")[1]
		p.PeakRank = strings.TrimSpace(peakRank)
		p.PeakRankPngURL = e.ChildAttr("img", "src")
	} else {
		currentRank := strings.Split(e.Text, "Rating")[1]
		p.CurrentRank = strings.TrimSpace(currentRank)
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

func (p *Profile) generateMarkdown() string {
	return fmt.Sprintf(
	`%s stats | RiotID: %s
Current Rank: %s
Peak Rank %s
Wins/Losses: %d/%d | Playtime: %s | Matches: %d
Headshot Percentage: %.2f | K/D Ratio: %.2f
Average Combat Score: %.2f`,
		p.EpisodeAct,
		p.Name,
		p.CurrentRank,
		p.PeakRank,
		p.Wins,
		p.Losses,
		p.PlayTime,
		p.Matches,
		p.HsPercent,
		p.KdRatio,
		p.AvgCombatScore,
	)
}

func (p *Profile) setNumericalInfo(e *colly.HTMLElement) {
	pattern := `^(\d+(\.\d+)?)`
	reg := regexp.MustCompile(pattern)

	switch {
	case strings.Contains(e.Text, "ACS"):
		cleanupLeft := strings.TrimLeft(e.Text, "ACS")
		cleanupRight := reg.FindString(cleanupLeft)
		acsValue, err := strconv.ParseFloat(cleanupRight, 64)
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

func (p *Profile) getActOverview(e *colly.HTMLElement) {
	splitText := strings.Split(e.Text, " ")
	act := splitText[0]+splitText[1]
	playTime := splitText[4]
	matches, err := strconv.ParseInt(splitText[7], 10, 64)
	if err != nil {
		log.Println("Unable to parse matches")
	}
	p.EpisodeAct = act
	p.PlayTime = playTime
	p.Matches = matches
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var (
		githubToken = os.Getenv("GITHUB_TOKEN")
		gistID      = os.Getenv("GIST_ID")
		profileID   = os.Getenv("TRACKER_PROFILE_ID")
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
	)

	// Get ranks
	c.OnHTML("div.rating-summary__content", func(e *colly.HTMLElement) {
		profile.setPeakAndCurrentRankInfo(e)
	})

	c.OnHTML("div.details", func(e *colly.HTMLElement) {
		profile.getActOverview(e)
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

	log.Println(profile)

	// Prepare the Gist content
	content := &github.Gist{Files: map[github.GistFilename]github.GistFile{
		github.GistFilename(gistName): {Content: github.String(profile.generateMarkdown())},
	}}

	// Prepare the oauth2 config using Github Token
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	// Prepare the Github client
	client := github.NewClient(tc)

	// Create the Gist
	gist, _, err := client.Gists.Edit(ctx, gistID, content)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s successfully updated at: %s\n", gistName, *gist.HTMLURL)
}
