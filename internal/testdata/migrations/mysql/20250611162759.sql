-- Create "users" table
CREATE TABLE `users` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NULL,
  `emails` varchar(255) NULL,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "stories" table
CREATE TABLE `stories` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `title` varchar(255) NULL,
  `author_id` bigint NULL,
  PRIMARY KEY (`id`),
  INDEX `author_id` (`author_id`),
  CONSTRAINT `stories_ibfk_1` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
