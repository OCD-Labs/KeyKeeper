CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "full_name" varchar NOT NULL,
  "hashed_password" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "is_activated" boolean NOT NULL DEFAULT true
);

CREATE TABLE "sessions" (
  "id" uuid PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "refresh_token" varchar NOT NULL,
  "user_agent" varchar NOT NULL,
  "client_ip" varchar NOT NULL,
  "is_blocked" boolean NOT NULL DEFAULT false,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "reminders" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "website_url" varchar NOT NULL,
  "interval" varchar NOT NULL,
  "updated_at" timestamptz NOT NULL DEFAULT (now()),
  "extension" jsonb
);

CREATE INDEX ON "reminders" ("user_id", "website_url");

ALTER TABLE "sessions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "reminders" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
