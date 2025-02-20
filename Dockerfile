FROM golang:1.23.5-alpine

WORKDIR /app

COPY . .

RUN go build -o out && ./out

EXPOSE 50051

CMD ["./auth-service"]

# docker build -t auth-service:latest .