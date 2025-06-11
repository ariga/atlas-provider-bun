CREATE TABLE "users" ("id" BIGSERIAL NOT NULL, "name" VARCHAR, "emails" JSONB, PRIMARY KEY ("id"));
CREATE TABLE "stories" ("id" BIGSERIAL NOT NULL, "title" VARCHAR, "author_id" BIGINT, PRIMARY KEY ("id"));
