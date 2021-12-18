#!/bin/bash

# migration
migrate -path "./migrations" -database "mysql://root:$DB_PWD@tcp(orientation_db)/utopia" up

# start server
go run main.go