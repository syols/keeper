INSERT INTO records (user_id, metadata, detailtype)
VALUES (:user_id, :metadata, :detailtype) ON CONFLICT (metadata) DO
    UPDATE
    SET user_id = excluded.user_id,
        metadata = excluded.metadata,
        detailtype = excluded.detailtype
RETURNING id, user_id, metadata, detailtype;