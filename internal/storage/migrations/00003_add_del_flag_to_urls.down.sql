BEGIN TRANSACTION;

ALTER TABLE urls
DROP COLUMN deleted_flag;

COMMIT;