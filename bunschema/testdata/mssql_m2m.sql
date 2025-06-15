-- atlas:pos items[type=table] internal/testdata/m2m/models/item.go:3-7
-- atlas:pos orders[type=table] internal/testdata/m2m/models/order.go:3-7
-- atlas:pos order_to_items[type=table] internal/testdata/m2m/models/orderitem.go:3-8

CREATE TABLE "items" ("id" BIGINT NOT NULL IDENTITY, PRIMARY KEY ("id"))
GO
CREATE TABLE "orders" ("id" BIGINT NOT NULL IDENTITY, PRIMARY KEY ("id"))
GO
CREATE TABLE "order_to_items" ("order_id" BIGINT NOT NULL, "item_id" BIGINT NOT NULL, PRIMARY KEY ("order_id", "item_id"), FOREIGN KEY ("item_id") REFERENCES "items" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION)
GO
