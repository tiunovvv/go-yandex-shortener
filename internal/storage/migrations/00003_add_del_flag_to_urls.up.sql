BEGIN TRANSACTION;

ALTER TABLE urls
ADD COLUMN deleted_flag BOOLEAN;

COMMIT;