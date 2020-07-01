/*

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
	"reflect"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"
	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	appsv1 "zookeeper-controller/api/v1"
	dp "zookeeper-controller/controllers/deployment"
	svc "zookeeper-controller/controllers/service"
)

// ZookeeperReconciler reconciles a Zookeeper object
type ZookeeperReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.handpay.cn,resources=zookeepers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.handpay.cn,resources=zookeepers/status,verbs=get;update;patch

func (r *ZookeeperReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("zookeeper", req.NamespacedName)

	// your logic here
	var zk appsv1.Zookeeper
	err := r.Get(ctx, req.NamespacedName, &zk)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	} else {

		// 关联关联OwnerReferences和Finalizer
		if err := updateComponent(r, ctx, zk); err != nil {
			return ctrl.Result{}, err
		}
		// 创建或者更新service
		serviceList := svc.ServiceLogic(zk.Spec, req)
		if err := CreateOrUpdateService(serviceList, zk, r, ctx); err != nil {
			return ctrl.Result{}, err
		}
		// 创建或者更新deployment
		deployList := dp.DeployMentLogic(zk.Spec, req)
		if err := CreateOrUpdateDeployMent(deployList, zk, r, ctx); err != nil {
			return ctrl.Result{}, err

		}
	}

	// 判断资源是否是删除
	if zk.DeletionTimestamp != nil {
		log.Info("资源删除：", req.Name, "  环境：", req.Namespace)
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, err

}

func CreateOrUpdateDeployMent(deployList []v1.Deployment, zk appsv1.Zookeeper, r *ZookeeperReconciler, ctx context.Context) (err error) {
	for _, k := range deployList {
		if k.ObjectMeta.OwnerReferences == nil {
			if err := ctrl.SetControllerReference(&zk, &k, r.Scheme); err != nil {
				log.Info("关联OwnerReferences错误")
				return err
			}
		}
		isExist := &v1.Deployment{}
		key, _ := client.ObjectKeyFromObject(&k)
		if err := r.Get(ctx, key, isExist); err != nil {
			if errors.IsNotFound(err) {
				log.Info("deployment 新建资源: ", k.Name)
				if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &k, func() error {
					return nil
				}); err != nil {
					log.Info("新建资源失败")
					return err
				}
			}
		} else {
			if !reflect.DeepEqual(k.Spec.Template.Spec.Volumes, isExist.Spec.Template.Spec.Volumes) ||
				!reflect.DeepEqual(k.Spec.Template.Spec.Containers[0].Env, isExist.Spec.Template.Spec.Containers[0].Env) ||
				!reflect.DeepEqual(k.Spec.Template.Spec.Containers[0].Image, isExist.Spec.Template.Spec.Containers[0].Image) ||
				!reflect.DeepEqual(k.Spec.Template.Spec.Containers[0].Ports, isExist.Spec.Template.Spec.Containers[0].Ports) {
				isExist.Spec = k.Spec
				log.Info("Deployment 资源更新: ", k.Name)
				if err := r.Update(context.TODO(), isExist); err != nil {
					return err
				}
			}
		}
	}
	return err
}

func CreateOrUpdateService(serviceList []apiv1.Service, zk appsv1.Zookeeper, r *ZookeeperReconciler, ctx context.Context) (err error) {
	for _, k := range serviceList {
		if k.ObjectMeta.OwnerReferences == nil {
			if err := ctrl.SetControllerReference(&zk, &k, r.Scheme); err != nil {
				log.Info("关联OwnerReferences错误")
				return err
			}
		}
		isExist := &apiv1.Service{}
		key, _ := client.ObjectKeyFromObject(&k)
		if err := r.Get(ctx, key, isExist); err != nil {
			if errors.IsNotFound(err) {
				log.Info("Service 新建资源: ", k.Name)
				if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &k, func() error {
					return nil
				}); err != nil {
					log.Info("Service 新建资源失败: ", k.Name)
					return err
				}
			}
		} else {
			if !reflect.DeepEqual(k.Spec.ClusterIP, isExist.Spec.ClusterIP) ||
				!reflect.DeepEqual(k.Spec.Selector, isExist.Spec.Selector) {
				k.Spec.ClusterIP = isExist.Spec.ClusterIP
				isExist.Spec = k.Spec
				log.Info("Service 资源更新:", k.Name)
				if err := r.Update(context.TODO(), isExist); err != nil {
					return err
				}
			}
		}
	}
	return err
}

func updateComponent(r *ZookeeperReconciler, ctx context.Context, zk appsv1.Zookeeper) (err error) {
	// Finalizer 异步删除数据
	myFinalizerName := "Finalizer.handpay.cn"
	if zk.ObjectMeta.DeletionTimestamp.IsZero() {
		// 添加OwnerReferences
		if zk.ObjectMeta.OwnerReferences == nil {
			log.Info("添加OwnerReferences")
			zk.ObjectMeta.OwnerReferences = getOwnerReferences(&zk)
			if err := r.Update(ctx, &zk); err != nil {
				log.Info("更新OwnerReferences错误")
			}
		}
		// 添加Finalizer
		if !containsString(zk.ObjectMeta.Finalizers, myFinalizerName) {
			zk.ObjectMeta.Finalizers = append(zk.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Update(context.Background(), &zk); err != nil {
			}
		}
	} else {
		if containsString(zk.ObjectMeta.Finalizers, myFinalizerName) {
			if err := r.deleteExternalResources(&zk); err != nil {
			}
			zk.ObjectMeta.Finalizers = removeString(zk.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Update(context.Background(), &zk); err != nil {
			}
		}
	}
	return err
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func (r *ZookeeperReconciler) deleteExternalResources(META *appsv1.Zookeeper) error {
	//
	// delete any external resources associated with the cronJob
	//
	// Ensure that delete implementation is idempotent and safe to invoke
	// multiple types for same object.
	return nil
}

func getOwnerReferences(meta *appsv1.Zookeeper) []metav1.OwnerReference {
	ownerRefs := []metav1.OwnerReference{}
	ownerRef := metav1.OwnerReference{}
	var k8sGC bool = true
	ownerRef.APIVersion = meta.APIVersion
	ownerRef.Name = meta.Name
	ownerRef.Kind = meta.Kind
	ownerRef.UID = meta.UID
	ownerRef.Controller = &k8sGC
	ownerRef.BlockOwnerDeletion = &k8sGC
	return append(ownerRefs, ownerRef)

}

func (r *ZookeeperReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Zookeeper{}).
		Complete(r)
}
