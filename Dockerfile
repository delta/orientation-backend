FROM golang:1.17

WORKDIR  /go/src/github.com/delta/orientation-backend

RUN  go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY go.mod go.sum ./

RUN go mod download 

COPY . .

CMD bash docker-entry.sh
