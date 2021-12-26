#!/bin/bash

# migration
./wait.sh -t 0 orientation_db:3306 && migrate -path "./migrations" -database "mysql://root:$DB_PWD@tcp(orientation_db)/utopia" up

# start server
go run main.go
