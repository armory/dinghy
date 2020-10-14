# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Fixed

### Removed

## [3.4.3] 2020-1014

### Added
- Created method `UpdatePermissions` that is used by `UpdateApplication`

### Fixed
- Fixed a bug in which application permissions were not updating in `UpdateApplication`

### Removed

## [3.4.2] 2020-06-18

### Added

### Fixed
- Fixed nil check for `ReadPermissable` interface

### Removed

## [3.4.1] 2020-06-10
### Added

### Changed
- Added more descriptive error message to `plank.ValidateAppNotification`

### Fixed

### Removed

## [3.4.0] 2020-06-10
### Added
- Added support for Application Notifications with object `plank.NotificationsType` 
  - `plank.UpdateApplicationNotifications` to create, update app notifications 
  - `plank.GetApplicationNotifications` to read app notifications
  - `plank.FillAppNotificationFields` to fill default values for them
  - `plank.ValidateAppNotification` to validate struct

### Changed

### Fixed

### Removed
## [3.3.0] 2020-06-08

### Added
- Improved Fiat client interface
    - Changed ReadPermissable methods to getters
    - Added support for passing in plank client options

## [3.2.0] 2020-06-04
### Added

- FiatPermissionsEvaluator for determining authorization based on User roles with Fiat

```go
// example usage
objs := []Foo{
	{name: "foo1", permisisons: []string{"engineering", "core-eng"}},
	{name: "foo2", permisisons: []string{"ico-team", "spin-team"}},
	{name: "foo3", permisisons: []string{"engineering"}},
}

evaluator := plank.NewFiatPermissionEvaluator(plank.WithOrMode(true))

user := "ethanfrogers"

for _, o := range objs {
	allowed, err := evaluator.HasReadPermission(user, o)
	if err != nil {
		fmt.Printf("error encountered: %s", err.Error())
		os.Exit(1)
	}
	if allowed {
		fmt.Printf("obj %s allowed\n", o.Name())
	}
}
```

## [3.1.0] 2020-05-25
### Added

### Changed
- `plank.ValidateRefIds` for `plank.Pipeline` will return a warning instead of an error when refId is not found for a stage.

### Fixed

### Removed


## [3.0.0] 2020-05-19
### Added

### Changed
- (breaking change) `plank.Client` change the defaults URLs from `armory-orca` etc to `localhost` as it's more common to do `kubectl port-forwards` for each service locally.
- (breaking change) `plank.Client` has the cient option `WithURLs` renamed to `WithOverrideAllURLs` to make it more obvious what this function is doing

### Fixed

### Removed

## [2.1.0] - 2020-05-18
### Added

- `plank.ValidateRefIds` method was created to validate refIds in stages for `plank.Pipeline`

### Changed

### Fixed

### Removed

## [2.0.0] - 2020-03-19
### Added

- `plank.Client` has learned how to _update_ and _delete_ Application objects
- `plank.Client` has learned how to make `PATCH` requests to Spinnaker services.

### Changed

- `plank.Application` objects have learned how to make the following fields
optional: `DataSourceType`, `PermissionsType`
  - If you weren't explicitly setting these fields then they will be ommitted
  from the request made to Spinnaker.  If you want to preserve existing behavior
  you can initialize these fields like so:

```go
app := plank.Application{}

// To set an empty datasourcetype
app.DataSources = &plank.DataSourceType{}

// To set an empty permissions block
app.Permissions = &plank.PermissionsType{}
```

- `plank.Pipeline` objects have learned how to make locking pipeline UI edits
optional.
  - Previous behavior required you to _opt out_ of locking pipelines changes
  in the UI. If you'd like to preserve this behavior make sure your objects
  are calling `Lock()` before passing them to `plank.Client`:

```go
p := plank.Pipeline{}
p.Lock() // UI edits will now be disallowed

// ... create pipeline via plank.Client
```

### Fixed

### Removed


## [1.3.0] - 2020-02-25
### Added
- `plank.Put` and `plank.Post` methods have learned how to return response
paylods from 4xx and 5xx responses in the `plank.FailedResponse` struct.
  - This allows the caller to unmarshall the response payload into whatever
  struct makes sense for the context.

[Unreleased]: https://github.com/armory/plank/compare/v1.3.0...HEAD
[3.4.3]: https://github.com/armory/plank/compare/v3.4.2...v3.4.3
[3.4.2]: https://github.com/armory/plank/compare/v3.4.1...v3.4.2
[3.4.1]: https://github.com/armory/plank/compare/v3.4.0...v3.4.1
[3.4.0]: https://github.com/armory/plank/compare/v3.3.0...v3.4.0
[3.3.0]: https://github.com/armory/plank/compare/v3.2.0...v3.3.0
[3.2.0]: https://github.com/armory/plank/compare/v3.1.0...v3.2.0
[3.1.0]: https://github.com/armory/plank/compare/v3.0.0...v3.1.0
[3.0.0]: https://github.com/armory/plank/compare/v2.1.0...v3.0.0
[2.1.0]: https://github.com/armory/plank/compare/v2.0.0...v2.1.0
[2.0.0]: https://github.com/armory/plank/compare/v1.3.0...v2.0.0
[1.3.0]: https://github.com/armory/plank/compare/v1.2.1...v1.3.0
