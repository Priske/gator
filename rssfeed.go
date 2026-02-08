package main

import (
	"context"
	"encoding/xml"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	rsf := RSSFeed{}
	err = xml.Unmarshal(data, &rsf)
	if err != nil {
		return nil, err
	}
	rsf.Channel.Title = html.UnescapeString(rsf.Channel.Title)
	rsf.Channel.Description = html.UnescapeString(rsf.Channel.Description)

	for i := range rsf.Channel.Item {
		rsf.Channel.Item[i].Title = html.UnescapeString(rsf.Channel.Item[i].Title)
		rsf.Channel.Item[i].Description = html.UnescapeString(rsf.Channel.Item[i].Description)
	}
	return &rsf, nil
}
