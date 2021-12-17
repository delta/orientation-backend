CREATE TABLE IF NOT EXISTS User(
    id bigint UNSIGNED NOT NULL AUTO_INCREMENT,
    email varchar(191) NOT NULL UNIQUE,
    name longtext,
    userName varchar(191) DEFAULT null,
    description longtext,
    gender longtext,
    department varchar(191) DEFAULT null,
    refreshToken varchar(191) DEFAULT null,
    PRIMARY KEY (id)
) AUTO_INCREMENT=1;
