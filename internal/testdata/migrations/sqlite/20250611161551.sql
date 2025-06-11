-- Create "users" table
CREATE TABLE `users` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `name` varchar NULL,
  `emails` varchar NULL
);
-- Create "stories" table
CREATE TABLE `stories` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `title` varchar NULL,
  `author_id` integer NULL,
  CONSTRAINT `0` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
