-- Create "users" table
CREATE TABLE "public"."users" (
  "id" bigserial NOT NULL,
  "name" character varying NULL,
  "emails" jsonb NULL,
  PRIMARY KEY ("id")
);
-- Create "stories" table
CREATE TABLE "public"."stories" (
  "id" bigserial NOT NULL,
  "title" character varying NULL,
  "author_id" bigint NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "stories_author_id_fkey" FOREIGN KEY ("author_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
