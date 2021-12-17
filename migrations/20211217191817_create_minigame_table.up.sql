CREATE TABLE IF NOT EXISTS MiniGame(
    id bigint UNSIGNED NOT NULL  AUTO_INCREMENT,
    `name` longtext NOT NULL,
    roomId bigint UNSIGNED NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (roomId) REFERENCES Room(id)
) AUTO_INCREMENT=1;