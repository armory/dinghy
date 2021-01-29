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

### Local Development

You will need a [golang toolchain] and [make] to work on this project.

You should complete and add the file located in `example/dinghy.yml` to `/opt/spinnaker/config/dinghy.yml` since 
this is the file that dinghy search for configuration.

#### Building & Testing

You can run the `make build` and `make test` targets to build and test the
project.  You will need Redis running (either locally or in your Spinnaker
cluster), as well as Front50 and Orca.

If you have an existing Spinnaker cluster, you can port-forward to your local
machine like so:

```shell
kubectl -n spinnaker port-forward svc/spin-redis   6379
kubectl -n spinnaker port-forward svc/spin-front50 8080
kubectl -n spinnaker port-forward svc/spin-orca    8083
kubectl -n spinnaker port-forward svc/spin-fiat    7003
kubectl -n spinnaker port-forward svc/spin-echo    8089
```



#### Sample Request

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

[golang toolchain]: https://golang.org/doc/install
[make]: https://www.gnu.org/software/make/
