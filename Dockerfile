# syntax=docker/dockerfile:1

## Build step
FROM golang:1.18-alpine AS build

WORKDIR /app

ENV CGO_ENABLED 0

COPY go.* ./
COPY *.go ./

RUN go build -o /kube-net-prober


## Deploy step
FROM gcr.io/distroless/base-debian10 as production

WORKDIR /

COPY --from=build /kube-net-prober /kube-net-prober

USER nonroot:nonroot

ENTRYPOINT ["/kube-net-prober"]