FROM golang:1.17-alpine as compiler

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /proxy


FROM scratch
WORKDIR /
COPY --from=compiler /proxy /

EXPOSE 80

ENTRYPOINT ["/proxy"]

