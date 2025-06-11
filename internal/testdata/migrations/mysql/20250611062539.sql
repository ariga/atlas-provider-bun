-- Create "stories" table
CREATE TABLE `stories` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `title` varchar(255) NULL,
  `author_id` bigint NULL,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "users" table
CREATE TABLE `users` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NULL,
  `emails` varchar(255) NULL,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
