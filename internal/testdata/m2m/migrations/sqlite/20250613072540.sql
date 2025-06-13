-- Create "items" table
CREATE TABLE `items` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT
);
-- Create "orders" table
CREATE TABLE `orders` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT
);
-- Create "order_to_items" table
CREATE TABLE `order_to_items` (
  `order_id` integer NOT NULL,
  `item_id` integer NOT NULL,
  PRIMARY KEY (`order_id`, `item_id`),
  CONSTRAINT `0` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `1` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
