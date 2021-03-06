package controllers

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"time"

	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cachev1alpha1 "github.com/mchirico/memcached-operator/api/v1alpha1"
)

// MemcachedReconciler reconciles a Memcached object
type MemcachedReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
	InTest   bool
}

// +kubebuilder:rbac:groups=*,resources=memcacheds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=*,resources=memcacheds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=*,resources=memcacheds/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;
// +kubebuilder:rbac:groups=*,resources=events,verbs=*

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Memcached object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *MemcachedReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("memcached", req.NamespacedName)

	// Fetch the Memcached instance
	memcached := &cachev1alpha1.Memcached{}
	err := r.Get(ctx, req.NamespacedName, memcached)

	if requeue, err := r.notifier(ctx, memcached, "t0", "Starting PIG 0"); requeue || err != nil {
		return ctrl.Result{Requeue: requeue}, err
	}

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Memcached resource not found. Ignoring since object must be deleted")
			r.Recorder.Event(memcached, corev1.EventTypeNormal, "Memcached resource not found", "Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Memcached")
		r.Recorder.Event(memcached, corev1.EventTypeWarning, "Failed to get Memcached", fmt.Sprintf("Error: %s", err.Error()))
		return ctrl.Result{}, err
	}

	if memcached.IsBeingDeleted() {
		r.Log.Info(fmt.Sprintf("HandleFinalizer for %v", req.NamespacedName))
		r.Recorder.Event(memcached, corev1.EventTypeNormal, "Start Delete Memcached", req.NamespacedName.String())
		if err := r.handleFinalizer(ctx, memcached); err != nil {
			r.Recorder.Event(memcached, corev1.EventTypeWarning, "Error Delete Memcached", fmt.Sprintf("Error: %s", err.Error()))
			return ctrl.Result{}, fmt.Errorf("error when handling finalizer: %w", err)
		}
		r.Recorder.Event(memcached, corev1.EventTypeNormal, "Deleted", "Object finalizer is deleted")
		return ctrl.Result{}, nil
	}

	if !memcached.HasFinalizer(cachev1alpha1.MemcachedFinalizerName) {
		r.Log.Info(fmt.Sprintf("AddFinalizer for %v", req.NamespacedName))
		if err := r.addFinalizer(ctx, memcached); err != nil {
			return ctrl.Result{}, fmt.Errorf("error when adding finalizer: %w", err)
		}
		r.Recorder.Event(memcached, corev1.EventTypeNormal, "Added", "Object finalizer is added")
		return ctrl.Result{}, nil
	}

	if requeue, err := r.notifier(ctx, memcached, "t1", "Starting PIG check deployment"); requeue || err != nil {
		return ctrl.Result{Requeue: requeue}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: memcached.Name, Namespace: memcached.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep, err := r.deploymentForMemcached(memcached)
		if err != nil {
			return ctrl.Result{}, err
		}
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	if requeue, err := r.notifier(ctx, memcached, "t2", "PIG Ensure the deployment size is the same as the spec"); requeue || err != nil {
		return ctrl.Result{Requeue: requeue}, err
	}

	// Ensure the deployment size is the same as the spec
	size := memcached.Spec.Size
	if *found.Spec.Replicas != size {

		r.Recorder.Event(memcached, corev1.EventTypeNormal, "Start Update memcached.Spec.Size", fmt.Sprintf("New size: (%d), Old size: (%d)", size, found.Spec.Replicas))

		found.Spec.Replicas = &size

		err = r.Update(ctx, found)
		if err != nil {
			r.Recorder.Event(memcached, corev1.EventTypeWarning, "Failed to  Update memcached.Spec.Size", fmt.Sprintf("New size: (%d), Old size: (%d), Err: %s", size, found.Spec.Replicas, err.Error()))

			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}

		r.Recorder.Event(memcached, corev1.EventTypeNormal, "Updated memcached.Spec.Size", fmt.Sprintf("New size: (%d), Old size: (%d), Err: %s", size, found.Spec.Replicas, err.Error()))

		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	if requeue, err := r.notifier(ctx, memcached, "t3", "PIG point 2"); requeue || err != nil {
		return ctrl.Result{Requeue: requeue}, err
	}

	// Update the Memcached status with the pod names
	// List the pods for this memcached's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(memcached.Namespace),
		client.MatchingLabels(labelsForMemcached(memcached.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "Memcached.Namespace", memcached.Namespace, "Memcached.Name", memcached.Name)
		return ctrl.Result{}, err
	}

	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, memcached.Status.Nodes) || int32(len(podNames)) != size {
		memcached.Status.Nodes = podNames
		// We only want last 8 in status
		if len(memcached.Status.Msg) > 7 {
			memcached.Status.Msg = append(memcached.Status.Msg[len(memcached.Status.Msg)-7:], statusMessageFromSize(len(podNames), size))
		} else {
			memcached.Status.Msg = append(memcached.Status.Msg, statusMessageFromSize(len(podNames), size))
		}

		err := r.Status().Update(ctx, memcached)
		if err != nil {
			log.Error(err, "Failed to update Memcached status on nodes")
			return ctrl.Result{}, err
		}

		r.Recorder.Event(memcached, corev1.EventTypeNormal, "Updated memcached.Status", fmt.Sprintf("Status: (%s)", statusMessageFromSize(len(podNames), size)))

		log.Info("Status updated. WE'RE GOOD!")
		return ctrl.Result{Requeue: true}, nil
	}

	if requeue, err := r.notifier(ctx, memcached, "done", "PIG all done"); requeue || err != nil {
		return ctrl.Result{Requeue: requeue}, err
	}

	r.Recorder.Event(memcached, corev1.EventTypeNormal, "memcached bottom loop", "All Good. Declarative Target Hit")
	return ctrl.Result{}, nil
}

// deploymentForMemcached returns a memcached Deployment object
func (r *MemcachedReconciler) deploymentForMemcached(m *cachev1alpha1.Memcached) (*appsv1.Deployment, error) {
	ls := labelsForMemcached(m.Name)
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   "memcached:1.4.36-alpine",
						Name:    "memcached",
						Command: []string{"memcached", "-m=64", "-o", "modern", "-v"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 11211,
							Name:          "memcached",
						}},
					}},
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	if !r.InTest {
		if err := ctrl.SetControllerReference(m, dep, r.Scheme); err != nil {
			return dep, err
		}
	}
	return dep, nil
}

func statusMessageFromSize(len int, size int32) string {
	if int32(len) != size {
		return fmt.Sprintf("Size Issue: (%d,%d)  %s", size, len, time.Now())
	}
	return fmt.Sprintf("Size Match: (%d,%d)  %s", size, len, time.Now())
}

// labelsForMemcached returns the labels for selecting the resources
// belonging to the given memcached CR name.
func labelsForMemcached(name string) map[string]string {
	return map[string]string{"app": "memcached", "memcached_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// SetupWithManager sets up the controller with the Manager.
func (r *MemcachedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.Memcached{}).
		Owns(&appsv1.Deployment{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Complete(r)
}
