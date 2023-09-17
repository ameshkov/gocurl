# gocurl changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][keepachangelog], and this project
adheres to [Semantic Versioning][semver].

[keepachangelog]: https://keepachangelog.com/en/1.0.0/

[semver]: https://semver.org/spec/v2.0.0.html

## [Unreleased]

[unreleased]: https://github.com/ameshkov/gocurl/compare/v1.0.5...HEAD

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