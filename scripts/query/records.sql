SELECT id, user_id, metadata, detailtype FROM records
WHERE records.user_id = :user_id;
