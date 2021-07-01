CREATE TABLE IF NOT EXISTS `entry_task`.`User` (
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

CREATE TABLE IF NOT EXISTS `entry_task`.`Session` (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    session VARCHAR(255) NOT NULL,
    UNIQUE INDEX idx_username (username)
)
COLLATE = utf8mb4_unicode_ci
ENGINE = innodb
DEFAULT charset=utf8mb4;
