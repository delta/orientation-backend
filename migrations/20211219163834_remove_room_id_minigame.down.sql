ALTER TABLE MiniGame
ADD COLUMN `roomId` bigint UNSIGNED NOT NULL,
ADD CONSTRAINT `MiniGame_ibfk_1` 
FOREIGN KEY (`roomId`)
REFERENCES `Room` (`id`);
