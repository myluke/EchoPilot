package main

import (
	"io/ioutil"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/gommon/log"
	"github.com/mylukin/EchoPilot/helper"
	"github.com/mylukin/EchoPilot/service/emoji"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var reqClient *req.Client

type lookup struct {
	Name string
	URL  string
}

func init() {
	reqClient = req.C()
	reqClient.SetTimeout(5 * time.Second)
}

func main() {
	log.Info("Updating Emoji Definition using Emojipedia…")

	// Grab the latest Apple Emoji Definitions
	res := GetBodyByURL("https://emojipedia.org/apple/")

	// Load the Apple Emoji HTML page into goquery so that we
	// can query the DOM
	doc, err := goquery.NewDocumentFromResponse(res.Response)
	if err != nil {
		log.Panic(err)
	}

	// Create a channel for lookups so that we can do this async
	lookups := make(chan lookup)

	go func() {
		// Find all emojis on the page
		doc.Find("ul.emoji-grid li").Each(func(i int, s *goquery.Selection) {
			// For each item found, get the band and title
			emojiPage, _ := s.Find("a").Attr("href")
			title, _ := s.Find("img").Attr("title")

			log.Infof("Adding Emoji %d to lookups: %s - %s", i, title, emojiPage)

			// Add this specific emoji to the lookups to complete
			lookups <- lookup{
				Name: title,
				URL:  "https://emojipedia.org" + emojiPage,
			}
		})

		close(lookups)

	}()

	emojis := map[string]emoji.Emoji{}

	// Process a lookup
	for lookup := range lookups {
		log.Infof("Looking up %s", lookup.Name)

		// Grab the emojipedia page for this emoji
		res := GetBodyByURL(lookup.URL)

		// Create a new goquery reader
		doc, err := goquery.NewDocumentFromResponse(res.Response)
		if err != nil {
			log.Panic(err)
		}

		// Grab the emoji from the "Copy emoji" input field on the HTML page
		emojiString, _ := doc.Find(".copy-paste input[type=text]").Attr("value")

		// Convert the raw Emoji value to our hex key
		hexString := helper.StringToHexKey(emojiString)

		// Add this emoji to our map
		emojis[hexString] = emoji.Emoji{
			Key:        hexString,
			Value:      emojiString,
			Descriptor: lookup.Name,
		}

		// Print our progress to the console
		log.Info(emojis[hexString])
	}

	// Marshal the emojis map as JSON and write to the data directory
	s, _ := json.MarshalIndent(emojis, "", " ")
	ioutil.WriteFile("../emoji.json", []byte(s), 0644)
}

// GetBodyByURL is get body by url
func GetBodyByURL(url string) *req.Response {
	retryNum := 0
retry:
	log.Infof("Get: %s", url)
	res, err := reqClient.R().Get(url)
	if err != nil {
		if retryNum < 5 {
			retryNum++
			log.Infof(`Retry (%d) after 5 seconds: %s, Error: %s`, retryNum, url, err)
			time.Sleep(5 * time.Second)
			goto retry
		}
		log.Error(err)
	}
	return res
}
