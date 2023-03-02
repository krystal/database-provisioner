# Kubernetes External Database Provisioner

This is a Kubernetes operator that provisions databases on external database services. Right now it supports the following backends and more will be added.

- MySQL / MariaDB

## Getting started

1. Grab a manifest.yaml from the [latest release](https://github.com/krystal/database-provisioner/releases).
2. Apply to the manifest to your cluster using `kubectl apply -f`.
3. Add appropriate

## Example

Apply a default `MySQLServer` CR to your cluster with details for an external service.

```yaml
apiVersion: databases.k8s.k.io/v1
kind: MySQLServer
metadata:
  name: default
spec:
  host: 185.22.208.10
  port: 3306
  username: root
  password: 5up3r53cr3t
```

Apply a `MySQLDatabase` CR to your cluster to create a database.

```yaml
apiVersion: databases.k8s.k.io/v1
kind: MySQLDatabase
metadata:
  name: my-database
  namespace: some-namespace
spec:
  serverName: default
  connectionDetailsSecretName: db-connection
```

- This will use the `default` server
- This will create a secret called `db-connection` containing details for the database connection. This secret will be created in the same namespace as the `MySQLDatabase` CR. It will have values for `host`, `username`, `password`, and `databaseName`.

## Developing

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster. **Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster

1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/database-provisioner:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/database-provisioner:tag
```

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller

UnDeploy the controller from the cluster:

```sh
make undeploy
```

### Test It Out

1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)
