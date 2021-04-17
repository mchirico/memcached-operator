package controllers

import (
	"context"
	"fmt"
	cachev1alpha1 "github.com/mchirico/memcached-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"strings"
	"time"
)

// Send notification if return true, requeue
func (r *MemcachedReconciler) notifier(ctx context.Context, memcached *cachev1alpha1.Memcached, msg string) (bool, error) {

	tag := strings.ReplaceAll(memcached.Name+msg, " ", "_")
	if err := r.Client.Create(ctx, makeEvent(memcached, msg, tag)); err != nil {
		if apiErr, ok := err.(*errors.StatusError); ok {

			if apiErr.Status().Reason == "AlreadyExists" {
				key := types.NamespacedName{
					Name:      memcached.Name,
					Namespace: memcached.Namespace,
				}
				event := &corev1.Event{}
				err := r.Client.Get(ctx, key, event)
				if err != nil {
					return false, nil
				}
				event.Count += 1
				event.LastTimestamp = metav1.Time{Time: time.Now()}
				err = r.Client.Update(ctx, event)
				if err != nil {
					r.Recorder.Event(memcached, corev1.EventTypeWarning, "Failed to update notifier event", fmt.Sprintf("Error: %s", err.Error()))
				}
			}

		}
		r.Log.Info(fmt.Sprintf("Event exists %s", memcached.GetName()))
	}
	return false, nil
}

func makeEvent(obj *cachev1alpha1.Memcached, msg string, tag string) *corev1.Event {

	event := &corev1.Event{
		TypeMeta:            metav1.TypeMeta{},
		ObjectMeta:          metav1.ObjectMeta{Namespace: obj.Namespace, Name: tag},
		InvolvedObject:      corev1.ObjectReference{Kind: obj.Kind, ResourceVersion: obj.ResourceVersion},
		Reason:              "Some silly reason",
		Message:             msg,
		Source:              corev1.EventSource{},
		FirstTimestamp:      metav1.Time{Time: time.Now()},
		LastTimestamp:       metav1.Time{Time: time.Now()},
		Count:               0,
		Type:                "Special",
		EventTime:           metav1.MicroTime{},
		Series:              nil,
		Action:              "Dropped jaw",
		Related:             nil,
		ReportingController: "A Giant Panda",
		ReportingInstance:   "",
	}
	return event
}
