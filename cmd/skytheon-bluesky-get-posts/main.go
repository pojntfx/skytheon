package main

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/pojntfx/skytheon/pkg/backends"
	"gopkg.in/yaml.v3"
)

var (
	errMissingDID         = errors.New("missing DID")
	errMissingAccessToken = errors.New("missing access token")
)

func main() {
	api := flag.String("api", backends.BlueskySocialURL, "Blusky API URL")
	did := flag.String("did", "", "DID to use")
	accessToken := flag.String("access-token", "", "Access token to use")
	limit := flag.Int("limit", 100, "Maximum amount of posts to search (including reposts etc.)")

	flag.Parse()

	if strings.TrimSpace(*did) == "" {
		panic(errMissingDID)
	}

	if strings.TrimSpace(*accessToken) == "" {
		panic(errMissingAccessToken)
	}

	b := backends.NewBluesky(*api)

	posts, err := b.GetPosts(*accessToken, *did, *limit)
	if err != nil {
		panic(err)
	}

	if err := yaml.NewEncoder(os.Stdout).Encode(posts); err != nil {
		panic(err)
	}
}
