INSERT INTO login_details (record_id, number, cardholder, cvc, expiration)
VALUES (:record_id, :number, :cardholder, :cvc, :expiration) ON CONFLICT (record_id) DO
    UPDATE
    SET record_id = excluded.record_id,
        number = excluded.number,
        cardholder = excluded.cardholder,
        cvc = excluded.cvc,
        expiration = excluded.expiration,