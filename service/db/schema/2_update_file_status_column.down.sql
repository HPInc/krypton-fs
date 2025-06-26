-- rollback column change introduced by version 2
ALTER TABLE files ALTER COLUMN status TYPE VARCHAR(10);
