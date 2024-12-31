package ctrlruntime

import (
	"context"
	"encoding/json"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type CRDWithConfigMapRef struct {
	metav1.TypeMeta   `json:",inline"` // inline flattens the child struct fields into the parent
	metav1.ObjectMeta `json:"metadata,omitempty"`
	ConfigMapRef      corev1.LocalObjectReference `json:"configMapRef"`
}

// DeepCopyObject implements client.Object
func (cm *CRDWithConfigMapRef) DeepCopyObject() runtime.Object {
	return deepCopyObject(cm)
}

type CRDWithConfigMapRefList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CRDWithConfigMapRef `json:"items"`
}

func (cm *CRDWithConfigMapRefList) DeepCopyObject() runtime.Object {
	return deepCopyObject(cm)
}

func deepCopyObject(arg any) runtime.Object {
	argBytes, err := json.Marshal(arg)
	if err != nil {
		panic(err)
	}
	out := &CRDWithConfigMapRefList{}
	if err := json.Unmarshal(argBytes, out); err != nil {
		panic(err)
	}
	return out
}

func CRDWithConfigMapRefController() {
	log := ctrl.Log.WithName("crd-with-config-map-ref")

	manager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		log.Error(err, "failed to create manager")
		os.Exit(1)
	}

	err = ctrl.NewControllerManagedBy(manager).
		Watches(&corev1.ConfigMap{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, cm client.Object) []ctrl.Request {
			crList := &CRDWithConfigMapRefList{}
			if err := manager.GetClient().List(ctx, crList); err != nil {
				manager.GetLogger().Error(err, "while listing CRDWithConfigMapRef")
				return nil
			}

			reqs := make([]ctrl.Request, 0, len(crList.Items))
			for _, item := range crList.Items {
				if item.ConfigMapRef.Name == cm.GetName() {
					reqs = append(reqs, ctrl.Request{
						NamespacedName: types.NamespacedName{
							Namespace: item.GetNamespace(),
							Name:      item.GetName(),
						},
					})
				}
			}
			return reqs

		})).
		Complete(reconcile.Func(func(ctx context.Context, r reconcile.Request) (reconcile.Result, error) {
			// add business logic
			return reconcile.Result{}, nil
		}))
	if err != nil {
		log.Error(err, "failed to create controller")
		os.Exit(1)
	}

	if err := manager.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "failed to start manager")
		os.Exit(1)
	}
}
