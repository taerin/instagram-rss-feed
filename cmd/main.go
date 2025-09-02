package main

import (
	"fmt"
	"log"
	"os"

	"instagram-rss-feed/pkg/rss"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <instagram_username>")
		fmt.Println("Example: go run . blackpinkofficial")
		os.Exit(1)
	}

	username := os.Args[1]
	fmt.Printf("Fetching Instagram posts for @%s...\n", username)

	// 인스타그램 포스트 가져오기
	posts, err := rss.FetchInstagramFeed(username)
	if err != nil {
		log.Fatalf("Error fetching Instagram feed: %v", err)
	}

	if len(posts) == 0 {
		log.Fatalf("No posts found for @%s", username)
	}

	fmt.Printf("Found %d posts\n", len(posts))

	// RSS 생성
	rssContent, err := rss.GenerateRSSFeed(posts, username)
	if err != nil {
		log.Fatalf("Error generating RSS: %v", err)
	}

	// RSS 파일 저장
	filename := fmt.Sprintf("%s_feed.xml", username)
	err = os.WriteFile(filename, []byte(rssContent), 0644)
	if err != nil {
		log.Fatalf("Error writing RSS file: %v", err)
	}

	fmt.Printf("RSS feed generated successfully: %s\n", filename)
	fmt.Printf("Posts included:\n")

	for i, post := range posts {
		fmt.Printf("  %d. %s (%s)\n",
			i+1,
			truncateString(post.Caption, 50),
			post.Timestamp.Format("2006-01-02 15:04"),
		)
	}
}

// 문자열을 길이 제한 후 ... 붙여주는 함수
func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}
