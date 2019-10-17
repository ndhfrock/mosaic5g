package mosaic5g

import (
	"context"

	mosaic5gv1alpha1 "github.com/ndhfrock/mosaic5g/pkg/apis/mosaic5g/v1alpha1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_mosaic5g")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Mosaic5g Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMosaic5g{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("mosaic5g-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Mosaic5g
	err = c.Watch(&source.Kind{Type: &mosaic5gv1alpha1.Mosaic5g{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are o: 11:30wned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Mosaic5g
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &mosaic5gv1alpha1.Mosaic5g{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMosaic5g implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMosaic5g{}

// ReconcileMosaic5g reconciles a Mosaic5g object
type ReconcileMosaic5g struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Mosaic5g object and makes changes based on the state read
// and what is in the Mosaic5g.Spec
//Note :
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
// How to reconcile Mosaic5g:
// 1. Create MySQL, OAI-CN and OAI-RAN in order
// 2. If the configuration changed, restart all OAI PODs
func (r *ReconcileMosaic5g) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Mosaic5g")

	// Fetch Mosaic5g instance
	instance := &mosaic5gv1alpha1.Mosaic5g{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	// If there is already Mosaic5g instances, delete it
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue

			//Delete unused ConfigMap
			//Config map is where Mosaic 5g spec config is
			conf := r.genConfigMap(instance)
			conf.Namespace = "default"
			err = r.client.Delete(context.TODO(), conf)
			if err != nil {
				reqLogger.Error(err, "Failed to delete Config Map")
			}
			reqLogger.Info("Mosaic5g resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		//Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get Mosaic5G instance")
		return reconcile.Result{}, err
	}

	//if there is none, create it
	new := r.genConfigMap(instance)
	config := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: new.GetName(), Namespace: instance.Namespace}, config)
	// if gen configmap succesfull continue to saves it in Kubernetes cluster
	if errors.IsNotFound(err) && err != nil {
		//Create a configmap from mosaic5g spec for cn and ran
		reqLogger.Info("Creating a new ConfigMap for CN and RAN")
		conf := r.genConfigMap(instance)
		reqLogger.Info("conf", "content", conf)
		err = r.client.Create(context.TODO(), conf)
		//If fail
		if err != nil {
			reqLogger.Error(err, "Failed to create a new ConfigMap")
		}
		//if genconfigmap not succesfull
	} else if err != nil {
		reqLogger.Error(err, "Generate CongifMap failed")
		return reconcile.Result{}, err
	}

	//Everything works fine, Reconcile will end
	return reconcile.Result{}, nil
}

//generate configmap from Reconcile Mosaic5g's spec
func (r *ReconcileMosaic5g) genConfigMap(m *mosaic5gv1alpha1.Mosaic5g) *corev1.ConfigMap {
	genLogger := log.WithValues("Mosaic5g", "genConfigMap")
	//Make specs into map[name][name]
	datas := make(map[string]string)
	d, err := yaml.Marshal(&m.Spec)
	if err != nil {
		log.Error(err, "Marshal fail")
	}
	datas["conf.yaml"] = string(d)
	cm := corev1.ConfigMap{
		Data: datas,
	}
	cm.Name = "Config"
	cm.Namespace = m.Namespace
	genLogger.Info("Done Creating Config Map")
	return &cm
}
