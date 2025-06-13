package models

type OrderToItem struct {
	OrderID int64  `bun:",pk"`
	Order   *Order `bun:"rel:belongs-to,join:order_id=id"`
	ItemID  int64  `bun:",pk"`
	Item    *Item  `bun:"rel:belongs-to,join:item_id=id"`
}
