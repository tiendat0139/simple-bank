# Build stage
FROM golang:1.24.4-alpine3.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

# Final stage
FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate
COPY app.env .
COPY start.sh .
COPY db/migration ./db/migration

EXPOSE 8080
CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]
