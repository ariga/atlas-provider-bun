package models

type Item struct {
	ID int64 `bun:",pk,autoincrement"`
	// Order and Item in join:Order=Item are fields in OrderToItem model
	Orders []Order `bun:"m2m:order_to_items,join:Item=Order"`
}
