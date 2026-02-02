FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/custom_exporter ./cmd/exporter

FROM alpine:3.19
WORKDIR /app
COPY --from=build /out/custom_exporter /app/custom_exporter
EXPOSE 9200
ENTRYPOINT ["/app/custom_exporter"]
