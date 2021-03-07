package controllers

import (
	"context"
	"github.com/mchirico/memcached-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

// TODO-mmc:  Need to finish setup of tests ...
//  Need to write stubs.  Currently this will call Reconcile
var _ = Describe("Namespecial controller", func() {

	// Define utility constants for object names and testing timeouts/intervals.
	const (
		kind      = "Memcached"
		namespace = "default"
		name      = "memcached-sample"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When Deleting Namespace `stuff`", func() {
		It("Should recreate ", func() {
			By("By creating a new namespace")

			ctx := context.Background()

			key := types.NamespacedName{
				Name:      name + "-abcd",
				Namespace: namespace,
			}

			cr := &v1alpha1.Memcached{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
			}

			cr.Kind = kind
			cr.Namespace = namespace
			cr.Name = name
			cr.Spec = v1alpha1.MemcachedSpec{}
			cr.Spec.Size = 3

			// We'll need to retry getting this, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Create(ctx, cr)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			cr.Spec.Size = 2
			Eventually(func() bool {
				err := k8sClient.Update(ctx, cr)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			//By("Expecting size to be 2")
			//Eventually(func() int32 {
			//	f := &v1alpha1.Memcached{}
			//	_ = k8sClient.Get(context.Background(), key, f)
			//	return f.Spec.Size
			//}, timeout, interval).Should(Equal(2))

			Eventually(func() bool {
				err := k8sClient.Delete(ctx, cr)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

		})
	})
})
