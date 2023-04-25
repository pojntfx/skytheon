package backends

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

const (
	BlueskySocialURL = "https://bsky.social/"
)

type resolveHandleResponse struct {
	DID string `json:"did"`
}

type Bluesky struct {
	url string
}

func NewBluesky(url string) *Bluesky {
	return &Bluesky{url}
}

func (b *Bluesky) ResolveHandle(handle string) (string, error) {
	u, err := url.Parse(b.url)
	if err != nil {
		return "", err
	}

	u = u.JoinPath("xrpc", "com.atproto.identity.resolveHandle")

	p := u.Query()
	p.Add("handle", handle)
	u.RawQuery = p.Encode()

	res, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return "", err
		}

		return "", errors.New(string(body))
	}

	var r resolveHandleResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return "", err
	}

	return r.DID, nil
}
