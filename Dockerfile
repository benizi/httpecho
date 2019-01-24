FROM golang:1.10
LABEL maintainer="Benjamin R. Haskell <go@benizi.com>"
ADD httpecho.go /src/
WORKDIR /src
RUN go build -o /httpecho
ENTRYPOINT /httpecho
