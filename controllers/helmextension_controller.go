/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	extensionv1alpha1 "github.com/sankalp-r/extension-operator/api/v1alpha1"
	"github.com/sankalp-r/extension-operator/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const extensionFinalizer = "extension.example.com/finalizer"

// HelmextensionReconciler reconciles a Helmextension object
type HelmextensionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=extension.example.com,resources=helmextensions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=extension.example.com,resources=helmextensions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=extension.example.com,resources=helmextensions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Helmextension object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *HelmextensionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// your logic here
	log := r.Log.WithValues("helmextension", req.NamespacedName)
	ext := &extensionv1alpha1.Helmextension{}
	err := r.Client.Get(ctx, req.NamespacedName, ext)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Error(err, "HelmExtension not found")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get HelmExtension")
		return ctrl.Result{}, err
	}

	fmt.Println(ext)

	isExtensionMarkedToBeDeleted := ext.GetDeletionTimestamp() != nil

	if isExtensionMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(ext, extensionFinalizer) {
			err = r.finalizeExtension(ext)
			if err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(ext, extensionFinalizer)
			err = r.Update(ctx, ext)
			if err != nil {
				log.Error(err, "Failed to get HelmExtension")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}

	if !controllerutil.ContainsFinalizer(ext, extensionFinalizer) {
		controllerutil.AddFinalizer(ext, extensionFinalizer)
		err = r.Update(ctx, ext)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	args := make(map[string]string)
	args["version"] = ext.Spec.Version
	err = util.RepoAdd(ext.Spec.Repo, ext.Spec.Url)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = util.RepoUpdate()
	if err != nil {
		return ctrl.Result{}, err
	}
	err = util.InstallChart(ext.Name, ext.Spec.Repo, ext.Spec.Chart, args)
	if err != nil {
		log.Error(err, "Error: ")
		return ctrl.Result{}, err
	}

	res, err := util.GetReleaseStatus(ext.Name)
	if err != nil {
		return ctrl.Result{}, err
	} else {
		ext.Status.State = res
		err = r.Status().Update(ctx, ext)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelmextensionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&extensionv1alpha1.Helmextension{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				res, _ := util.GetReleaseStatus(e.Object.GetName())
				return res != "deployed"

			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return e.ObjectNew.GetGeneration() != e.ObjectOld.GetGeneration()
			},
		}).
		Complete(r)
}

func (r *HelmextensionReconciler) finalizeExtension(ext *extensionv1alpha1.Helmextension) error {
	log.Print("Deleting Helm chart")
	return util.UnInstallChart(ext.Name)
}
