DELIMITER $$

CREATE PROCEDURE createTable()
Begin
  CREATE TABLE IF NOT EXISTS `User` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    avatar VARCHAR(255),
    nickname VARCHAR(255),
    UNIQUE INDEX idx_username (username)
  )
  COLLATE = utf8mb4_unicode_ci
  ENGINE = innodb
  DEFAULT charset=utf8mb4;

  CREATE TABLE IF NOT EXISTS `Session` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    session VARCHAR(255) NOT NULL,
    UNIQUE INDEX idx_username (username)
  )
  COLLATE = utf8mb4_unicode_ci
  ENGINE = innodb
  DEFAULT charset=utf8mb4;
END$$

CREATE PROCEDURE insertData()
BEGIN
  DECLARE num INT DEFAULT 0;

  WHILE num < 10000 DO
    INSERT INTO User(username, nickname, password, avatar) VALUES(CONCAT("USER_", num), CONCAT("USER_", num), "12345678", "");
    SET num = num + 1;
  End WHILE;
End$$
