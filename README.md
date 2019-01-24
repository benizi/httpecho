# httpecho

Just an HTTP Echo server.

# Intro

Responds to HTTP requests with a `text/plain` dump of the request.

I got sick of trying to make an HTTP echo server using `socat`¹.

# Usage

```sh
$ docker network create --subnet 10.11.12.0/24 testnet
0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
$ docker run --network testnet -d --name httpecho benizi/httpecho
456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123
$ docker run --network testnet --rm appropriate/curl http://httpecho/wat
Method: GET
Path: /wat
Request Headers:
  User-Agent: curl/7.59.0
  Accept: */*
```

# Features

- [x] Prints anything whatsoever in response to an HTTP request

# License

Copyright © 2019 Benjamin R. Haskell

Distributed under the MIT License (included in file: [LICENSE](LICENSE)).

#### Notes

¹: Attempts at making a `socat` HTTP echo server

```sh
## (whispers:) The quoting. The quoting.
socat -d -d -d \
  tcp-l:8888,reuseaddr,fork \
  system:'printf '\\\''%s\\r\\n'\\\'' HTTP/1.1\\ 200\\ OK Host\:\\ $(hostname)\:8888 Content-Type\:\\ text/plain \"\" Blah'

socat -d -d -d \
  tcp-l:8888,reuseaddr,fork \
  system:'echo \"HTTP/1.1 200 OK\"; echo \"Host\: $(hostname)\:8888\"; echo \"Content-Type\: text/plain\"; echo \"\"; echo Blah'

socat -d -d -d \
  tcp-l:8888,reuseaddr,fork \
  system:'set -- SFRUUC8xLjEgMjAwIE9LDQpIb3N0OiA= $(hostname | base64)
Ojg4ODgNCkNvbnRlbnQtVHlwZTogdGV4dC9wbGFpbg0KDQpCbGFoCg==;
for b; do echo $b | base64 -d; done'
```
