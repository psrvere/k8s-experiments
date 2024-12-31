package ctrlruntime

import (
	"context"
	"fmt"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReplicaSetControllerExample() {
	log := ctrl.Log.WithName("replica-set-controller-example")

	manager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		log.Error(err, "failed to create manager")
		os.Exit(1)
	}

	// corev1 is the old api group. It includes resources like Pods, Services, Nodes, ConfigMaps, Secrets, etc
	// appsv1 is newer api group. It includes resources like Deployments, ReplicaSeta, StatefulSets, DaemonSets
	err = ctrl.NewControllerManagedBy(manager).For(&appsv1.ReplicaSet{}).Owns(&corev1.Pod{}).Complete(&ReplicaSetReconciler{Client: manager.GetClient()})
	if err != nil {
		log.Error(err, "could not create controller")
		os.Exit(1)
	}

	// SetupSignalHandler is respondible for graceful and forced shutdown
	if err := manager.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "failed to start manager")
		os.Exit(1)
	}
}

type ReplicaSetReconciler struct {
	client.Client
}

func (r *ReplicaSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	rs := &appsv1.ReplicaSet{}
	err := r.Get(ctx, req.NamespacedName, rs)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get replica set: %v", err)
	}

	pods := &corev1.PodList{}
	err = r.List(ctx, pods, client.InNamespace(req.Namespace), client.MatchingLabels(rs.Spec.Template.Labels))
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get pod list: %v", err)
	}

	rs.Labels["pod-count"] = fmt.Sprintf("%v", len(pods.Items))
	err = r.Update(ctx, rs)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update replia set: %v", err)
	}

	return ctrl.Result{}, nil
}
