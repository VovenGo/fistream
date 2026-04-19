FROM golang:1.24-bookworm AS build

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/fistream ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=build /out/fistream /app/fistream
COPY --from=build /src/web /app/web

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/fistream"]

