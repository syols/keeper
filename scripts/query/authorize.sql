SELECT id, login, password FROM users
WHERE login = :login
  AND password = :password;