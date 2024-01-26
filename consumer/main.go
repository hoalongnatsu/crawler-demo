package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func main() {
	godotenv.Load()

	// Redis
	rc := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASS"),
	})

	_, err := rc.Ping().Result()
	if err != nil {
		log.Fatal("Consumer unbale to connect to Redis", err)
	}

	log.Println("Consumer connected to Redis server")

	// Postgres
	dsn := os.Getenv("POSTGRES_DNS")
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	db := bun.NewDB(sqldb, pgdialect.New())

	/** For consumer group
	err = rc.XGroupCreate("posts", "posts-consumer-group", "0").Err()
	if err != nil {
		log.Println(err)
	}

	id := xid.New().String()
	*/

	for {
		/** For consumer group
		entries, err := rc.XReadGroup(&redis.XReadGroupArgs{
			Group:    "posts-consumer-group",
			Consumer: id,
			Streams:  []string{"posts", ">"},
			Count:    5,
			Block:    0,
			NoAck:    false,
		}).Result()
		*/

		entries, err := rc.XRead(&redis.XReadArgs{
			Streams: []string{"posts", "0"},
			Count:   5,
		}).Result()

		if err != nil {
			log.Fatal(err)
		}

		for i := 0; i < len(entries[0].Messages); i++ {
			ctx := context.Background()
			values := entries[0].Messages[i].Values

			link := fmt.Sprintf("%v", values["link"])

			exists, err := db.NewSelect().Model((*Post)(nil)).Where("link = ?", link).Exists(ctx)
			if err != nil {
				log.Fatal("Query error: ", err)
			}

			if !exists {
				p := &Post{}
				p.Link = link
				p.Title = fmt.Sprintf("%v", values["title"])

				err = json.Unmarshal([]byte(fmt.Sprintf("%v", values["tags"])), &p.Tags)
				if err != nil {
					log.Fatal("Unmarshal tags error: ", err)
				}

				fmt.Println("-------")
				fmt.Printf("Insert new link: %v\n", p)

				_, err = db.NewInsert().Model(p).Exec(ctx)
				if err != nil {
					log.Fatal(err)
				}
			}

			rc.XDel("posts", entries[0].Messages[i].ID)
		}
	}
}
