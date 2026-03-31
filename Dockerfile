FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /visa-tracker ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /visa-tracker .
COPY data/ data/
COPY static/ static/
COPY internal/templates/ internal/templates/
EXPOSE 8080
CMD ["./visa-tracker"]
