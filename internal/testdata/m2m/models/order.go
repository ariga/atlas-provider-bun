package models

type Order struct {
	ID int64 `bun:",pk,autoincrement"`
	// Order and Item in join:Order=Item are fields in OrderToItem model
	Items []Item `bun:"m2m:order_to_items,join:Order=Item"`
}
