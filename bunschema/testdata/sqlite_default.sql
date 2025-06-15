-- atlas:pos users[type=table] internal/testdata/models/user.go:3-7
-- atlas:pos stories[type=table] internal/testdata/models/story.go:3-8

CREATE TABLE "users" ("id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, "name" VARCHAR, "emails" VARCHAR);
CREATE TABLE "stories" ("id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, "title" VARCHAR, "author_id" INTEGER, FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
