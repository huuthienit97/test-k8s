FROM golang:1.23-alpine AS build
WORKDIR /src
ARG GIT_SHA=""
ARG GIT_REF=""
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X main.buildSHA=${GIT_SHA} -X main.buildRef=${GIT_REF}" \
    -o /app ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /app ./app
EXPOSE 8080
CMD ["./app"]
