FROM golang:1.17

WORKDIR  /go/src/github.com/delta/orientation-backend

RUN  go get -tags 'mysql' -u github.com/golang-migrate/migrate/v4/cmd/migrate/

COPY go.mod go.sum ./

RUN go mod download 

COPY . .

CMD bash docker-entry.sh
