# gocurl changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][keepachangelog], and this project
adheres to [Semantic Versioning][semver].

[keepachangelog]: https://keepachangelog.com/en/1.0.0/

[semver]: https://semver.org/spec/v2.0.0.html

## [Unreleased]

### Added

* Added support for the `--tls-servername` argument.

### Changed

* Use `http` as the default scheme.  ([#27][#27])
* Updated dependencies, now using newer versions of quic-go, dnsproxy, cfcrypto.

[#27]: https://github.com/ameshkov/gocurl/issues/27

[unreleased]: https://github.com/ameshkov/gocurl/compare/v1.4.2...HEAD

## [1.4.2] - 2024-03-31

### Added

* Added support for the `--tls-max` argument.
* Added support for the `--ciphers` argument.

[See changes][1.4.2changes].

[1.4.2changes]: https://github.com/ameshkov/gocurl/compare/v1.4.1...v1.4.2

[1.4.2]: https://github.com/ameshkov/gocurl/releases/tag/v1.4.2

## [1.4.1] - 2024-02-07

### Added

* Added support for `--ipv4` and `--ipv6` arguments.  ([#25][#25])
* Added [a Docker image][dockerimage] for `gocurl`.

### Fixed

* Fixed a bug introduced in v1.4.0 with `gocurl` not printing the response body
  when the protocol is HTTP/2 or HTTP/3.  ([#26][#26])
* Fixed unnecessary warning `connection doesn't allow setting of receive buffer
  size` when HTTP/3 is used.

[See changes][1.4.1changes].

[1.4.1changes]: https://github.com/ameshkov/gocurl/compare/v1.4.0...v1.4.1

[1.4.1]: https://github.com/ameshkov/gocurl/releases/tag/v1.4.1

[dockerimage]: https://github.com/ameshkov/gocurl/pkgs/container/gocurl

[#25]: https://github.com/ameshkov/gocurl/issues/25

[#26]: https://github.com/ameshkov/gocurl/issues/26

## [1.4.0] - 2024-02-04

### Added

* Added initial WebSocket support. `gocurl` now supports `ws://` and `wss://`
  URLs. `-d` can be used to specify initial data to send.  ([#17][#17])

[See changes][1.4.0changes].

[1.4.0changes]: https://github.com/ameshkov/gocurl/compare/v1.3.0...v1.4.0

[1.4.0]: https://github.com/ameshkov/gocurl/releases/tag/v1.4.0

[#17]: https://github.com/ameshkov/gocurl/issues/17

## [1.3.0] - 2023-09-23

### Added

* Added an option to use Post-Quantum secure algorithms for establishing TLS
  connections. This option is hidden under a new `--experiment` flag that is
  described in README.md. ([#15][#15])

### Fixed

* Fixed an issue with `--http2` not being able to work together with `--ech`. In
  addition to that there's also one more change: `gocurl` by default will send
  both `h2` and `http/1.1` in TLS ALPN extension and use the protocol selected
  by the server. ([#14][#14])

[See changes][1.3.0changes].

[1.3.0changes]: https://github.com/ameshkov/gocurl/compare/v1.2.0...v1.3.0

[1.3.0]: https://github.com/ameshkov/gocurl/releases/tag/v1.3.0

[#14]: https://github.com/ameshkov/gocurl/issues/14

[#15]: https://github.com/ameshkov/gocurl/issues/15

## [1.2.0] - 2023-09-22

### Added

* Added `--dns-servers` command-line argument support. Besides regular DNS,
  `gocurl` also supports encrypted DNS, see examples in README.md to learn
  more. ([#6][#6])

### Fixed

* TLS state is now printed to the output for ECH-enabled connections. In
  addition to that, much more TLS-related information is printed to the output
  including information about TLS certificates. ([#8][#8])

[See changes][1.2.0changes].

[1.2.0changes]: https://github.com/ameshkov/gocurl/compare/v1.1.0...v1.2.0

[1.2.0]: https://github.com/ameshkov/gocurl/releases/tag/v1.2.0

[#6]: https://github.com/ameshkov/gocurl/issues/6

[#8]: https://github.com/ameshkov/gocurl/issues/8

## [1.1.0] - 2023-09-21

### Added

* `gocurl` now supports Encrypted Client Hello. Added `--ech` and `--echconfig`
  command-line arguments, see examples in README.md to learn more. ([#3][#3])
* Added `--resolve` command-line argument support. It works similarly to the one
  in `curl` with one important difference: `gocurl` ignores `port` there and
  simply returns specified IP addresses for the host. ([#5][#5])

[See changes][1.1.0changes].

[1.1.0changes]: https://github.com/ameshkov/gocurl/compare/v1.0.6...v1.1.0

[1.1.0]: https://github.com/ameshkov/gocurl/releases/tag/v1.1.0

[#3]: https://github.com/ameshkov/gocurl/issues/3

[#5]: https://github.com/ameshkov/gocurl/issues/5

## [1.0.6] - 2023-09-17

### Added

* `--connect-to` and `--proxy` now also support HTTP/3. ([#1][#1])

[See changes][1.0.6changes].

[1.0.6changes]: https://github.com/ameshkov/gocurl/compare/v1.0.5...v1.0.6

[1.0.6]: https://github.com/ameshkov/gocurl/releases/tag/v1.0.6

[#1]: https://github.com/ameshkov/gocurl/issues/1

## [1.0.5] - 2023-09-15

### Added

* Added the changelog.

### Changed

* Added more debug logging to all request stages.

### Fixed

* Fixed the way `--connect-to` works when a proxy is specified. Before this
  change gocurl could only redirect the proxy connection, but not the one that
  goes through proxy. This behavior is now fixed.

[See changes][1.0.5changes].

[1.0.5changes]: https://github.com/ameshkov/gocurl/compare/v1.0.4...v1.0.5

[1.0.5]: https://github.com/ameshkov/gocurl/releases/tag/v1.0.5

## [1.0.4] - 2023-09-12

### Fixed

* Fixed the issue with the output being written not to stdout by default.

[See changes][1.0.4changes].

[1.0.4changes]: https://github.com/ameshkov/gocurl/compare/v1.0.3...v1.0.4

[1.0.4]: https://github.com/ameshkov/gocurl/releases/tag/v1.0.4

## [1.0.3] - 2023-09-12

### Fixed

* Minor improvements.

[See changes][1.0.3changes].

[1.0.3changes]: https://github.com/ameshkov/gocurl/compare/v1.0.2...v1.0.3

[1.0.3]: https://github.com/ameshkov/gocurl/releases/tag/v1.0.3

## [1.0.2] - 2023-09-12

### Added

* Automate the release process.

[See changes][1.0.2changes].

[1.0.2changes]: https://github.com/ameshkov/gocurl/compare/v1.0.1...v1.0.2

[1.0.2]: https://github.com/ameshkov/gocurl/releases/tag/v1.0.2

## [1.0.1] - 2023-09-12

### Fixed

* Logging improvements.

[See changes][1.0.1changes].

[1.0.1changes]: https://github.com/ameshkov/gocurl/compare/v1.0.0...v1.0.1

[1.0.1]: https://github.com/ameshkov/gocurl/releases/tag/v1.0.1

## [1.0.0] - 2023-09-12

### Added

* The first version with base functionality.

[1.0.0]: https://github.com/ameshkov/gocurl/releases/tag/v1.0.0