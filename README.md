# image-hooks ![Go](https://github.com/bigkevmcd/image-hooks/workflows/Go/badge.svg)

This is a small tool and service for updating YAML files with image references,
to simplify continuous deployment pipelines.

You get to choose, orchestration or choreography.

## Command-line tool

```shell
$ ./image-hooks update --file-path service-a/person.yaml --image-repo quay.io/myorg/my-image --source-repo mysource/my-repo --new-image-url quay.io/myorg/my-image:v1.1.0 --update-key person.name
```

## Webhook Service

A micro-service for updating Git Repos when a hook is received indicating that a
new image has been pushed from an image repository.

This supports receiving hooks from Docker and Quay.io.

## WARNING

Neither Docker Hub nor Quay.io provide a way for receivers to authenticate Webhooks, which makes this insecure, a malicious user could trigger the creation of pull requests in your git hosting service.

Please understand the risks of using this component.

## Configuration

This service uses a really simple configuration:

```yaml
repositories:
  - name: testing/repo-image
    sourceRepo: my-org/my-project
    sourceBranch: master
    filePath: service-a/deployment.yaml
    updateKey: spec.template.spec.containers.0.image
    branchGenerateName: repo-imager-
```

This is a single repository configuration, Repo Push notifications from the
image `testing/repo-image`, will trigger an update in the repo
`my-org/my-project`.

The change will be based off the `master` branch, and updating the file
`service-a/deployment.yaml`.

Within that file, the `spec.template.spec.containers.0.image` field will be replaced
with the incoming image.

A new branch will be created based on the `branchGenerateName` field, which
would look something like `repo-imager-kXzdf`.

### Updating the sourceBranch directly

If no value is provided for `branchGenerateName`, then the `sourceBranch` will
be updated directly, this means that if you use `master`, then the token must
have access to push a change directly to master.

### Creating the configuration

The tool reads a YAML definition, which in the provided Deployment is mounted
in from a `ConfigMap`.

```shell
$ kubectl create configmap image-hooks-config --from-file=config.yaml
```

```shell
$ export GITHUB_TOKEN=<insert github token>
$ kubectl create secret generic image-hooks-secret --from-literal=token=$GITHUB_TOKEN
```

## Deployment

A Kubernetes `Deployment` is provided in [./deploy/deployment.yaml](./deploy/deployment.yaml).

The service is **not** dependent on being executed within a Kubernetes cluster.

## Choosing a hook parser

By default, this accepts hooks from Docker hub but the deployment can easily be
changed to support Quay.io.

The `--parser` command-line option chooses which of the supported (Quay, Docker)
hook formats to parse.

## Exposing the Handler

The Service exposes a Hook handler at `/` on port 8080 that handles the
configured hook type.

## Building

A `Dockerfile` is provided for building a container, but otherwise:

```shell
$ go build ./cmd/quay-hooks
```

## Testing

```shell
$ go build ./...
```
