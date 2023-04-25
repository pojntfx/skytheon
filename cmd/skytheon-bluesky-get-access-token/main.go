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
	errMissingAppPassword = errors.New("missing app password")
)

func main() {
	api := flag.String("api", backends.BlueskySocialURL, "Blusky API URL")
	did := flag.String("did", "", "DID to use")
	appPassword := flag.String("app-password", "", "App password to use (get one from https://staging.bsky.app/settings/app-passwords)")

	flag.Parse()

	if strings.TrimSpace(*did) == "" {
		panic(errMissingDID)
	}

	if strings.TrimSpace(*appPassword) == "" {
		panic(errMissingAppPassword)
	}

	b := backends.NewBluesky(*api)

	token, err := b.GetAccessToken(*did, *appPassword)
	if err != nil {
		panic(err)
	}

	fmt.Println(token)
}
