INSERT INTO records (id, user_id, metadata, detail)
VALUES (:id, :user_id, :metadata, :detail) ON CONFLICT (metadata) DO
    UPDATE
    SET user_id = excluded.user_id,
        metadata = excluded.metadata,
        detail = excluded.detail;