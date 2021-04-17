package controllers

import (
	"context"
	"fmt"
	"github.com/mchirico/memcached-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

		// TODO mmc: set this back to 10 when done testing
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
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Note that your Memorycached’s GroupVersionKind is required to set up this owner reference.
			//kind := reflect.TypeOf(cachev1alpha1.Memcached{}).Name()
			//gvk := cachev1alpha1.GroupVersion.WithKind(kind)
			//controllerRef := metav1.NewControllerRef(cr, gvk)
			//cr.SetOwnerReferences([]metav1.OwnerReference{*controllerRef})

			//MergePatchType
			//PatchType

			patch := []byte(`{"spec":{"size": 5}}`)

			cr.Status.Nodes = []string{"one"}
			cr.Status.Msg = []string{"one"}

			err := k8sClient.Get(ctx, client.ObjectKey{
				Namespace: cr.Namespace,
				Name:      cr.Name,
			}, cr)

			fmt.Println(err)

			Eventually(func() bool {
				err := k8sClient.Patch(ctx, cr, client.RawPatch(types.MergePatchType, patch))
				if err != nil {
					return false
				}
				if len(cr.Status.Nodes) == 5 {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())

			Eventually(func() bool {
				err := k8sClient.Delete(ctx, cr)
				return err == nil

			}, timeout, interval).Should(BeTrue())

		})
	})
})
