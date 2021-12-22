## Backend for Utopia - Orientation'21

### Prerequisites

- Go 1.16+
- MySQL
- redis [download link](https://redis.io/download)
- [air](https://github.com/cosmtrek/air) (for live reload) [installation link](https://github.com/cosmtrek/air#prefer-installsh)

### Migrations

- Install

```bash
go get -tags 'mysql'  -u github.com/golang-migrate/migrate/v4/cmd/migrate/
```

- Create Migrations

```bash
migrate create -ext sql -dir ./migrations <MIGRATION_NAME>
```

- Up Migrations

```bash
migrate -path "./migrations" -database "mysql://root:YOUR_MYSQL_PASSWORD@/DB_NAME" up
```

- Down Migrations

```bash
migarte -path "./migrations"  -database "mysql://root:YOUR_MYSQL_PASSWORD@/DB_NAME" down
```

### Setup

- Run `cp .sample.env .env`, fill the env variables
- Run `cp dauth.config.example.json dauth.config.json`, fill the required creds
- start server

  - development

    ```bash
    air
    ```

  - production
    ```bash
    go run main.go
    ```

### Docker Setup
 - Install docker and docker-compose
 - Run `cp .docker.sample.env .docker.env`, fill the env varaiables
 - Run `docker-compose up` to start the services
