FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/server

FROM scratch
COPY --from=build /app/server /app/server
WORKDIR /app
EXPOSE 8000
ENTRYPOINT ["/app/server"]
