FROM golang:1.23.2

WORKDIR /app

COPY . .

RUN go build -o filestorage

EXPOSE 8080

CMD ["./filestorage"]