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

## How to use gocurl

Use it the same way you use original curl.

* `gocurl https://httpbin.agrd.workers.dev/get`
* `gocurl -d "test" -v https://httpbin.agrd.workers.dev/post`
* `gocurl -I https://httpbin.agrd.workers.dev/head`
* `gocurl -I --insecure https://expired.badssl.com/`

```shell
% gocurl --help

Usage:
  gocurl [OPTIONS]

Application Options:
      --url=<URL>                                           URL the request will be made to. Can be specified without any flags.
  -X, --request=<method>                                    HTTP method. GET by default.
  -I, --head                                                Fetch the headers only.
  -k, --insecure                                            Disables TLS verification of the connection.
  -d, --data=<data>                                         Sends the specified data to the HTTP server using content type
                                                            application/x-www-form-urlencoded.
  -H, --header=                                             Extra header to include in the request. Can be specified multiple times.
  -x, --proxy=[protocol://username:password@]host[:port]    Use the specified proxy. The proxy string can be specified with a protocol://
                                                            prefix.
      --connect-to=<HOST1:PORT1:HOST2:PORT2>                For a request to the given HOST1:PORT1 pair, connect to HOST2:PORT2 instead.
  -o, --output=<file>                                       Defines where to write the received data. If not set, gocurl will write
                                                            everything to stdout.
  -v, --verbose                                             Verbose output (optional).

Help Options:
  -h, --help                                                Show this help message
```