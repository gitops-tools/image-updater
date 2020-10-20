# image-updater ![Go](https://github.com/gitops-tools/image-updater/workflows/Go/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/gitops-tools/image-updater)](https://goreportcard.com/report/github.com/gitops-tools/image-updater)

This is a small tool and service for updating YAML files with image references,
to simplify continuous deployment pipelines.

It updates a YAML file in a Git repository, and optionally opens a Pull Request.

## Command-line tool

```shell
$ ./image-updater --help
Update YAML files in a Git service, with optional automated Pull Requests

Usage:
  image-updater [command]

Available Commands:
  help        Help about any command
  http        update repositories in response to image hooks
  pubsub      update repositories in response to gcr pubsub events
  update      update a repository configuration

Flags:
  -h, --help   help for image-updater

Use "image-updater [command] --help" for more information about a command.
```

There are three sub-commands, `http`, `pubsub` and `update`.

`http` provides a Webhook service, `pubsub` subscribes to pubsub events and `update` will perform the same
functionality from the command-line.

## Update tool

This requires a `AUTH_TOKEN` environment variable with a token.

```shell
$ ./image-updater update --file-path service-a/deployment.yaml --image-repo quay.io/myorg/my-image --source-repo mysource/my-repo --new-image-url quay.io/myorg/my-image:v1.1.0 --update-key spec.template.spec.containers.0.image
```

This would update a file `service-a/deployment.yaml` in a GitHub repo at `mysource/my-repo`, changing the `spec.template.spec.containers.0.image` key in the file to `quay.io/myorg/my-image:v1.1.0`, the PR will indicate that this is an update from `quay.io/myorg/my-image`.

If you need to access a private GitLab or GitHub installation, you can provide
the `--api-endpoint` e.g.

```shell
$ ./image-updater update --file-path service-a/deployment.yaml --image-repo quay.io/myorg/my-image --source-repo mysource/my-repo --new-image-url quay.io/myorg/my-image:v1.1.0 --update-key spec.template.spec.containers.0.image
```

For the HTTP service, you will likely need to adapt the deployment.

You can also opt to allow for insecure TLS access with `--insecure`.

## Webhook Service

This is a micro-service for updating Git Repos when a hook is received indicating that a new image has been pushed from an image repository.

This currently supports receiving hooks from Docker and Quay.io.

### WARNING

Neither Docker Hub nor Quay.io provide a way for receivers to authenticate Webhooks, which makes this insecure, a malicious user could trigger the creation of pull requests in your git hosting service.

Please understand the risks of using this component.

## Pubsub Service
Similarly to the Webhook service, the pubsub services allows to update Git Repos when a pubsub Event is received. 

This currently supports Events from [Google Cloud Registry](https://cloud.google.com/container-registry/docs/configuring-notifications).

It requires two arguments `--project-id` and `--subscription-name`. See [below](#google-container-registry-setup) for more details on how to setup the subscription.

## Configuration

Both the Webhook and Pubsub service uses a really simple configuration:

```yaml
repositories:
  - name: testing/repo-image
    sourceRepo: my-org/my-project
    sourceBranch: main
    filePath: service-a/deployment.yaml
    updateKey: spec.template.spec.containers.0.image
    branchGenerateName: repo-imager-
    tagMatch: "^main-.*"
```

This is a single repository configuration, Repo Push notifications from the
image `testing/repo-image`, will trigger an update in the repo
`my-org/my-project`.

The change will be based off the `main` branch, and updating the file
`service-a/deployment.yaml`.

Within that file, the `spec.template.spec.containers.0.image` field will be replaced
with the incoming image.

A new branch will be created based on the `branchGenerateName` field, which
would look something like `repo-imager-kXzdf`.

The presence of the `tagMatch` field means that it should only apply the update,
if the tag being changed matches this regular expression, in this case, tags
like "main-c1f79ab" would match, but "test-pr-branch-c1f79ab" would not.

### Updating the sourceBranch directly

If no value is provided for `branchGenerateName`, then the `sourceBranch` will
be updated directly, this means that if you use `main`, then the token must
have access to push a change directly to `main`.

### Creating the configuration

The tool reads a YAML definition, which in the provided `Deployment` is mounted
in from a `ConfigMap`.

```shell
$ kubectl create configmap image-updater-config --from-file=config.yaml
```

The default deployment requires a secret to expose the `GITHUB_TOKEN` to the
service.


```shell
$ export GITHUB_TOKEN=<insert github token>
$ kubectl create secret generic image-updater-secret --from-literal=token=$GITHUB_TOKEN
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

## Tekton

A Tekton task is provided in [./tekton](./tekton) which allows you to apply
updates to repos from a Tekton pipeline run.


## Google Container registry setup
```bash
gcloud pubsub topics create gcr
gcloud pubsub subscriptions create gcr-image-updater --topic projects/$GOOGLE_PROJECT/topics/gcr

gcloud iam service-accounts create 
gcloud iam service-accounts keys create credentials.json \
  --iam-account $SA_NAME@$GOOGLE_PROJECT.iam.gserviceaccount.com

gcloud pubsub subscriptions add-iam-policy-binding gcr-image-updater \
--member=serviceAccount:$SA_NAME@$GOOGLE_PROJECT.iam.gserviceaccount.com --role=roles/pubsub.subscriber
```

You then need to set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to the path the generated  `credentials.json` file.

## Building

A `Dockerfile` is provided for building a container, but otherwise:

```shell
$ go build ./cmd/image-updater
```

## Testing

```shell
$ go test ./...
```
