-- Table
CREATE TABLE IF NOT EXISTS "builds" (
  build_uuid TEXT PRIMARY KEY,
  run_uuid TEXT NOT NULL,
  create_time TIMESTAMP WITH TIME ZONE NOT NULL,
  bbid BIGINT NOT NULL,
  build_target TEXT NOT NULL,
  milestone INT NOT NULL,
  version TEXT NOT NULL,
  image_path TEXT NOT NULL,
  board TEXT NOT NULL,
  variant TEXT DEFAULT ''
);


-- Roles
CREATE ROLE kronwriter LOGIN;
GRANT INSERT ON "builds" TO kronwriter;

CREATE ROLE kronreader LOGIN;
GRANT SELECT ON "builds" TO kronreader;