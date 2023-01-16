INSERT INTO login_details (record_id, login, password)
VALUES (:record_id, :login, :password) ON CONFLICT (record_id) DO
    UPDATE
    SET record_id = excluded.record_id,
        login = excluded.login,
        password = excluded.password;