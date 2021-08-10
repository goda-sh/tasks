package tasks

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/mmcdole/gofeed"
	"pkg.goda.sh/utils"
)

type item struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Published   string `json:"published"`
}

// Feed pulls different types of RSS feeds
func Feed(args *TaskArgs) Result {
	result := NewResult(args.Task)
	params := utils.ParamsParser(args.Task.Params, utils.DefaultParams{
		"limit": 5,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	fp := gofeed.NewParser()
	fp.UserAgent = fmt.Sprintf("%s/%s", Project, Version)
	feed, err := fp.ParseURLWithContext(params.Get("url").String(), ctx)
	if err != nil {
		result.Error = err
	} else {
		items := []item{}
		for _, i := range feed.Items {
			items = append(items, item{
				Title:       i.Title,
				Description: i.Description,
				Link:        i.Link,
				Published:   i.Published,
			})
			if int(params.Get("limit").Int64()) <= len(items) {
				break
			}
		}
		result.Update = struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Items       []item `json:"items"`
		}{
			Title:       feed.Title,
			Description: feed.Description,
			Items:       items,
		}
	}
	return result
}

// FakeFeed generates fake data for demo dashboards
func FakeFeed(args *TaskArgs) Result {
	result := NewResult(args.Task)
	limit := 5
	if args.Task.Params["limit"] != nil {
		limit = int(args.Task.Params["limit"].(float64))
	}
	fp := gofeed.NewParser()
	feed, err := fp.ParseString(genFakeFeed(limit))
	if err != nil {
		result.Error = err
	} else {
		items := []item{}
		for _, i := range feed.Items {
			items = append(items, item{
				Title:       i.Title,
				Description: i.Description,
				Link:        i.Link,
				Published:   i.Published,
			})
			if limit <= len(items) {
				break
			}
		}
		result.Update = struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Items       []item `json:"items"`
		}{
			Title:       feed.Title,
			Description: feed.Description,
			Items:       items,
		}
	}
	return result
}

func genFakeFeed(limit int) string {
	feed := `<?xml version="1.0" encoding="UTF-8" ?>
	<rss version="2.0">
	<channel>
		<title>Demo RSS Feed</title>
		<link>https://github.com/LouisT/godash</link>
		<description>Example RSS feed!</description>
	`
	for i := 1; i <= limit; i++ {
		feed += Templater(`<item>
			<title>This is example #{{ .ID }} for "FakeFeed" generator.</title>
			<link>https://example.org/item-{{ .ID }}</link>
			<description>Example RSS item #{{ .ID }}</description>
			<pubDate>{{ .Date }}</pubDate>
		</item>`, struct {
			ID   int
			Date string
		}{
			ID:   i,
			Date: randate().Format("Mon, 02 Jan 2006 15:04 MST"),
		})
	}
	feed += `</channel></rss>`
	return feed
}

func randate() time.Time {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(time.Now().Local().Year(), 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	return time.Unix(rand.Int63n(max-min)+min, 0)
}
