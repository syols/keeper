INSERT INTO text_details (record_id, data)
VALUES (:record_id, :data) ON CONFLICT (record_id) DO
    UPDATE
    SET record_id = excluded.record_id,
        data = excluded.data;