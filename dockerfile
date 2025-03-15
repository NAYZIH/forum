FROM golang:1.23.1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o forum ./main.go

EXPOSE 1945

CMD ["./forum"]