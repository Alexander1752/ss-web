FROM golang:1.25-alpine AS build-env

WORKDIR /app
COPY . /app

# Use the platform provided by buildx/compose instead of forcing amd64
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o mqtt ./main.go
FROM scratch
COPY --from=build-env /app/mqtt /mqtt

EXPOSE 8080

CMD ["./mqtt"]
