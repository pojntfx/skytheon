package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/pojntfx/skytheon/pkg/backends"
)

var (
	errMissingDID         = errors.New("missing DID")
	errMissingAccessToken = errors.New("missing access token")
	errMissingContent     = errors.New("missing content")
)

func main() {
	api := flag.String("api", backends.BlueskySocialURL, "Blusky API URL")
	did := flag.String("did", "", "DID to use")
	accessToken := flag.String("access-token", "", "Access token to use")
	content := flag.String("content", "", "Text content to post")

	flag.Parse()

	if strings.TrimSpace(*did) == "" {
		panic(errMissingDID)
	}

	if strings.TrimSpace(*accessToken) == "" {
		panic(errMissingAccessToken)
	}

	if strings.TrimSpace(*content) == "" {
		panic(errMissingContent)
	}

	b := backends.NewBluesky(*api)

	uri, err := b.CreatePost(*accessToken, *did, *content)
	if err != nil {
		panic(err)
	}

	fmt.Println(uri)
}
