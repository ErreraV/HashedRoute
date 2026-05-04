# syntax=docker/dockerfile:1

FROM golang:1.24-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/hashedroute-api .

FROM gcr.io/distroless/static-debian12
WORKDIR /
COPY --from=build /out/hashedroute-api /hashedroute-api

EXPOSE 8080
ENTRYPOINT ["/hashedroute-api"]
