package front

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	frontv1 "src/op-demo-front/pkg/apis/front/v1"
)

var log = logf.Log.WithName("controller_front")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Front Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileFront{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("front-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Front
	err = c.Watch(&source.Kind{Type: &frontv1.Front{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Front
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &frontv1.Front{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileFront implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileFront{}

// ReconcileFront reconciles a Front object
type ReconcileFront struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Front object and makes changes based on the state read
// and what is in the Front.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileFront) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Front")

	// Fetch the Front instance
	instance := &frontv1.Front{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	//qweLogger := log.WithValues("zzz")
	//qweLogger.Info("zzz")

	//创建 or 更新 deployment、service
	dep := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace},dep)
	if err != nil && errors.IsNotFound(err) {
		dep := newDepForCR(instance)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return reconcile.Result{}, err
		}
		//instance.Status.DeploymentStatus = dep.Status
		//err := r.client.Status().Update(context.TODO(), instance)
		//if err != nil {
		//	reqLogger.Error(err, "Failed to update Front status")
		//	return reconcile.Result{}, err
		//}
	}else {
		depNew := newDepForCR(instance)
		if !reflect.DeepEqual(dep.Spec, depNew.Spec) {
			reqLogger.Info("Updating a Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			err := r.client.Update(context.TODO(), depNew);
			if err != nil {
				reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
				return reconcile.Result{}, err
			}
		}
	}
	svc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace},svc)
	if err != nil && errors.IsNotFound(err) {
		svc := newSvcForCR(instance)
		reqLogger.Info("Create a new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		if err := r.client.Create(context.TODO(), svc); err != nil {
			reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return reconcile.Result{}, err
		}
	}else {
		svcNew := newSvcForCR(instance)
		if !reflect.DeepEqual(svc.Spec.Ports, svcNew.Spec.Ports) {
			svc.Spec.Ports = svcNew.Spec.Ports
			reqLogger.Info("Updating a Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			if err := r.client.Update(context.TODO(), svc); err != nil {
				reqLogger.Error(err, "Failed to update Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
				return reconcile.Result{}, err
			}
		}
	}

	// Update status
	// Update the Memcached status with the pod names
	// List the pods for this memcached's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(labelsForFront(instance.Name)),
	}
	if err = r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "Front.Namespace", instance.Namespace, "Front.Name", instance.Name)
		return reconcile.Result{}, err
	}
	podstatus := getPodNamesAndStatus(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podstatus, instance.Status) {
		instance.Status.Status = podstatus
		reqLogger.Info("Updating a Front Status", "Front.Namespace", svc.Namespace, "Front.Name", svc.Name)
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update instance status")
			return reconcile.Result{}, err
		}
	}
	//if !reflect.DeepEqual(dep.Status, instance.Status.DeploymentStatus) {
	//	instance.Status.DeploymentStatus = dep.Status
	//	reqLogger.Info("Updating a status", "front.Namespace", instance.Namespace, "front.Name", instance.Name)
	//	err := r.client.Status().Update(context.TODO(), instance)
	//	if err != nil {
	//		reqLogger.Error(err, "Failed to update Front status")
	//		return reconcile.Result{}, err
	//	}
	//	patch := client.MergeFrom(instance)
	//	err = r.client.Status().Patch(context.TODO(), instance, patch)
	//	if err != nil {
	//		reqLogger.Error(err, "Failed to update Front status")
	//		return reconcile.Result{}, err
	//	}
	//}

	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newDepForCR(cr *frontv1.Front) *appsv1.Deployment {
	replicas := cr.Spec.Replicas
	containerPorts := []corev1.ContainerPort{}
	for _, svcPort := range cr.Spec.Ports {
		cport := corev1.ContainerPort{}
		cport.ContainerPort = svcPort.TargetPort.IntVal
		containerPorts = append(containerPorts, cport)
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: cr.Name,
			Namespace: cr.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cr, schema.GroupVersionKind{
					Group: appsv1.SchemeGroupVersion.Group,
					Version: appsv1.SchemeGroupVersion.Version,
					Kind: "Front",
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": cr.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": cr.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  cr.Name + "-" + "pod",
							Image: cr.Spec.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: containerPorts,
						},
					},
				},
			},
		},
	}
	return deployment
}

func newSvcForCR(cr *frontv1.Front) *corev1.Service {
	containerPorts := []corev1.ContainerPort{}
	for _, svcPort := range cr.Spec.Ports {
		cport := corev1.ContainerPort{}
		cport.ContainerPort = svcPort.TargetPort.IntVal
		containerPorts = append(containerPorts, cport)
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: cr.Name,
			Namespace: cr.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cr, schema.GroupVersionKind{
					Group: appsv1.SchemeGroupVersion.Group,
					Version: appsv1.SchemeGroupVersion.Version,
					Kind: "Front",
				}),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app" : cr.Name,
			},
			Ports: cr.Spec.Ports,
			Type: corev1.ServiceTypeNodePort,
		},
	}
	return service
}

func int32Ptr(i int32) *int32 { return &i }

func labelsForFront(name string) map[string]string {
	return map[string]string{"app": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNamesAndStatus(pods []corev1.Pod) []frontv1.PodStatus {

	frontStatus := []frontv1.PodStatus{}
	getstatus := frontv1.PodStatus{}
	for _, pod := range pods {
		getstatus.PodNames = pod.Name
		//getstatus.PodStatus = pod.Status
		frontStatus = append(frontStatus,getstatus)
		//frontStatus.podNames = append(frontStatus.podNames, pod.Name)
		//podStatus = append(podStatus, pod.Status)
	}
	return frontStatus
}

type PodStatus struct {
	podNames string
	podStatus corev1.PodStatus
}