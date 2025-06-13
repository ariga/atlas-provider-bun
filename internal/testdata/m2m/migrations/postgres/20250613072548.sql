-- Create "items" table
CREATE TABLE "public"."items" (
  "id" bigserial NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "orders" table
CREATE TABLE "public"."orders" (
  "id" bigserial NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "order_to_items" table
CREATE TABLE "public"."order_to_items" (
  "order_id" bigint NOT NULL,
  "item_id" bigint NOT NULL,
  PRIMARY KEY ("order_id", "item_id"),
  CONSTRAINT "order_to_items_item_id_fkey" FOREIGN KEY ("item_id") REFERENCES "public"."items" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "order_to_items_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "public"."orders" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
