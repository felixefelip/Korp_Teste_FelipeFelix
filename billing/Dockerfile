FROM golang:1.26

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8001

CMD ["sleep", "infinity"]
