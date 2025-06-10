package models

type User struct {
	ID     int64 `bun:",pk,autoincrement"`
	Name   string
	Emails []string
}
