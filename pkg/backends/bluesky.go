package backends

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
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
		URI    string `json:"uri"`
		Record struct {
			Type      string `json:"$type"`
			Text      string `json:"text"`
			CreatedAt string `json:"createdAt"`
			Reply     struct {
				Parent any `json:"parent"`
			} `json:"reply"`
		} `json:"record"`
		ReplyCount  int `json:"replyCount"`
		RepostCount int `json:"repostCount"`
		LikeCount   int `json:"likeCount"`
		Author      struct {
			Handle      string `json:"handle"`
			DisplayName string `json:"displayName"`
			Avatar      string `json:"avatar"`
		} `json:"author"`
	} `json:"post"`
	Reason any `json:"reason"`
}

type createPostInput struct {
	Collection string                 `json:"collection"`
	Repo       string                 `json:"repo"`
	Record     createPostInputContent `json:"record"`
}

type createPostInputContent struct {
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
	Type      string `json:"$type"`
}

type createPostOutput struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

type FeedItem struct {
	URI               string `json:"uri"`
	Text              string `json:"text"`
	CreatedAt         string `json:"createdAt"`
	ReplyCount        int    `json:"replyCount"`
	RepostCount       int    `json:"repostCount"`
	LikeCount         int    `json:"likeCount"`
	AuthorHandle      string `json:"authorHandle"`
	AuthorDisplayName string `json:"authorDisplayName"`
	AuthorAvatar      string `json:"authorAvatar"`
}

func idToURI(id string) string {
	return "at://" + strings.TrimPrefix(id, "at://")
}

func uriToID(uri string) string {
	return strings.TrimPrefix(uri, "at://")
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

	return idToURI(r.DID), nil
}

func (b *Bluesky) GetAccessToken(did, appPassword string) (string, error) {
	u, err := url.Parse(b.url)
	if err != nil {
		return "", err
	}

	u = u.JoinPath("xrpc", "com.atproto.server.createSession")

	inputJson, err := json.Marshal(createSessionInput{
		Identifier: uriToID(did),
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
	p.Add("actor", uriToID(did))
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
				URI:               item.Post.URI,
				Text:              item.Post.Record.Text,
				CreatedAt:         item.Post.Record.CreatedAt,
				ReplyCount:        item.Post.ReplyCount,
				RepostCount:       item.Post.RepostCount,
				LikeCount:         item.Post.LikeCount,
				AuthorHandle:      item.Post.Author.Handle,
				AuthorDisplayName: item.Post.Author.DisplayName,
				AuthorAvatar:      item.Post.Author.Avatar,
			})
		}
	}

	return feedItems, nil
}

func (b *Bluesky) CreatePost(accessToken, did, text string) (string, error) {
	u, err := url.Parse(b.url)
	if err != nil {
		return "", err
	}

	u = u.JoinPath("xrpc", "com.atproto.repo.createRecord")

	postRecord := createPostInput{
		Collection: "app.bsky.feed.post",
		Repo:       uriToID(did),
		Record: createPostInputContent{
			Text:      text,
			CreatedAt: time.Now().Format(time.RFC3339),
			Type:      "app.bsky.feed.post",
		},
	}

	inputJson, err := json.Marshal(postRecord)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(inputJson))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

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

	var output createPostOutput
	if err := json.NewDecoder(res.Body).Decode(&output); err != nil {
		return "", err
	}

	return output.URI, nil
}
