

```
make docker-build IMG=ghcr.io/haunshila/kafi:v1.0.0
make deploy IMG=ghcr.io/haunshila/kafi:v1.0.0
```

* Kind Cluster
  ```bash
    kind create cluster --name kafi --config kind-config.yaml
  ```
* Create Ceph storage class
```bash
    kubectl apply -f ceph-storageclass.yaml  
```

* To allow a Kubernetes cluster to pull images from the GitHub Container Registry (ghcr.io)
```bash
  kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=haunshila \
  --docker-password=SECRET \
  --docker-email=haunshila@gmail.com \
  -n kafi-system
```

* Deploy Operator
```bash
make generate
make manifests
make install
make docker-build IMG=ghcr.io/haunshila/kafi:v1.0.0
make deploy IMG=ghcr.io/haunshila/kafi:v1.0.0
```