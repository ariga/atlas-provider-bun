-- atlas:pos users[type=table] internal/testdata/models/user.go:3-7
-- atlas:pos stories[type=table] internal/testdata/models/story.go:3-8

CREATE TABLE `users` (`id` BIGINT NOT NULL AUTO_INCREMENT, `name` VARCHAR(255), `emails` VARCHAR(255), PRIMARY KEY (`id`));
CREATE TABLE `stories` (`id` BIGINT NOT NULL AUTO_INCREMENT, `title` VARCHAR(255), `author_id` BIGINT, PRIMARY KEY (`id`), FOREIGN KEY (`author_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION);
