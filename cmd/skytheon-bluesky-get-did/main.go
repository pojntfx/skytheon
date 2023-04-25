package main

import (
	"flag"
	"fmt"

	"github.com/pojntfx/skytheon/pkg/backends"
)

func main() {
	api := flag.String("api", backends.BlueskySocialURL, "Blusky API URL")
	handle := flag.String("handle", "felicitas.pojtinger.com", "Handle to resolve")

	flag.Parse()

	b := backends.NewBluesky(*api)

	did, err := b.GetDID(*handle)
	if err != nil {
		panic(err)
	}

	fmt.Println(did)
}
