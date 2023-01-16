SELECT id, user_id, metadata, detail FROM records
WHERE records.user_id = :user_id;
