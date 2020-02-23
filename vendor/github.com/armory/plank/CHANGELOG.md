# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added

### Changed

### Fixed

### Removed

## [1.3.0] - 2020-02-25
### Added
- `plank.Put` and `plank.Post` methods have learned how to return response
paylods from 4xx and 5xx responses in the `plank.FailedResponse` struct.
  - This allows the caller to unmarshall the response payload into whatever
  struct makes sense for the context.

[Unreleased]: https://github.com/armory/plank/compare/v1.3.0...HEAD
[1.3.0]: https://github.com/armory/plank/compare/v1.2.1...v1.3.0
