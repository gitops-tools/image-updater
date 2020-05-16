# image-hooks ![Go](https://github.com/bigkevmcd/image-hooks/workflows/Go/badge.svg)

A micro-service for updating Git Repos when a Quay image hook is received.

## WARNING

Quay.io provides no way for receivers to authenticate Webhooks, which makes this insecure, a malicious user could trigger the creation of pull requests in your git hosting service.

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

### Creating the configuration

The tool reads a YAML definition, which by default is mounted in from a
`ConfigMap`.

```shell
$ kubectl create configmap image-hooks-config --from-file=config.yaml
```

```shell
$ export GITHUB_TOKEN=<insert github token>
$ kubectl create secret generic image-hooks-secret --from-literal=token=$GITHUB_TOKEN
```

## Deployment

A Kubernetes `Deployment` is provided in [./deploy/deployment.yaml](./deploy/deployment.yaml).

The service is not dependent on being executed within a Kubernetes cluster.

## Exposing the Handler

The Service exposes a Hook handler at `/` on port 8080 that handles the hooks.

## Building

A `Dockerfile` is provided for building a container, but otherwise:

```shell
$ go build ./cmd/quay-hooks
```

## Testing

```shell
$ go build ./...
```
