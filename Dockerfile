FROM golang:1.10-alpine as builder
ADD httpecho.go /src/
WORKDIR /src
RUN go build -o /httpecho

FROM alpine
LABEL maintainer="Benjamin R. Haskell <go@benizi.com>"
COPY --from=builder /httpecho /
ENTRYPOINT /httpecho
