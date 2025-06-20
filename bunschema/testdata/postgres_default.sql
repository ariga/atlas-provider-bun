-- atlas:pos users[type=table] internal/testdata/models/user.go:3-7
-- atlas:pos stories[type=table] internal/testdata/models/story.go:3-8

CREATE TABLE "users" ("id" BIGSERIAL NOT NULL, "name" VARCHAR, "emails" JSONB, PRIMARY KEY ("id"));
CREATE TABLE "stories" ("id" BIGSERIAL NOT NULL, "title" VARCHAR, "author_id" BIGINT, PRIMARY KEY ("id"), FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
