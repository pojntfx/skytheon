package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v3"
)

var errMissingUsername = fmt.Errorf("missing username")

type FeedItem struct {
	URL               string `json:"url"`
	Text              string `json:"text"`
	CreatedAt         string `json:"createdAt"`
	AuthorHandle      string `json:"authorHandle"`
	AuthorDisplayName string `json:"authorDisplayName"`
	AuthorAvatar      string `json:"authorAvatar"`
}

func containsNitterLink(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" && strings.Contains(attr.Val, "nitter.net") {
				return true
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if containsNitterLink(c) {
			return true
		}
	}
	return false
}

func main() {
	api := flag.String("api", "https://nitter.net/", "Nitter API URL")
	usernameFlag := flag.String("username", "", "Twitter username to use")
	limit := flag.Int("limit", 100, "Maximum amount of tweets to search (including retweets etc.)")

	flag.Parse()

	if strings.TrimSpace(*usernameFlag) == "" {
		panic(errMissingUsername)
	}

	parser := gofeed.NewParser()

	u, err := url.Parse(*api)
	if err != nil {
		panic(err)
	}
	u.Path = path.Join(u.Path, *usernameFlag, "rss")

	feed, err := parser.ParseURL(u.String())
	if err != nil {
		panic(err)
	}

	names := strings.Split(feed.Title, " / ")

	userDisplayName := names[0]
	userName := strings.TrimPrefix(names[1], "@")
	userProfilePictureURL := feed.Image.URL

	tweets := make([]FeedItem, 0)

	count := 0
	for _, sourceTweet := range feed.Items {
		if strings.HasPrefix(sourceTweet.Title, "R to "+names[1]) || strings.HasPrefix(sourceTweet.Title, "QT by") || strings.HasPrefix(sourceTweet.Title, "RT by") {
			continue
		}

		doc, err := html.Parse(strings.NewReader(sourceTweet.Description))
		if err != nil {
			panic(err)
		}

		if containsNitterLink(doc) {
			continue
		}

		if count >= *limit {
			break
		}

		count++

		createdAt, err := time.Parse(time.RFC1123, sourceTweet.Published)
		if err != nil {
			panic(err)
		}

		tweet := FeedItem{
			CreatedAt:         createdAt.Format(time.RFC3339),
			Text:              sourceTweet.Description,
			URL:               sourceTweet.Link,
			AuthorHandle:      userName,
			AuthorDisplayName: userDisplayName,
			AuthorAvatar:      userProfilePictureURL,
		}

		tweets = append(tweets, tweet)
	}

	if err := yaml.NewEncoder(os.Stdout).Encode(tweets); err != nil {
		panic(err)
	}
}
