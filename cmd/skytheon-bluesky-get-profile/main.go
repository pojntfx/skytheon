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

	flag.Parse()

	if strings.TrimSpace(*did) == "" {
		panic(errMissingDID)
	}

	if strings.TrimSpace(*accessToken) == "" {
		panic(errMissingAccessToken)
	}

	b := backends.NewBluesky(*api)

	profile, err := b.GetProfile(*accessToken, *did)
	if err != nil {
		panic(err)
	}

	if err := yaml.NewEncoder(os.Stdout).Encode(profile); err != nil {
		panic(err)
	}
}
