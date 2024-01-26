package main

import "github.com/uptrace/bun"

type Post struct {
	bun.BaseModel `bun:"table:posts,alias:p"`

	ID    int64    `bun:"id,pk,autoincrement"`
	Title string   `bun:"title,notnull"`
	Link  string   `bun:"link,notnull"`
	Tags  []string `bun:"tags,jsonb"`
}
