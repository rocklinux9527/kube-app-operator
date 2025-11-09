/*
Copyright 2025.

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

package controller

import (
	"context"
	appsv1alpha1 "github.com/k8s/kube-app-operator/api/v1alpha1"
	custom "github.com/k8s/kube-app-operator/internal/custom"
	"k8s.io/apimachinery/pkg/api/errors"
	// appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// KubeAppReconciler reconciles a kubeapp object

type KubeAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.dgplus.com,resources=digiapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.dgplus.com,resources=digiapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.dgplus.com,resources=digiapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the kubeapp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile



// defines Reconcile core logic 

// 创建日志记录器
var log_controller = logf.Log.WithName("controller-creator")

func (r *KubeAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var kubeapp appsv1alpha1.KubeApp

	// TODO(user): your logic here

	if err := r.Get(ctx, req.NamespacedName, &kubeapp); err != nil {
		if errors.IsNotFound(err) {
			log_controller.Info("kubeapp resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	}


	//  controller deployment resource create or delete  ture eq create  false eq delete 

	if kubeapp.Spec.EnableDeployment {
		dep, err := custom.NewDeployment(&kubeapp, req.Namespace)
		if err != nil {
			return ctrl.Result{}, err
		}
		ctrl.SetControllerReference(&kubeapp, dep, r.Scheme)
		if err := r.createOrUpdate(ctx, dep); err != nil {
			return ctrl.Result{}, err
		}
	}else{
		if err := custom.DeleteDeployment(ctx, r.Client, &kubeapp, req.Namespace); err != nil {
			return ctrl.Result{}, err
		}
	}
	//  controller service resource create or delete  ture eq create  false eq delete
	if kubeapp.Spec.EnableService {
		svc, err := custom.NewService(&kubeapp, req.Namespace)
		if err != nil {
			return ctrl.Result{}, err
		}
		ctrl.SetControllerReference(&kubeapp, svc, r.Scheme)
		if err := r.createOrUpdate(ctx, svc); err != nil {
			return ctrl.Result{}, err
		}
	}else {
		if err := custom.DeleteService(ctx, r.Client, &kubeapp, req.Namespace); err != nil {
			return ctrl.Result{}, err
		}
	}

	//  controller ingress resource create or delete  ture eq create  false eq delete
	if kubeapp.Spec.EnableIngress {
		ing, err := custom.NewIngress(&kubeapp, req.Namespace)
		if err != nil {
			return ctrl.Result{}, err
		}
		ctrl.SetControllerReference(&kubeapp, ing, r.Scheme)
		if err := r.createOrUpdate(ctx, ing); err != nil {
			return ctrl.Result{}, err
		}
	}else {
		if err := custom.DeleteIngress(ctx, r.Client, &kubeapp, req.Namespace); err != nil {
			return ctrl.Result{}, err
		}
	}

	// controller pvc resource create or delete  ture eq create  false eq delete 
	/* EnablePvc = false 且 ForceDelete = true  删除 PVC
	   EnablePvc = false 且 ForceDelete = false 仅日志提醒
	   EnablePvc = true 创建或更新 PVC（Apply）
	*/
	var pvcName string
	if kubeapp.Spec.Pvc != nil && kubeapp.Spec.Pvc.Name != "" {
		pvcName = kubeapp.Spec.Pvc.Name
	} else {
		pvcName = kubeapp.Name
	}

	// 创建或删除 PVC
	if !kubeapp.Spec.EnablePvc {
		if kubeapp.Spec.Pvc != nil && kubeapp.Spec.Pvc.ForceDelete {
			log_controller.Info("启用了 PVC 强制删除标志，开始尝试删除 PVC","PVC名称", pvcName, "命名空间", req.Namespace)
			err := custom.DeletePvc(ctx, r.Client, &kubeapp, req.Namespace)
			if err != nil {
				log_controller.Error(err, "PVC 删除失败", "PVC名称", pvcName)
				return ctrl.Result{}, err
			}

			log_controller.Info("PVC 删除成功", "PVC名称", pvcName)
		} else {
			log_controller.Info("PVC 被禁用，但未启用强制删除。为保护数据，不执行删除操作，请管理员手动删除。","PVC名称", pvcName, "命名空间", req.Namespace)
		}

		return ctrl.Result{}, nil
	} else {
		// 启用了 PVC，尝试创建
		pvcObj, err := custom.NewPvc(ctx, &kubeapp, req.Namespace)
		if err != nil {
			log_controller.Error(err, "构建 PVC 对象失败", "PVC名称", pvcName)
			return ctrl.Result{}, err
		}

		// 使用 server-side apply 方式创建 PVC
		if err := r.Client.Patch(ctx, pvcObj, client.Apply, &client.PatchOptions{
			FieldManager: "kubeapp-operator",
			Force:        pointer.Bool(true),
		}); err != nil {
			log_controller.Error(err, "PVC apply 失败", "PVC名称", pvcName)
			return ctrl.Result{}, err
		}

		log_controller.Info("PVC 创建或更新成功", "PVC名称", pvcName)
	}
	return ctrl.Result{}, nil
}

func (r *KubeAppReconciler) createOrUpdate(ctx context.Context, obj client.Object) error {
	existing := obj.DeepCopyObject().(client.Object)
	err := r.Get(ctx, client.ObjectKeyFromObject(obj), existing)
	if errors.IsNotFound(err) {
		return r.Create(ctx, obj)
	} else if err != nil {
		return err
	}
	obj.SetResourceVersion(existing.GetResourceVersion())
	return r.Update(ctx, obj)
}


// SetupWithManager sets up the controller with the Manager.
func (r *KubeAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.KubeApp{}).
		Named("kubeapp").
		Complete(r)
}





