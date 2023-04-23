CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "full_name" varchar NOT NULL,
  "hashed_password" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "profile_image_url" varchar,
  "password_changed_at" timestamptz NOT NULL DEFAULT (now()),
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "is_active" boolean NOT NULL DEFAULT false,
  "is_email_verified" boolean NOT NULL DEFAULT false
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

CREATE OR REPLACE FUNCTION delete_expired_sessions()
RETURNS void AS $$
BEGIN
  DELETE FROM sessions WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;
