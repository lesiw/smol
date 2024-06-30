# short: a shortened URL host for k8s

## Local development

### Setup minikube

``` sh
minikube start
minikube addons enable ingress
echo "$(minikube ip) smol.lan" | sudo tee -a /etc/hosts
```

### Build images for minikube

``` sh
minikube image build -t lesiw/smol .
minikube image build -t lesiw/smol:db -f Dockerfile.pg .
```

## Deploy

``` sh
go run internal/secrets/generate.go -n "${DOMAIN:-smol.lan}"
helm install smol ./chart \
    --set domain="${DOMAIN:-smol.lan}" \
    --set dbsize=1Gi \
    --set memory=32Mi \
    --set dbmemory=64Mi \
    --set image=lesiw/smol \
    --set dbimage=lesiw/smol:db
```

## Manage URLs

Exec into the db-0 pod. For example, if you have deployed to the default domain
and namespace:

``` sh
kubectl exec -ti -n smol-lan db-0 -- /bin/bash
```

Run the `smol` utility to add a URL.

``` shellsession
$ smol https://example.com
smol.lan/JgHXr6
$ smol -a example https://example.com
smol.lan/example
```
