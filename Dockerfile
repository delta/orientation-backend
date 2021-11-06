FROM golang:1.17

WORKDIR  /go/src/github.com/delta/orientation-backend

COPY go.mod go.sum ./

RUN go mod download 

COPY . .

RUN go mod download

CMD bash -c "./wait.sh -t 0 orientation_db:3306 && go run main.go"
