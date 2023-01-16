SELECT id, record_id, number, cardholder, cvc, expiration FROM card_details
WHERE record_id = :id;