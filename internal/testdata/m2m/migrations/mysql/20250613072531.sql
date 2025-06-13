-- Create "items" table
CREATE TABLE `items` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "orders" table
CREATE TABLE `orders` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "order_to_items" table
CREATE TABLE `order_to_items` (
  `order_id` bigint NOT NULL,
  `item_id` bigint NOT NULL,
  PRIMARY KEY (`order_id`, `item_id`),
  INDEX `item_id` (`item_id`),
  CONSTRAINT `order_to_items_ibfk_1` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `order_to_items_ibfk_2` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
