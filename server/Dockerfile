FROM golang:1.25-alpine AS build-env

WORKDIR /app
COPY . /app

# install tesseract with english language data
RUN apk add --no-cache \
    tesseract-ocr \
    tesseract-ocr-dev \
    leptonica-dev \
    g++ \
    tesseract-ocr-data-eng \
    tesseract-ocr-data-ron \
    ffmpeg \
    && rm -rf /var/cache/apk/*

# Use the platform provided by buildx/compose instead of forcing amd64
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=1 GOOS=${TARGETOS:-linux} go build -o mqtt ./main.go
FROM alpine@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c
WORKDIR /app
COPY --from=build-env /app/mqtt /app/

RUN apk add --no-cache \
    tesseract-ocr \
    tesseract-ocr-dev \
    leptonica-dev \
    g++ \
    tesseract-ocr-data-eng \
    tesseract-ocr-data-ron \
    ffmpeg \
    && rm -rf /var/cache/apk/*


EXPOSE 8080

CMD ["./mqtt"]
