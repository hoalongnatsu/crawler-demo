package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-redis/redis"
	"github.com/gocolly/colly/v2"
	"github.com/joho/godotenv"
)

type Post struct {
	Title string   `json:"title"`
	Link  string   `json:"link"`
	Tags  []string `json:"tags"`
}

func main() {
	godotenv.Load()

	// Redis
	rc := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASS"),
	})

	_, err := rc.Ping().Result()
	if err != nil {
		log.Fatal("Unbale to connect to Redis ", err)
	}

	log.Println("Connected to Redis server")

	c := colly.NewCollector(
		colly.AllowedDomains("viblo.asia"),
	)

	c.OnHTML(".post-feed .post-title--inline", func(e *colly.HTMLElement) {
		p := Post{}

		e.ForEach(".link", func(i int, h *colly.HTMLElement) {
			link := h.Attr("href")

			if strings.Contains(link, "/p/") {
				p.Title = h.Text
				p.Link = link
			}
		})

		e.ForEach(".el-tag--info", func(i int, h *colly.HTMLElement) {
			link := h.Attr("href")

			if strings.Contains(link, "/tags/") {
				p.Tags = append(p.Tags, strings.TrimSpace(h.Text))
			}
		})

		t, err := json.Marshal(p.Tags)
		if err != nil {
			log.Fatal(err)
		}

		if p.Link != "" {
			err = rc.XAdd(&redis.XAddArgs{
				Stream:       "posts",
				MaxLen:       0,
				MaxLenApprox: 0,
				ID:           "",
				Values: map[string]interface{}{
					"title": p.Title,
					"link":  p.Link,
					"tags":  t,
				},
			}).Err()

			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("-------")
			fmt.Printf("Discover post: %v\n", p)
		}

	})

	c.Visit("https://viblo.asia/newest")
}
