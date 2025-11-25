# build
FROM golang:1.24.0-alpine AS build

# Install git (required for go mod download)
RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod go.sum ./

# Use Go proxy with fallback
RUN go env -w GOPROXY=https://proxy.golang.org,direct

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./cmd/library-api

# runtime
FROM gcr.io/distroless/static:nonroot

COPY --from=build /app /app

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app"]
