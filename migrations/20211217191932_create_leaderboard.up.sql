CREATE TABLE IF NOT EXISTS LeaderBoard(
    miniGameId  bigint UNSIGNED NOT NULL,
    userId  bigint UNSIGNED NOT NULL,
    score  bigint  NOT NULL,
    FOREIGN KEY(miniGameId) REFERENCES MiniGame(id),
    FOREIGN KEY(userId) REFERENCES User(id)
);
