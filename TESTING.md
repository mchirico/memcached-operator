

```bash
kubectl apply -f config/samples/cache_v1alpha1_memcached.yaml

kubectl get deployment
kubectl get pods

kubectl get memcached/memcached-sample -o yaml


kubectl patch memcached memcached-sample -p '{"spec":{"size": 5}}' --type=merge


```
