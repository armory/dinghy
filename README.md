[![Coverage Status](https://coveralls.io/repos/github/armory/dinghy/badge.svg?branch=testcover)](https://coveralls.io/github/armory/dinghy?branch=testcover)

# dinghy

Dinghy allows you to create and maintain Spinnaker pipeline templates in source
control.

Read more in our
[documentation](https://docs.armory.io/docs/armory-admin/dinghy-enable/).

### How It Works

There are two primitives:
- Stage/Task templates: These are all kept in a single GitHub repo. They are
  json files with replacable values in them.
- Pipeline definitions: These define a pipeline for an application. You can
  compose stage/task templates to make a full definition.

How it works:
- GitHub webhooks are sent off when either the templates or the definitions are
  modified.
- Templates should be versioned by hash when they are used.
- Dinghy will keep a dependency graph of downstream templates. When a
  dependency is modified, the pipeline definition will be rebuilt and re-posted
  to Spinnaker. (sound familiar? haha)

### Testing Manually

```shell
curl -X POST \
  -H "Content-Type: application/json" \
  -d "@example/github_payload.json" \
  http://localhost:8089/webhooks/git/github
```

(The github_payload.json file in the example directory is a minimal set for
testing the git webhook, as an example)

Dinghy is also embedded in the [arm cli](https://github.com/armory-io/arm) tool
for local validation of pipelines.
