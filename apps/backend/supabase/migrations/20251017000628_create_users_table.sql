CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS "users" (
  "user_id" uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  "username" varchar(100) UNIQUE NOT NULL,
  "role" varchar,
  "created_at" timestamp DEFAULT now(),
  "updated_at" timestamp DEFAULT now()
);
