CREATE TABLE `users` (`id` bigint AUTO_INCREMENT,
                        `email` varchar(191) NOT NULL UNIQUE,
                        `name` longtext,
                        `username` varchar(191) DEFAULT null,
                        `description` longtext,
                        `gender` longtext,
                        `department` varchar(191) DEFAULT null,
                        `refresh_token` varchar(191) DEFAULT null,
                        `sprite_sheet_id` bigint,PRIMARY KEY (`id`),CONSTRAINT 
                        `fk_users_spritesheet` FOREIGN KEY (`sprite_sheet_id`) REFERENCES `sprite_sheets`(`id`))