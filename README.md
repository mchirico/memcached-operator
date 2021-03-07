# memcached-operator

The goal of this repo is to create a simple operator
that closely matches [operator sdk tutoral](https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/)
with tests.

## Note

The memcached-operator does not include Ginkgo tests, or live testing
with KinD. This repo is setup for easy experimentation, and sharing
of ideas. 

See added tests: 

[memcached_types_test.go](https://github.nike.com/npe/memcached-operator/blob/04a4ade1715791501c67fef919d4904df012c929/api/v1alpha1/memcached_types_test.go#L15)

[controllers/suite_test.go](https://github.nike.com/npe/memcached-operator/blob/04a4ade1715791501c67fef919d4904df012c929/controllers/suite_test.go#L76)

## Reference

Reference the [azure-databricks-operator](https://github.com/microsoft/azure-databricks-operator) and all the
advanced testing.
# memcached-operator
