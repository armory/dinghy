# dinghy

_A little boat to take you to the big boat_ - Miker

Note: This is a prototype.

## Developing

### Docker

You can make the docker image locally.

```bash
make docker
```

And push it to Artifactory for use with Halyard

```bash
make docker-push
``` 

You can of course combine these steps into a single action

```bash
make docker docker-push
```

## Yo, but what do it do?

Users should be able to specify a pipeline in code in their GitHub repo. Dinghy should keep the pipeline in Spinnaker in sync with what is in the GitHub repo. Also, users should be able to make a pipeline by composing other pipelines, stages, or tasks and templating certain values.

### Deets

There are two primitives:
- Stage/Task templates: These are all kept in a single GitHub repo. They are json files with replacable values in them.
- Pipeline definitions: These define a pipeline for an application. You can compose stage/task templates to make a full definition.

How it might work:
- GitHub webhooks are sent off when either the templates or the definitions are modified.
- Templates should be versioned by hash when they are used.
- Dinghy will keep a dependency graph of downstream templates. When a dependency is modified, the pipeline definition will be rebuilt and re-posted to Spinnaker. (sound familiar? haha)

<!-- made using ./bin/makeDiagrams.sh -->
![](diagrams/workflow.mmd.svg)

### Updating oss dinghy version

This dinghy implementation (armory-io) has dependency on oss implementation. Once new code is added to oss, armory-io dinghy needs to be updated with new library version.
Update is as simple as changing current oss dinghy version with new one in two files:
- vendor/modules.txt
- go.mod

Of course, you want to keep the version in those two file the same to prevent from inconsistent behaviour and build issues.
Dinghy's version has format like that: v0.0.0-X-Y where `X` is a date and time of commit and `Y` is shortened commit hash.
An example of that may be: `v0.0.0-20221025163127-2d465e0cea94`

Once updated oss dinghy version, run:

```bash
go mod vendor
go mod tidy
```

It should pull specified version from oss repository.

Don't forget about interfaces implementation. If interface definition is in oss dinghy, but there's additional implementation in armory-io dinghy, add missing methods.

### Backporting oss features

In order to backport new feature to older release, you need to:
1. Create new branch from destination release branch: git checkout -b "bp/A.BC.D/ossprX", where `A.BC.D` is a destination release, `X` is a number of PR in oss repository
2. Update oss dinghy version - [Update](#updating-oss-dinghy-version)
3. Merge your branch `bp/A.BC.D/ossprX` to release branch `release-A.BC.D`
