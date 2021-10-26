## Backend for Utopia - Orientation'21

### Prerequistes

- Go 1.16+
- MySQL

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

- Run `cp .sample.env .env`, fill in the env variables
- Run `go run main.go`
