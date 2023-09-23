# gocurl changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][keepachangelog], and this project
adheres to [Semantic Versioning][semver].

[keepachangelog]: https://keepachangelog.com/en/1.0.0/

[semver]: https://semver.org/spec/v2.0.0.html

## [Unreleased]

[unreleased]: https://github.com/ameshkov/gocurl/compare/v1.3.0...HEAD

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

[#14]: https://github.com/ameshkov/gocurl/issues/14

[#15]: https://github.com/ameshkov/gocurl/issues/15

[1.3.0changes]: https://github.com/ameshkov/gocurl/compare/v1.2.0...v1.3.0

[1.3.0]: https://github.com/ameshkov/gocurl/releases/tag/v1.3.0

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

[#6]: https://github.com/ameshkov/gocurl/issues/6

[#8]: https://github.com/ameshkov/gocurl/issues/8

[1.2.0changes]: https://github.com/ameshkov/gocurl/compare/v1.1.0...v1.2.0

[1.2.0]: https://github.com/ameshkov/gocurl/releases/tag/v1.2.0

## [1.1.0] - 2023-09-21

### Added

* `gocurl` now supports Encrypted Client Hello. Added `--ech` and `--echconfig`
  command-line arguments, see examples in README.md to learn more. ([#3][#3])
* Added `--resolve` command-line argument support. It works similarly to the one
  in `curl` with one important difference: `gocurl` ignores `port` there and
  simply returns specified IP addresses for the host. ([#5][#5])

[See changes][1.1.0changes].

[#3]: https://github.com/ameshkov/gocurl/issues/3

[#5]: https://github.com/ameshkov/gocurl/issues/5

[1.1.0changes]: https://github.com/ameshkov/gocurl/compare/v1.0.6...v1.1.0

[1.1.0]: https://github.com/ameshkov/gocurl/releases/tag/v1.1.0

## [1.0.6] - 2023-09-17

### Added

* `--connect-to` and `--proxy` now also support HTTP/3. ([#1][#1])

[See changes][1.0.6changes].

[#1]: https://github.com/ameshkov/gocurl/issues/1

[1.0.6changes]: https://github.com/ameshkov/gocurl/compare/v1.0.5...v1.0.6

[1.0.6]: https://github.com/ameshkov/gocurl/releases/tag/v1.0.6

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