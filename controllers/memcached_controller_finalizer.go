package controllers

import (
	"context"
	cachev1alpha1 "github.com/mchirico/memcached-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// memcached := &cachev1alpha1.Memcached{}
func (r *MemcachedReconciler) addFinalizer(ctx context.Context, customResource *cachev1alpha1.Memcached) error {
	controllerutil.AddFinalizer(customResource, cachev1alpha1.MemcachedFinalizerName)

	return r.Update(ctx, customResource)
}

// handleFinalizer returns a bool and an error. If error is set then the attempt failed, otherwise boolean indicates whether it completed
func (r *MemcachedReconciler) handleFinalizer(ctx context.Context, customResource *cachev1alpha1.Memcached) error {
	if !customResource.HasFinalizer(cachev1alpha1.MemcachedFinalizerName) {
		return nil
	}

	// remove finalizer from resource to allow object deletion
	customResource.RemoveFinalizer(cachev1alpha1.MemcachedFinalizerName)
	return r.Update(ctx, customResource)

}
