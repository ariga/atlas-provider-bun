-- Create "stories" table
CREATE TABLE "public"."stories" (
  "id" bigserial NOT NULL,
  "title" character varying NULL,
  "author_id" bigint NULL,
  PRIMARY KEY ("id")
);
-- Create "users" table
CREATE TABLE "public"."users" (
  "id" bigserial NOT NULL,
  "name" character varying NULL,
  "emails" jsonb NULL,
  PRIMARY KEY ("id")
);
