DELIMITER $$

CREATE PROCEDURE insertData()
BEGIN
  DECLARE num INT DEFAULT 0;

  WHILE num < 10000 DO
    INSERT INTO User(username, nickname, password, avatar) VALUES(CONCAT("USER_", num), CONCAT("USER_", num), "12345678", "");
    SET num = num + 1;
  End WHILE;
End$$


-- /usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/go/bin
-- /usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/mysql/bin