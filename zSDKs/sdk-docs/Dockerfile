FROM golang:1.21-alpine as builder

WORKDIR /app

RUN go mod init server

# Copy the server.go file.
COPY server.go ./

# Copy the 'out' directory
COPY out/ ./out/

RUN go build -o /server

FROM gcr.io/distroless/base

WORKDIR /

COPY --from=builder /server /server
COPY --from=builder /app/out/ /out/

ENV PORT=8080

EXPOSE 8080

ENTRYPOINT ["/server"]
