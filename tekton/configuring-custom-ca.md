# Configuring a Custom CA Chain

You may need image-updater to interact with Git servers exposed using TLS which do not have a certificate signed by a well-known Certificate Authority.

In such cases you can use the `--insecure` flag, or even better, load your custom CA Chain into the container which runs the task.

## Get the CA Chain

First, we need to get the CA Chain so we can configure the Tekton task to make use of it. Usually, the CA Chain can be obtained from the vendor/team who signed
the certificate being used by your Git Server. You can get it using a web browser as well, the process of getting the CA Chain is out of the scope of this document.

## Configuring a ConfigMap with you CA Chain

Once we have our CA Chain file with PEM format, we can go ahead and create a ConfigMap in order to store the CA Chain in the Kubernetes cluster where Tekton is running.

The ConfigMap must be created in the same namespace where the Tekton task for `image-updater` is defined.

~~~sh
kubectl -n <tekton-task-namespace> create configmap custom-ca-chain --from-file=ca-bundle.crt=</path/to/your/cachain/in/pem/format>
~~~

## Configuring a Volume on the Tekton Task

We have our Custom CA Chain loaded into Kubernetes as a ConfigMap, now we need to configure the `image-updater` Tekton task to make use of it.

You can find the definition of the `image-updater` Tekton Task [here](https://github.com/gitops-tools/image-updater/blob/main/tekton/image-updater.yaml). We are using this Task as reference.

You need to add the `volumeMounts` and `volumes` sections to the Tekton Task spec.

~~~yaml
<OUTPUT_OMMITED>
  steps:
    - name: update-image
      image: bigkevmcd/image-updater:latest
      args:
        - "update"
        - "--driver=$(params.driver)"
        - "--file-path=$(params.file-path)"
        - "--image-repo=$(params.image-repo)"
        - "--new-image-url=$(params.new-image-url)"
        - "--source-branch=$(params.source-branch)"
        - "--source-repo=$(params.source-repo)"
        - "--update-key=$(params.update-key)"
        - "--branch-generate-name=$(params.branch-generate-name)"
        - "--api-endpoint=$(params.api-endpoint)"
        - "--insecure=$(params.insecure)"
      env:
        - name: AUTH_TOKEN
          valueFrom:
            secretKeyRef:
              name: image-updater-secret
              key: token
      volumeMounts:
      - mountPath: /etc/pki/tls/certs/
        name: custom-ca-chain
  volumes:
  - configMap:
      name: custom-ca-chain
    name: custom-ca-chain
~~~

At this point you will be able to connect to your Git server without the need of using `--insecure`.
