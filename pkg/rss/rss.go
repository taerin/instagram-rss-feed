package rss

import (
	"encoding/xml"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	Language      string `xml:"language"`
	LastBuildDate string `xml:"lastBuildDate"`
	Generator     string `xml:"generator"`
	Items         []Item `xml:"item"`
}

type Item struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	PubDate     string    `xml:"pubDate"`
	GUID        string    `xml:"guid"`
	Enclosure   Enclosure `xml:"enclosure"`
}

type Enclosure struct {
	URL    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

func GenerateRSSFeed(posts []Post, username string) (string, error) {
	rss := RSS{
		Version: "2.0",
		Channel: Channel{
			Title:         fmt.Sprintf("%s's Instagram Feed", username),
			Link:          fmt.Sprintf("https://www.instagram.com/%s/", username),
			Description:   fmt.Sprintf("Latest 8 posts from @%s Instagram account", username),
			Language:      "en-us",
			LastBuildDate: time.Now().Format(time.RFC1123Z),
			Generator:     "Instagram RSS",
		},
	}

	for _, post := range posts {
		item := createRSSItem(post)
		rss.Channel.Items = append(rss.Channel.Items, item)
	}

	output, err := xml.MarshalIndent(rss, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to generate RSS XML: %v", err)
	}

	xmlHeader := `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
	return xmlHeader + string(output), nil
}

func createRSSItem(post Post) Item {
	title := createItemTitle(post)
	description := createItemDescription(post)

	item := Item{
		Title:       title,
		Link:        post.PostURL,
		Description: description,
		PubDate:     post.Timestamp.Format(time.RFC1123Z),
		GUID:        post.PostURL,
	}

	if post.IsVideo && post.VideoURL != "" {
		item.Enclosure = Enclosure{
			URL:    post.VideoURL,
			Type:   "video/mp4",
			Length: "0",
		}
	} else if post.ImageURL != "" {
		item.Enclosure = Enclosure{
			URL:    post.ImageURL,
			Type:   "image/jpeg",
			Length: "0",
		}
	}

	return item
}

func createItemTitle(post Post) string {
	lines := strings.Split(post.Caption, "\n")
	title := lines[0]
	if len(title) > 50 {
		title = title[:50] + "..."
	}

	words := strings.Fields(title)
	var cleanWords []string
	for _, word := range words {
		if !strings.HasPrefix(word, "#") && !strings.HasPrefix(word, "@") {
			cleanWords = append(cleanWords, word)
		}
	}

	if len(cleanWords) > 0 {
		return strings.Join(cleanWords, " ")
	}
	return ""
}

func createItemDescription(post Post) string {
	var desc strings.Builder

	if post.IsVideo && post.VideoURL != "" {
		desc.WriteString(fmt.Sprintf("<video controls style=\"max-width:100%%;height:auto;\"><source src=\"%s\" type=\"video/mp4\">Your browser does not support the video tag.</video><br><br>", html.EscapeString(post.VideoURL)))
	} else if post.ImageURL != "" {
		desc.WriteString(fmt.Sprintf("<img src=\"%s\" alt=\"Instagram post\" style=\"max-width:100%%;height:auto;\"><br><br>", html.EscapeString(post.ImageURL)))
	}

	if post.Caption != "" {
		caption := html.EscapeString(post.Caption)
		caption = convertHashtags(caption)
		caption = convertMentions(caption)
		caption = strings.ReplaceAll(caption, "\n", "<br>")
		desc.WriteString(fmt.Sprintf("<p>%s</p>", caption))
	}

	desc.WriteString("<!-- comments -->")
	desc.WriteString(fmt.Sprintf("<small>👍 %d likes • 💬 %d comments</small><br>", post.Likes, post.Comments))
	desc.WriteString(fmt.Sprintf("<small>🔗 <a href=\"%s\">View on Instagram</a></small>", post.PostURL))

	return desc.String()
}

func convertHashtags(text string) string {
	hashtagRegex := regexp.MustCompile(`#(\w+)`)
	return hashtagRegex.ReplaceAllStringFunc(text, func(match string) string {
		hashtag := strings.TrimPrefix(match, "#")
		return fmt.Sprintf("<a href=\"https://www.instagram.com/explore/tags/%s\">#%s</a>", hashtag, hashtag)
	})
}

func convertMentions(text string) string {
	mentionRegex := regexp.MustCompile(`@(\w+)`)
	return mentionRegex.ReplaceAllStringFunc(text, func(match string) string {
		username := strings.TrimPrefix(match, "@")
		return fmt.Sprintf("<a href=\"https://www.instagram.com/%s\">@%s</a>", username, username)
	})
}
