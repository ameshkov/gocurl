# gocurl

[![Go Report Card](https://goreportcard.com/badge/github.com/ameshkov/gocurl)](https://goreportcard.com/report/ameshkov/gocurl)
[![Latest release](https://img.shields.io/github/release/ameshkov/gocurl/all.svg)](https://github.com/ameshkov/gocurl/releases)

Simplified version of [`curl`](https://curl.se/) written in Go.

1. Supports a limited subset of curl options.
2. Supports some flags that curl does
   not. [Read more about the new stuff](#newstuff).

- [Why in the world you need another curl?](#why)
- [How to install gocurl?](#install)
- [How to use gocurl?](#howtouse)
- [New stuff](#newstuff)
    - [Encrypted ClientHello](#ech)
    - [Custom DNS servers](#dns)
    - [Oblivious HTTP](#ohttp)
    - [Experimental flags](#exp)
        - [Post-quantum cryptography](#pq)
    - [WebSocket support](#websocket)
- [All command-line arguments](#allcmdarguments)

<a id="why"></a>

## Why in the world you need another curl?

Curl is certainly awesome, but sometimes I need to have better control over
what's happening on the inside and to be able to debug it. It seemed easier to
me to implement the necessary parts of curl in Go.

Also, I'd like to be able to extend it with what fits my specific needs.
Unfortunately, curl is a bit too huge for that now.

<a id="install"></a>

## How to install gocurl?

- Using homebrew:

    ```shell
    brew install ameshkov/tap/gocurl
    ```

- From source:

    ```shell
    go install github.com/ameshkov/gocurl@latest
    ```

- You can use [a Docker image][dockerimage]:

    ```shell
    docker run --rm ghcr.io/ameshkov/gocurl --help
    ```

- You can get a binary from the [releases page][releases].

[dockerimage]: https://github.com/ameshkov/gocurl/pkgs/container/gocurl

[releases]: https://github.com/ameshkov/gocurl/releases

<a id="howtouse"></a>

## How to use gocurl?

Use it the same way you use original curl.

- `gocurl https://httpbin.agrd.workers.dev/get` make a `GET` request.
- `gocurl -d "test" -v https://httpbin.agrd.workers.dev/post` make a `POST`
  request with `test` data.
- `gocurl -I https://httpbin.agrd.workers.dev/head` make a `HEAD` request.
- `gocurl -I --insecure https://expired.badssl.com/` do not verify TLS
  certificate.
- `gocurl -I --http1.1 https://httpbin.agrd.workers.dev/head` force use
  HTTP/1.1.
- `gocurl -I --http2 https://httpbin.agrd.workers.dev/head` force use HTTP/2.
- `gocurl -I --http3 https://httpbin.agrd.workers.dev/head` force use HTTP/3.
- `gocurl -x socks5://user:pass@host:port https://httpbin.agrd.workers.dev/get`
  use a proxy server.
- `gocurl -I --tlsv1.3 https://tls-v1-2.badssl.com:1012/` force use TLS v1.3.
- `gocurl -I --connect-to "httpbin.agrd.workers.dev:443:172.67.152.85:443"
  https://httpbin.agrd.workers.dev/head` connect to the specified IP addresses.
- `gocurl -I --resolve "httpbin.agrd.workers.dev:443:172.67.152.85"
  https://httpbin.agrd.workers.dev/head` resolve the hostname to the specified
  IP address. Note, that unlike `curl`, `gocurl` ignores port in this option.
- `gocurl -I --connect-timeout 3 http://10.255.255.1:9999/test`
  set connection timeout to 3 seconds.

<a id="newstuff"></a>

### New stuff

Also, you can use some new stuff that is not supported by curl.

- `gocurl --json-output https://httpbin.agrd.workers.dev/get` write output in
  machine-readable format (JSON).
- `gocurl --tls-split-hello 5:50 https://httpbin.agrd.workers.dev/get` split
  TLS ClientHello in two parts and make a 50ms delay after sending the first
  part.
- `gocurl --tls-random "gyufwmGYeIiq0B4nUjEYu3NcqVdlHbIXhx74fq4terc=" https://httpbin.agrd.workers.dev/get`
  use a custom TLS ClientHello random value.
- `gocurl -v --ech https://crypto.cloudflare.com/cdn-cgi/trace` enables support
  for ECH (Encrypted Client Hello) for the request. More on this [below](#ech).
- `gocurl --dns-servers "tls://dns.google" https://httpbin.agrd.workers.dev/get`
  uses custom DNS-over-TLS server to resolve hostnames. More on this
  [below](#dns).
- `gocurl --experiment pq https://pq.cloudflareresearch.com/` enables
  post-quantum cryptography support for the request. More on this [below](#pq).
- `gocurl wss://httpbin.agrd.workers.dev/ws` sends a WS upgrade request.
- `gocurl -d "test message" wss://httpbin.agrd.workers.dev/ws` establishes a WS
  connection, sends the first message through it and reads the response.

<a id="ech"></a>

#### Encrypted ClientHello

ECH or Encrypted Client Hello is a new standard that allows completely
encrypting TLS Client Hello. Currently, the RFC is in the [draft stage][echrfc],
but it is already supported by some big names [like Cloudflare][echcloudflare].
`gocurl` supports ECH and provides several options to use it.

The simple option is just add `--ech` flag and see what happens:

```shell
gocurl -v --ech https://crypto.cloudflare.com/cdn-cgi/trace
```

In this case, `gocurl` will try to discover the ECH configuration from DNS
records and then use them to establish the connection.

Instead of that, you may choose to supply your own configuration in the same
base64-encoded format as used by the SVCB record:

```shell
# Send a type=https query and find ech record there.
% dig -t https crypto.cloudflare.com. +short
1 . alpn="http/1.1,h2" ipv4hint=162.159.137.85,162.159.138.85 ech=AEX+DQBBvgAgACARWS42g5NmDZo5pIpTWSwHzTwzdRKPdUW732QbyUeyDQAEAAEAAQASY2xvdWRmbGFyZS1lY2guY29tAAA= ipv6hint=2606:4700:7::a29f:8955,2606:4700:7::a29f:8a55

# You can now pass it to gocurl.
gocurl -v \
  --echconfig "AEX+DQBBvgAgACARWS42g5NmDZo5pIpTWSwHzTwzdRKPdUW732QbyUeyDQAEAAEAAQASY2xvdWRmbGFyZS1lY2guY29tAAA=" \
  https://crypto.cloudflare.com/cdn-cgi/trace
```

> Interesting thing about ECH is that it may connect even if you use an expired
> configuration (see HelloRetryRequest in the RFC). It depends on both the
> server and the client implementation and does not work with Cloudflare at the
> moment.

Here's what happens under the hood:

1. `gocurl` resolves `crypto.cloudflare.com` IP address and connects to it.
2. It sends TLS ClientHello (outer) with encrypted inner ClientHello to that IP
   address. The ServerName field in the outer ClientHello is set to the one that
   is encoded in the ECH configuration (in this example it will be
   `cloudflare-ech.com`), and in the inner encrypted ClientHello it will be
   set to `crypto.cloudflare.com`.

You may want to configure a specific "client-facing" server instead and the way
to do that is to use `--connect-to`. Let's send a request to `cloudflare.com`
and use `crypto.cloudflare.com` as a client-facing server for that.

```shell
gocurl -v \
  --connect-to "cloudflare.com:443:crypto.cloudflare.com:443" \
  --echconfig "AEX+DQBBvgAgACARWS42g5NmDZo5pIpTWSwHzTwzdRKPdUW732QbyUeyDQAEAAEAAQASY2xvdWRmbGFyZS1lY2guY29tAAA=" \
  https://cloudflare.com/cdn-cgi/trace
```

> For this command to work you may need to replace `--echconfig` with the
> current one discovered using DNS as was explained before.

Here's what happens now:

1. `gocurl` connects to `crypto.cloudflare.com` (client-facing relay).
2. Sends a TLS ClientHello with `cloudflare-ech.com` in the Server Name
   extension.
3. Establishes a TLS connection with `cloudflare.com` using the inner encrypted
   ClientHello.

[echrfc]: https://datatracker.ietf.org/doc/draft-ietf-tls-esni/

[echcloudflare]: https://blog.cloudflare.com/handshake-encryption-endgame-an-ech-update/

<a id="dns"></a>

#### Custom DNS servers

`gocurl` allows using custom DNS servers to resolve hostnames when making
requests. This can be achieved by using `--dns-servers` command-line argument.
`curl` with `c-ares` also supports this command-line argument, but `gocurl`
adds encrypted DNS support on top of it, it supports all popular DNS encryption
protocols: DNS-over-QUIC, DNS-over-HTTPS, DNS-over-TLS and DNSCrypt.

You can specify multiple DNS servers, in this case `gocurl` will attempt to use
them one by one until it receives a response or until all of them fail:

```shell
gocurl \
  --dns-servers "tls://dns.adguard-dns.com,tls://dns.google" \
  https://example.org/

```

- DNS-over-QUIC

  ```shell
  gocurl --dns-servers "quic://dns.adguard-dns.com" https://example.org/
  ```

- DNS-over-HTTPS

  ```shell
  gocurl --dns-servers "https://dns.adguard-dns.com/dns-query" https://example.org/
  ```

- DNS-over-TLS

  ```shell
  gocurl --dns-servers "tls://dns.adguard-dns.com" https://example.org/
  ```

- DNSCrypt

  ```shell
  gocurl \
      --dns-servers sdns://AQIAAAAAAAAAFDE3Ni4xMDMuMTMwLjEzMDo1NDQzINErR_JS3PLCu_iZEIbq95zkSV2LFsigxDIuUso_OQhzIjIuZG5zY3J5cHQuZGVmYXVsdC5uczEuYWRndWFyZC5jb20 \
      https://example.org/
  ```

<a id="ohttp"></a>

#### Oblivious HTTP

[Oblivious HTTP (OHTTP)][ohttp] is an IETF standard protocol that provides
end-to-end encryption for HTTP requests and responses while hiding the client's
identity from the target server. It works by routing encrypted requests through
a relay and gateway, ensuring that:

- The **relay** sees who is making requests but not what is being requested.
- The **gateway** sees what is being requested but not who is making the request.
- The **target server** receives a normal HTTP request from the gateway.

This separation provides strong privacy guarantees, making OHTTP useful for
privacy-sensitive applications.

`gocurl` has built-in support for OHTTP and can send requests through any
OHTTP gateway by specifying two command-line arguments:

- `--ohttp-gateway-url` - URL of the OHTTP gateway where the encrypted request
  will be sent.
- `--ohttp-keys-url` - URL from which to retrieve the OHTTP KeyConfig needed to
  encrypt the request.

Here's how to make an OHTTP request to `httpbin.agrd.workers.dev` using a demo
gateway:

```shell
gocurl -v \
  --ohttp-gateway-url "https://httpbin.agrd.workers.dev/ohttp/gateway" \
  --ohttp-keys-url "https://httpbin.agrd.workers.dev/ohttp/config" \
  https://httpbin.agrd.workers.dev/get
```

This command will:

1. Download the OHTTP KeyConfig from the keys URL.
2. Encrypt your request to `https://httpbin.agrd.workers.dev/get` using OHTTP.
3. Send the encrypted request to the gateway.
4. Receive the encrypted response from the gateway.
5. Decrypt and display the response.

You can also make POST requests through OHTTP:

```shell
gocurl -v \
  --ohttp-gateway-url "https://httpbin.agrd.workers.dev/ohttp/gateway" \
  --ohttp-keys-url "https://httpbin.agrd.workers.dev/ohttp/config" \
  -d "test data" \
  https://httpbin.agrd.workers.dev/post
```

One more example that uses a demo gateway from [Oblivious Network][ohttpdemo]:

```shell
gocurl -v \
  --ohttp-gateway-url "https://demo-gateway.oblivious.network/gateway" \
  --ohttp-keys-url "https://demo-gateway.oblivious.network/ohttp-configs" \
  https://httpbin.agrd.workers.dev/get
```

[ohttp]: https://www.ietf.org/rfc/rfc9458.html

[ohttpdemo]: https://docs.oblivious.network/docs/quickstart/

<a id="websocket"></a>

#### WebSocket support

`gocurl` provides some initial support for WebSocket protocol. It may be
extended in the future, see the corresponding [Github issue][wsissue].

- `gocurl wss://httpbin.agrd.workers.dev/ws` sends a WS upgrade request.
- `gocurl -d "test message" wss://httpbin.agrd.workers.dev/ws` establishes a WS
  connection, sends the first message through it and reads the response.

[wsissue]: https://github.com/ameshkov/gocurl/issues/17

<a id="exp"></a>

#### Experimental flags

Experimental flags are added to `gocurl` whenever there's a feature that may be
completely changed or removed in the future. Experiments can be enabled using
the `--experiment=<name[:value]>` argument where `name` is the experiment name
and `value` is an optional string value (the need for it depends on the actual
experiment).

<a id="pq"></a>

##### Post-quantum cryptography

Post-quantum (PQ) cryptography has been designed to be secure against the
threat of quantum computers. You can learn more about it from Cloudflare's
[blog post][postquantum]. `gocurl` supports it via the `--experiment=pq` flag.

Note, that it is not available for `--http3` at the moment.

```shell
gocurl --experiment pq https://pq.cloudflareresearch.com/
```

[postquantum]: https://blog.cloudflare.com/post-quantum-for-all/

<a id="allcmdarguments"></a>

## All command-line arguments

```shell
Usage:
  gocurl [OPTIONS]

Application Options:
      --url=<URL>                                           URL the request will be made to. Can be specified without
                                                            any flags.
  -X, --request=<method>                                    HTTP method. GET by default.
  -d, --data=<data>                                         Sends the specified data to the HTTP server using content
                                                            type application/x-www-form-urlencoded.
  -H, --header=                                             Extra header to include in the request. Can be specified
                                                            multiple times.
  -x, --proxy=[protocol://username:password@]host[:port]    Use the specified proxy. The proxy string can be specified
                                                            with a protocol:// prefix.
      --connect-to=<HOST1:PORT1:HOST2:PORT2>                For a request to the given HOST1:PORT1 pair, connect to
                                                            HOST2:PORT2 instead. Can be specified multiple times.
      --connect-timeout=<seconds>                           Maximum time in seconds allowed for the connection phase.
  -I, --head                                                Fetch the headers only.
  -k, --insecure                                            Disables TLS verification of the connection.
      --tlsv1.3                                             Forces gocurl to use TLS v1.3 or newer.
      --tlsv1.2                                             Forces gocurl to use TLS v1.2 or newer.
      --tls-max=<VERSION>                                   (TLS) VERSION defines maximum supported TLS version. Can be
                                                            1.2 or 1.3. The minimum acceptable version is set by
                                                            tlsv1.2 or tlsv1.3.
      --ciphers=<space-separated list of ciphers>           Specifies which ciphers to use in the connection, see
                                                            https://go.dev/src/crypto/tls/cipher_suites.go for the full
                                                            list of available ciphers.
      --tls-servername=<HOSTNAME>                           Specifies the server name that will be sent in TLS
                                                            ClientHello
      --http1.1                                             Forces gocurl to use HTTP v1.1.
      --http2                                               Forces gocurl to use HTTP v2.
      --http3                                               Forces gocurl to use HTTP v3.
      --ech                                                 Enables ECH support for the request.
      --echgrease                                           Forces sending ECH grease in the ClientHello, but does not
                                                            try to resolve the ECH configuration.
      --echconfig=<base64-encoded data>                     ECH configuration to use for this request. Implicitly
                                                            enables --ech when specified.
  -4, --ipv4                                                This option tells gocurl to use IPv4 addresses only when
                                                            resolving host names.
  -6, --ipv6                                                This option tells gocurl to use IPv6 addresses only when
                                                            resolving host names.
      --dns-servers=<DNSADDR1,DNSADDR2>                     DNS servers to use when making the request. Supports
                                                            encrypted DNS: tls://, https://, quic://, sdns://
      --resolve=<[+]host:port:addr[,addr]...>               Provide a custom address for a specific host. port is
                                                            ignored by gocurl. '*' can be used instead of the host
                                                            name. Can be specified multiple times.
      --tls-split-hello=<CHUNKSIZE:DELAY>                   An option that allows splitting TLS ClientHello in two
                                                            parts in order to avoid common DPI systems detecting TLS.
                                                            CHUNKSIZE is the size of the first bytes before ClientHello
                                                            is split, DELAY is delay in milliseconds before sending the
                                                            second part.
      --tls-random=<base64>                                 Base64-encoded 32-byte TLS ClientHello random value.
      --json-output                                         Makes gocurl write machine-readable output in JSON format.
  -o, --output=<file>                                       Defines where to write the received data. If not set,
                                                            gocurl will write everything to stdout.
      --experiment=<name[:value]>                           Allows enabling experimental options. See the documentation
                                                            for available options. Can be specified multiple times.
      --ohttp-gateway-url=<URL>                             URL of the Oblivious HTTP gateway where the request should
                                                            be sent.
      --ohttp-keys-url=<URL>                                URL from which to retrieve Oblivious HTTP KeyConfig to use
                                                            for encrypting the request.
  -v, --verbose                                             Verbose output (optional).

Help Options:
  -h, --help                                                Show this help message
```
