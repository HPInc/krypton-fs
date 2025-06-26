-- Initial schema for the files database.
-- Create the buckets table.
CREATE TABLE buckets
(
  bucket_name VARCHAR(64) UNIQUE NOT NULL,
  is_archived BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY(bucket_name)
);

-- Create the files table.
CREATE TABLE files
(
  file_id BIGSERIAL NOT NULL,
  tenant_id VARCHAR(36) NOT NULL,
  device_id VARCHAR(36) NOT NULL,
  name VARCHAR(128) NOT NULL,
  checksum VARCHAR(64) NOT NULL,
  size BIGINT,
  status VARCHAR(10) NOT NULL,
  created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
  bucket_name VARCHAR(64) NOT NULL,
  PRIMARY KEY(file_id),
  CONSTRAINT fk_bucket 
    FOREIGN KEY(bucket_name) 
      REFERENCES buckets(bucket_name)
);

-- Create an index to enable queries for expired files. This
-- is used by the DB scavenger goroutine.
CREATE INDEX idx_files_created_at ON files(created_at);

-- Create the tombstoned files table.
CREATE TABLE tombstoned_files
(
  file_id BIGINT NOT NULL,
  tenant_id VARCHAR(36) NOT NULL,
  device_id VARCHAR(36) NOT NULL,
  deleted_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
  bucket_name VARCHAR(64) NOT NULL,
  PRIMARY KEY(file_id),
  CONSTRAINT fk_tf_bucket 
    FOREIGN KEY(bucket_name) 
      REFERENCES buckets(bucket_name)
);

-- Create an index to enable queries for tombstoned files. This
-- is used by the DB scavenger goroutine.
CREATE INDEX idx_tombstoned_files_deleted_at ON tombstoned_files(deleted_at)
