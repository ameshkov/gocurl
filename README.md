[![Go Report Card](https://goreportcard.com/badge/github.com/ameshkov/gocurl)](https://goreportcard.com/report/ameshkov/gocurl)
[![Latest release](https://img.shields.io/github/release/ameshkov/gocurl/all.svg)](https://github.com/ameshkov/gocurl/releases)

# gocurl

Simplified version of [`curl`](https://curl.se/) written in Golang.

1. Supports a limited subset of curl options.
2. Supports some flags that curl does not (for instance, `json-output`).

## Why in the world you need another curl?

Curl is certainly awesome, but sometimes I need to have a better control over
what's happening on the inside and be able to debug it. It seemed easier to me
to rewrite the necessary parts of curl.

Also, I'd like to be able to extend it with what fits my specific needs.
Unfortunately, curl is a bit too huge for that now.

## How to install gocurl

* Using homebrew:
    ```shell
    brew install ameshkov/tap/gocurl
    ```
* From source:
    ```shell
    go install github.com/ameshkov/gocurl@latest
    ```
* You can get a binary from
  the [releases page](https://github.com/ameshkov/gocurl/releases).

## How to use gocurl

Use it the same way you use original curl.

* `gocurl https://httpbin.agrd.workers.dev/get` make a `GET` request.
* `gocurl -d "test" -v https://httpbin.agrd.workers.dev/post` make a `POST`
  request with `test` data.
* `gocurl -I https://httpbin.agrd.workers.dev/head` make a `HEAD` request.
* `gocurl -I --insecure https://expired.badssl.com/` do not verify TLS
  certificate.
* `gocurl -I --http1.1 https://httpbin.agrd.workers.dev/head` force use
  HTTP/1.1.
* `gocurl -I --http2 https://httpbin.agrd.workers.dev/head` force use HTTP/2.
* `gocurl -I --http3 https://httpbin.agrd.workers.dev/head` force use HTTP/3.
* `gocurl -x socks5://user:pass@host:port https://httpbin.agrd.workers.dev/get`
  use a proxy server.
* `gocurl -I --tlsv1.3 https://tls-v1-2.badssl.com:1012/` force use TLS v1.3.

Also, you can use some new stuff.

* `gocurl --json-output https://httpbin.agrd.workers.dev/get` write output in
  machine-readable format (JSON).
* `gocurl --tls-split-hello=5:50 https://httpbin.agrd.workers.dev/get` split
  TLS ClientHello in two parts and make a 50ms delay after sending the first
  part.

## All command-line arguments

```shell
% gocurl --help

Usage:
  gocurl [OPTIONS]

Application Options:
      --url=<URL>                                           URL the request will be made to. Can be specified without any flags.
  -X, --request=<method>                                    HTTP method. GET by default.
  -d, --data=<data>                                         Sends the specified data to the HTTP server using content type
                                                            application/x-www-form-urlencoded.
  -H, --header=                                             Extra header to include in the request. Can be specified multiple times.
  -x, --proxy=[protocol://username:password@]host[:port]    Use the specified proxy. The proxy string can be specified with a protocol://
                                                            prefix.
      --connect-to=<HOST1:PORT1:HOST2:PORT2>                For a request to the given HOST1:PORT1 pair, connect to HOST2:PORT2 instead.
  -I, --head                                                Fetch the headers only.
  -k, --insecure                                            Disables TLS verification of the connection.
      --tlsv1.3                                             Forces gocurl to use TLS v1.3.
      --tlsv1.2                                             Forces gocurl to use TLS v1.2.
      --http1.1                                             Forces gocurl to use HTTP v1.1.
      --http2                                               Forces gocurl to use HTTP v2.
      --http3                                               Forces gocurl to use HTTP v2.
      --tls-split-hello=<CHUNKSIZE:DELAY>                   An option that allows splitting TLS ClientHello in two parts in order to avoid
                                                            common DPI systems detecting TLS. CHUNKSIZE is the size of the first bytes
                                                            before ClientHello is split, DELAY is delay in milliseconds before sending the
                                                            second part.
      --json-output                                         Makes gocurl write machine-readable output in JSON format.
  -o, --output=<file>                                       Defines where to write the received data. If not set, gocurl will write
                                                            everything to stdout.
  -v, --verbose                                             Verbose output (optional).

Help Options:
  -h, --help                                                Show this help message
```