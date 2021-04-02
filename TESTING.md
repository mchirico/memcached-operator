
```bash
k apply -f config/samples/cache_v1alpha1_memcached.yaml

k get memcacheds
k get memcacheds  memcached-sample -o yaml

k delete memcacheds  memcached-sample

kubectl patch memcacheds memcached-sample -p '{"spec":{"size": 5}}' --type=merge

kubectl patch memcacheds memcached-sample -p '{"metadata":{"finalizers":[]}}' --type=merge


# Initial Loading
kubectl apply -f config/samples/environment_v1alpha1_environment.yaml


kubectl patch memcached memcached-sample -p '{"spec":{"size": 5}}' --type=merge


# Making changes
kubectl patch environments environment-sample -p '{"spec":{"namespaces": ["a","b","c","d","e"]}}' --type=merge
k get environments/environment-sample -o yaml

k logs crd/environments.npe.operators.npe.nike.com


# Use the following new way
k get environment.npe.nike.com/environment-sample -o yaml

kubectl patch environment.npe.nike.com environment-sample -p '{"spec":{"namespaces": ["bozo"]}}' --type=merge


k get environment.npe.nike.com environment-sample -o yaml


golangci-lint run


```
