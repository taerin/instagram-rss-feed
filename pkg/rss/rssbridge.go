package rss

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type AtomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Title   string      `xml:"title"`
	Link    []AtomLink  `xml:"link"`
	Updated string      `xml:"updated"`
	Entries []AtomEntry `xml:"entry"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
}

type AtomEntry struct {
	ID      string      `xml:"id"`
	Title   string      `xml:"title"`
	Link    []AtomLink  `xml:"link"`
	Updated string      `xml:"updated"`
	Content AtomContent `xml:"content"`
	Author  AtomAuthor  `xml:"author"`
}

type AtomContent struct {
	Type string `xml:"type,attr"`
	Text string `xml:",chardata"`
}

type AtomAuthor struct {
	Name string `xml:"name"`
}

// RSS-Bridge 서버 목록 (공개 인스턴스들)
var rssBridgeServers = []string{
	"https://rss-bridge.org/bridge01/",
	"https://wtf.roflcopter.fr/rss-bridge/",
	"https://rssbridge.flossboxin.org.in/",
}

func FetchFromRSSBridge(username string) ([]Post, error) {
	for _, server := range rssBridgeServers {
		posts, err := tryRSSBridgeServer(server, username)
		if err == nil && len(posts) > 0 {
			fmt.Printf("Successfully fetched data from: %s\n", server)
			return posts, nil
		}
		fmt.Printf("Failed to fetch from %s: %v\n", server, err)
	}
	return nil, fmt.Errorf("all RSS-Bridge servers failed")
}

func tryRSSBridgeServer(server, username string) ([]Post, error) {
	bridgeURL := fmt.Sprintf("%s?bridge=Instagram&context=Username&format=Atom&server_url=%s", server, url.QueryEscape(username))

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest("GET", bridgeURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Instagram-RSS-Feed/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseAtomFeed(body, username)
}

func parseAtomFeed(xmlData []byte, username string) ([]Post, error) {
	var feed AtomFeed
	if err := xml.Unmarshal(xmlData, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse Atom feed: %v", err)
	}

	var posts []Post
	maxPosts := 8
	if len(feed.Entries) < maxPosts {
		maxPosts = len(feed.Entries)
	}

	for i := 0; i < maxPosts; i++ {
		entry := feed.Entries[i]
		post := convertAtomEntryToPost(entry, username)
		posts = append(posts, post)
	}

	return posts, nil
}

func convertAtomEntryToPost(entry AtomEntry, username string) Post {
	post := Post{
		ID:      extractIDFromAtomEntry(entry),
		Caption: entry.Title,
		PostURL: extractLinkFromAtomEntry(entry),
	}

	if timestamp, err := time.Parse(time.RFC3339, entry.Updated); err == nil {
		post.Timestamp = timestamp
	} else {
		post.Timestamp = time.Now()
	}

	post.ImageURL = extractImageURLFromContent(entry.Content.Text)
	post.IsVideo = strings.Contains(entry.Content.Text, "video") || strings.Contains(entry.Content.Text, ".mp4")
	post.Code = extractShortcodeFromURL(post.PostURL)
	post.Likes = 0
	post.Comments = 0

	return post
}

func extractIDFromAtomEntry(entry AtomEntry) string {
	id := entry.ID
	if strings.Contains(id, "/") {
		parts := strings.Split(id, "/")
		return parts[len(parts)-1]
	}
	return id
}

func extractLinkFromAtomEntry(entry AtomEntry) string {
	for _, link := range entry.Link {
		if link.Rel == "alternate" || link.Rel == "" {
			return link.Href
		}
	}
	return ""
}

func extractImageURLFromContent(content string) string {
	imgStart := strings.Index(content, "<img")
	if imgStart == -1 {
		return ""
	}
	imgStart += len("<img")

	imgEnd := strings.Index(content[imgStart:], ">")
	if imgEnd == -1 {
		return ""
	}
	return content[imgStart : imgStart+imgEnd]
}

func extractShortcodeFromURL(postURL string) string {
	if strings.Contains(postURL, "/p/") {
		parts := strings.Split(postURL, "/p/")
		if len(parts) > 1 {
			shortcode := strings.Split(parts[1], "/")[0]
			return shortcode
		}
	}
	return ""
}
