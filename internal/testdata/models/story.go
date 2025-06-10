package models

type Story struct {
	ID       int64 `bun:",pk,autoincrement"`
	Title    string
	AuthorID int64
	Author   *User `bun:"rel:belongs-to,join:author_id=id"`
}
