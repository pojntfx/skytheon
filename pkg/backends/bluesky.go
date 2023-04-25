package backends

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const (
	BlueskySocialURL = "https://bsky.social/"
)

type resolveHandleOutput struct {
	DID string `json:"did"`
}

type createSessionInput struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type createSessionOutput struct {
	AccessJwt string `json:"accessJwt"`
}

type feedItemInput struct {
	Post struct {
		Record struct {
			Type      string `json:"$type"`
			Text      string `json:"text"`
			CreatedAt string `json:"createdAt"`
			Reply     struct {
				Parent any `json:"parent"`
			} `json:"reply"`
		} `json:"record"`
		ReplyCount  int    `json:"replyCount"`
		RepostCount int    `json:"repostCount"`
		LikeCount   int    `json:"likeCount"`
		Author      Author `json:"author"`
	} `json:"post"`
	Reason any `json:"reason"`
}

type FeedItem struct {
	Text        string `json:"text"`
	CreatedAt   string `json:"createdAt"`
	ReplyCount  int    `json:"replyCount"`
	RepostCount int    `json:"repostCount"`
	LikeCount   int    `json:"likeCount"`
	Author      Author `json:"author"`
}

type Author struct {
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
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

	var r resolveHandleOutput
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return "", err
	}

	return r.DID, nil
}

func (b *Bluesky) GetAccessToken(did, appPassword string) (string, error) {
	u, err := url.Parse(b.url)
	if err != nil {
		return "", err
	}

	u = u.JoinPath("xrpc", "com.atproto.server.createSession")

	inputJson, err := json.Marshal(createSessionInput{
		Identifier: did,
		Password:   appPassword,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(inputJson))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
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

	var r createSessionOutput
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return "", err
	}

	return r.AccessJwt, nil
}

func (b *Bluesky) GetPosts(accessToken, did string, limit int) ([]FeedItem, error) {
	u, err := url.Parse(b.url)
	if err != nil {
		return nil, err
	}

	u = u.JoinPath("xrpc", "app.bsky.feed.getAuthorFeed")

	p := u.Query()
	p.Add("actor", did)
	p.Add("limit", strconv.Itoa(limit))
	u.RawQuery = p.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		return nil, errors.New(string(body))
	}

	var rawFeed struct {
		Feed []feedItemInput `json:"feed"`
	}

	if err := json.NewDecoder(res.Body).Decode(&rawFeed); err != nil {
		return nil, err
	}

	var feedItems []FeedItem
	for _, item := range rawFeed.Feed {
		if item.Post.Record.Type == "app.bsky.feed.post" &&
			item.Post.Record.Reply.Parent == nil &&
			item.Reason == nil {

			feedItems = append(feedItems, FeedItem{
				Text:        item.Post.Record.Text,
				CreatedAt:   item.Post.Record.CreatedAt,
				ReplyCount:  item.Post.ReplyCount,
				RepostCount: item.Post.RepostCount,
				LikeCount:   item.Post.LikeCount,
				Author:      item.Post.Author,
			})
		}
	}

	return feedItems, nil
}
