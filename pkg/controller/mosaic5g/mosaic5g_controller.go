package mosaic5g

import (
	"context"
	"reflect"
	"time"

	Err "errors"

	"github.com/ndhfrock/mosaic5g/internal/util"
	mosaic5gv1alpha1 "github.com/ndhfrock/mosaic5g/pkg/apis/mosaic5g/v1alpha1"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"

	//v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
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

	// Fetch the Mosaic5g instance
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
				reqLogger.Error(err, "Failed to delete ")
			}
			reqLogger.Info("Mosaic5g resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get Mosaic5g instance")
		return reconcile.Result{}, err
	}

	//Create config map for mosaic 5g instances
	new := r.genConfigMap(instance)
	config := &v1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: new.GetName(), Namespace: instance.Namespace}, config)
	if err != nil && errors.IsNotFound(err) {
		// Create a configmap for cn and ran
		reqLogger.Info("Creating a new ConfigMap")
		conf := r.genConfigMap(instance)
		reqLogger.Info("conf", "content", conf)
		err = r.client.Create(context.TODO(), conf)
		if err != nil {
			reqLogger.Error(err, "Failed to create new ConfigMap")
		}
	} else if err != nil {
		reqLogger.Error(err, "Generating ConfigMap failed")
		return reconcile.Result{}, err
	}

	// Define a new MySQL deployment
	mysql := &appsv1.Deployment{}
	mysqlDeployment := r.deploymentForMySQL(instance)
	// Check if MySQL deployment already exists, if not create a new one
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: mysqlDeployment.GetName(), Namespace: instance.Namespace}, mysql)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", mysqlDeployment.Namespace, "Deployment.Name", mysqlDeployment.Name)
		err = r.client.Create(context.TODO(), mysqlDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", mysqlDeployment.Namespace, "Deployment.Name", mysqlDeployment.Name)
			return reconcile.Result{}, err
		}
		// Define a new mysql service
		mysqlService := r.genMySQLService(instance)
		err = r.client.Create(context.TODO(), mysqlService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", mysqlService.Namespace, "Service.Name", mysqlService.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "MySQL Failed to get Deployment")
		return reconcile.Result{}, err
	}

	// Creat an oaicn deployment
	cn := &appsv1.Deployment{}
	cnDeployment := r.deploymentForCN(instance)
	// Check if the oai-cn deployment already exists, if not create a new one
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: cnDeployment.GetName(), Namespace: instance.Namespace}, cn)
	if err != nil && errors.IsNotFound(err) {
		if mysql.Status.ReadyReplicas == 0 {
			return reconcile.Result{Requeue: true}, Err.New("No mysql POD is ready, NO!!!")
		}
		reqLogger.Info("MME domain name " + instance.Spec.MmeDomainName)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", cnDeployment.Namespace, "Deployment.Name", cnDeployment.Name)
		err = r.client.Create(context.TODO(), cnDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", cnDeployment.Namespace, "Deployment.Name", cnDeployment.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully. Let's wait for it to be ready
		d, _ := time.ParseDuration("30s")
		return reconcile.Result{Requeue: true, RequeueAfter: d}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get a CN Deployment")
		return reconcile.Result{}, err
	}

	// Create an oaicn service, so that OAICN could connect with other pods
	service := &v1.Service{}
	cnService := r.genCNService(instance)
	// Check if the oai-cn service already exists, if not create a new one
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: cnService.GetName(), Namespace: instance.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(context.TODO(), cnService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", cnService.Namespace, "Service.Name", cnService.Name)
			return reconcile.Result{}, err
		}
	}

	flexran := &appsv1.Deployment{}
	flexranDeployment := r.deploymentForFlexRAN(instance)
	// If flexran true then deploy FlexRAN
	if instance.Spec.FlexRAN == true {
		// Creat a flexran deployment
		// Check if theflexran deployment already exists, if not create a new one
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: flexranDeployment.GetName(), Namespace: instance.Namespace}, flexran)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", flexranDeployment.Namespace, "Deployment.Name", flexranDeployment.Name)
			err = r.client.Create(context.TODO(), flexranDeployment)
			if err != nil {
				reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", flexranDeployment.Namespace, "Deployment.Name", flexranDeployment.Name)
				return reconcile.Result{}, err
			}
			// Deployment created successfully. Let's wait for it to be ready
			d, _ := time.ParseDuration("30s")
			return reconcile.Result{Requeue: true, RequeueAfter: d}, nil
		} else if err != nil {
			reqLogger.Error(err, "Failed to get a FlexRAN Deployment")
			return reconcile.Result{}, err
		}

		// Create an flexran service, so that flexran could connect with other pods
		fservice := &v1.Service{}
		flexranService := r.genFlexRANService(instance)
		// Check if the oai-cn service already exists, if not create a new one
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: flexranService.GetName(), Namespace: instance.Namespace}, fservice)
		if err != nil && errors.IsNotFound(err) {
			err = r.client.Create(context.TODO(), flexranService)
			if err != nil {
				reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", flexranService.Namespace, "Service.Name", flexranService.Name)
				return reconcile.Result{}, err
			}
		}
	}

	//if elasticsearch config is true, deploy elasticsearch
	// Define a new Elasticsearch statefulset
	es := &appsv1.StatefulSet{}
	esDeployment := r.deploymentForElasticsearch(instance)
	if instance.Spec.Elasticsearch == true {
		// Check if Elasticsearch StatefulSet already exists, if not create a new one
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: esDeployment.GetName(), Namespace: instance.Namespace}, es)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new StatefulSet", "Statefulset.Namespace", esDeployment.Namespace, "Statefulset.Name", esDeployment.Name)
			err = r.client.Create(context.TODO(), esDeployment)
			if err != nil {
				reqLogger.Error(err, "Failed to create new StatefulSet", "StatefulSet.Namespace", esDeployment.Namespace, "StatefulSet.Name", esDeployment.Name)
				return reconcile.Result{}, err
			}
			// Define a new elasticsearch service
			esService := r.genESService(instance)
			err = r.client.Create(context.TODO(), esService)
			if err != nil {
				reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", esService.Namespace, "Service.Name", esService.Name)
				return reconcile.Result{}, err
			}
			// Deployment created successfully - return and requeue
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Elasticsearch Failed to get StatefulSet")
			return reconcile.Result{}, err
		}
	}

	//if kibana config is true, deploy kibana
	// Define a new kibana deployment
	kib := &appsv1.Deployment{}
	kibDeployment := r.deploymentForKibana(instance)
	if instance.Spec.Kibana == true {
		// Check if Kibana Deployment already exists, if not create a new one
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: kibDeployment.GetName(), Namespace: instance.Namespace}, kib)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", kibDeployment.Namespace, "Deployment.Name", kibDeployment.Name)
			err = r.client.Create(context.TODO(), kibDeployment)
			if err != nil {
				reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", kibDeployment.Namespace, "Deployment.Name", kibDeployment.Name)
				return reconcile.Result{}, err
			}
			// Define a new kibana service
			kibService := r.genKibanaService(instance)
			err = r.client.Create(context.TODO(), kibService)
			if err != nil {
				reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", kibService.Namespace, "Service.Name", kibService.Name)
				return reconcile.Result{}, err
			}
			// Deployment created successfully - return and requeue
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Kibana Failed to get Deployment")
			return reconcile.Result{}, err
		}
	}

	//if drone config is true, deploy drone store app
	// Define a new drone deployment
	drone := &appsv1.Deployment{}
	droneDeployment := r.deploymentForDrone(instance)
	if instance.Spec.DroneStore == true {
		// Check if Drone Deployment already exists, if not create a new one
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: droneDeployment.GetName(), Namespace: instance.Namespace}, drone)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", droneDeployment.Namespace, "Deployment.Name", droneDeployment.Name)
			err = r.client.Create(context.TODO(), droneDeployment)
			if err != nil {
				reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", droneDeployment.Namespace, "Deployment.Name", droneDeployment.Name)
				return reconcile.Result{}, err
			}
			// Define a new drone service
			droneService := r.genDroneService(instance)
			err = r.client.Create(context.TODO(), droneService)
			if err != nil {
				reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", droneService.Namespace, "Service.Name", droneService.Name)
				return reconcile.Result{}, err
			}
			// Deployment created successfully - return and requeue
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "Drone Failed to get Deployment")
			return reconcile.Result{}, err
		}
	}

	//if rrmkpi config is true, deploy rrm-kpi store app
	// Define a new rrmkpi deployment
	rrmkpi := &appsv1.Deployment{}
	rrmkpiDeployment := r.deploymentForRRMKPI(instance)
	if instance.Spec.RRMKPIStore == true {
		// Check if RRM-KPI Deployment already exists, if not create a new one
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: rrmkpiDeployment.GetName(), Namespace: instance.Namespace}, rrmkpi)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", rrmkpiDeployment.Namespace, "Deployment.Name", rrmkpiDeployment.Name)
			err = r.client.Create(context.TODO(), rrmkpiDeployment)
			if err != nil {
				reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", rrmkpiDeployment.Namespace, "Deployment.Name", rrmkpiDeployment.Name)
				return reconcile.Result{}, err
			}
			// Define a new rrmkpi service
			rrmkpiService := r.genRRMKPIService(instance)
			err = r.client.Create(context.TODO(), rrmkpiService)
			if err != nil {
				reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", rrmkpiService.Namespace, "Service.Name", rrmkpiService.Name)
				return reconcile.Result{}, err
			}
			// Deployment created successfully - return and requeue
			return reconcile.Result{Requeue: true}, nil
		} else if err != nil {
			reqLogger.Error(err, "RRM-KPI Failed to get Deployment")
			return reconcile.Result{}, err
		}
	}

	// Create an oairan deployment
	ran := &appsv1.Deployment{}
	ranDeployment := r.deploymentForRAN(instance)
	// Check if the oai-ran deployment already exists, if not create a new one
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: ranDeployment.GetName(), Namespace: instance.Namespace}, ran)
	if err != nil && errors.IsNotFound(err) {
		if flexran.Status.ReadyReplicas == 0 && instance.Spec.FlexRAN == true {
			d, _ := time.ParseDuration("60s")
			return reconcile.Result{Requeue: true, RequeueAfter: d}, Err.New("No flexran POD is ready, 60 seconds backoff")
		}
		if cn.Status.ReadyReplicas == 0 {
			d, _ := time.ParseDuration("60s")
			return reconcile.Result{Requeue: true, RequeueAfter: d}, Err.New("No oai-cn POD is ready, 60 seconds backoff")
		}
		reqLogger.Info("CN are ready")
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", ranDeployment.Namespace, "Deployment.Name", ranDeployment.Name)
		err = r.client.Create(context.TODO(), ranDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", ranDeployment.Namespace, "Deployment.Name", ranDeployment.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "RAN Failed to get Deployment")
		return reconcile.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := instance.Spec.Size
	if *cn.Spec.Replicas != size {
		cn.Spec.Replicas = &size
		err = r.client.Update(context.TODO(), cn)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", cn.Namespace, "Deployment.Name", cn.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		reqLogger.Info("All deployment size are the same as the spec")
		return reconcile.Result{Requeue: true}, nil
	}

	// Update the Mosaic5g status with the pod names
	// List the pods for this instance's deployment
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(util.LabelsForMosaic5g(instance.GetName()))
	listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
	err = r.client.List(context.TODO(), listOps, podList)
	if err != nil {
		reqLogger.Error(err, "Failed to list pods", "Mosaic5g.Namespace", instance.Namespace, "Mosaic5g.Name", instance.Name)
		return reconcile.Result{}, err
	}
	podNames := util.GetPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
		instance.Status.Nodes = podNames
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update Mosaic5g status")
			return reconcile.Result{}, err
		}
	}

	// Check configmap is fine or not. If it's changed, update ConfigMap and restart cn ran
	if err == nil {
		if reflect.DeepEqual(new.Data, config.Data) {
			reqLogger.Info("newConf equals config")
		} else {
			reqLogger.Info("newConf does not equals config")
			reqLogger.Info("Update ConfigMap and deleting CN and RAN")
			err = r.client.Update(context.TODO(), new)
			//Should only kill the POD
			//Delete CN and ran pod first
			err = r.client.Delete(context.TODO(), cnDeployment)
			err = r.client.Delete(context.TODO(), ranDeployment)
			// Delete other pods that is set to false in the new config
			if instance.Spec.RRMKPIStore == false {
				err = r.client.Delete(context.TODO(), rrmkpiDeployment)
			}
			if instance.Spec.DroneStore == false {
				err = r.client.Delete(context.TODO(), droneDeployment)
			}
			if instance.Spec.Kibana == false {
				err = r.client.Delete(context.TODO(), kibDeployment)
			}
			if instance.Spec.Elasticsearch == false {
				err = r.client.Delete(context.TODO(), esDeployment)
			}
			if instance.Spec.FlexRAN == false {
				err = r.client.Delete(context.TODO(), flexranDeployment)
			}
			// Spec updated - return and requeue
			d, _ := time.ParseDuration("10s")
			return reconcile.Result{Requeue: true, RequeueAfter: d}, nil
		}

	}

	// Everything is fine, Reconcile ends
	return reconcile.Result{}, nil
}

// deploymentForRAN returns a Radio Access Network Deployment object
func (r *ReconcileMosaic5g) deploymentForRAN(m *mosaic5gv1alpha1.Mosaic5g) *appsv1.Deployment {
	//ls := util.LabelsForMosaic5g(m.Name)
	replicas := m.Spec.Size
	labels := make(map[string]string)
	labels["app"] = "ran"
	Annotations := make(map[string]string)
	Annotations["container.apparmor.security.beta.kubernetes.io/"+m.Name+"-"+"ran"] = "unconfined"
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "ran",
			Namespace:   m.Namespace,
			Annotations: Annotations,
			Labels:      labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           m.Spec.RANImage,
						Name:            "ran",
						Command:         []string{"/sbin/init"},
						SecurityContext: &corev1.SecurityContext{Privileged: util.NewTrue()},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "cgroup",
							ReadOnly:  true,
							MountPath: "/sys/fs/cgroup/",
						}, {
							Name:      "module",
							ReadOnly:  true,
							MountPath: "/lib/modules/",
						}, {
							Name:      "usrp",
							ReadOnly:  true,
							MountPath: "/dev/bus/usb/",
						}, {
							Name:      "mosaic5g-config",
							MountPath: "/root/config",
						}},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
							Name:          "mosaic5g-ran",
						}},
					}},
					Affinity: util.GenAffinity("ran"),
					Volumes: []corev1.Volume{{
						Name: "cgroup",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/sys/fs/cgroup/",
								Type: util.NewHostPathType("Directory"),
							},
						}}, {
						Name: "module",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/lib/modules/",
								Type: util.NewHostPathType("Directory"),
							},
						}}, {
						Name: "mosaic5g-config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: "mosaic5g-config"},
							},
						}}, {
						Name: "usrp",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/dev/bus/usb/",
								Type: util.NewHostPathType("Directory"),
							},
						}},
					},
				},
			},
		},
	}
	// Set Mosaic5g instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep
}

// genConfigMap will generate a configmap from ReconcileMosaic5g's spec
func (r *ReconcileMosaic5g) genConfigMap(m *mosaic5gv1alpha1.Mosaic5g) *v1.ConfigMap {
	genLogger := log.WithValues("Mosaic5g", "genConfigMap")
	// Make specs into map[name][value]
	datas := make(map[string]string)
	d, err := yaml.Marshal(&m.Spec)
	if err != nil {
		log.Error(err, "Marshal fail")
	}
	datas["conf.yaml"] = string(d)
	cm := v1.ConfigMap{
		Data: datas,
	}
	cm.Name = "mosaic5g-config"
	cm.Namespace = m.Namespace
	genLogger.Info("Done")
	return &cm
}

// deploymentForCN returns a Core Network Deployment object
func (r *ReconcileMosaic5g) deploymentForCN(m *mosaic5gv1alpha1.Mosaic5g) *appsv1.Deployment {
	//cnName := m.Spec.MmeDomainName
	//ls := util.LabelsForMosaic5g(m.Name + cnName)
	replicas := m.Spec.Size
	labels := make(map[string]string)
	labels["app"] = "cn"
	Annotations := make(map[string]string)
	Annotations["container.apparmor.security.beta.kubernetes.io/oaicn"] = "unconfined"
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "cn",
			Namespace:   m.Namespace,
			Labels:      labels,
			Annotations: Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           m.Spec.CNImage,
						Name:            "cn",
						Command:         []string{"/sbin/init"},
						SecurityContext: &corev1.SecurityContext{Privileged: util.NewTrue()},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "cgroup",
							ReadOnly:  true,
							MountPath: "/sys/fs/cgroup/",
						}, {
							Name:      "module",
							ReadOnly:  true,
							MountPath: "/lib/modules/",
						}, {
							Name:      "mosaic5g-config",
							MountPath: "/root/config",
						}},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
							Name:          "mosaic5g-cn",
						}},
					}},
					Affinity: util.GenAffinity("cn"),
					Volumes: []corev1.Volume{{
						Name: "cgroup",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/sys/fs/cgroup/",
								Type: util.NewHostPathType("Directory"),
							},
						}}, {
						Name: "module",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/lib/modules/",
								Type: util.NewHostPathType("Directory"),
							},
						}}, {
						Name: "mosaic5g-config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: "mosaic5g-config"},
							},
						}},
					},
				},
			},
		},
	}
	// Set Mosaic5g instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep
}

// genCNService will generate a service for oaicn
func (r *ReconcileMosaic5g) genCNService(m *mosaic5gv1alpha1.Mosaic5g) *v1.Service {
	var service *v1.Service
	selectMap := make(map[string]string)
	selectMap["app"] = "cn"
	service = &v1.Service{}
	service.Spec = v1.ServiceSpec{
		Ports: []v1.ServicePort{
			{Name: "enb", Port: 2152},
			{Name: "hss-1", Port: 3868},
			{Name: "hss-2", Port: 5868},
			{Name: "mme", Port: 2123},
			{Name: "spgw-1", Port: 3870},
			{Name: "spgw-2", Port: 5870},
		},
		Selector:  selectMap,
		ClusterIP: "None",
	}
	service.Name = "cn"
	service.Namespace = m.Namespace
	service.Labels = selectMap
	// Set Mosaic5g instance as the owner and controller
	controllerutil.SetControllerReference(m, service, r.scheme)
	return service
}

// deploymentForMySQL returns a Core Network Deployment object
func (r *ReconcileMosaic5g) deploymentForMySQL(m *mosaic5gv1alpha1.Mosaic5g) *appsv1.Deployment {
	//ls := util.LabelsForMosaic5g(m.Name + cnName)
	var replicas int32
	replicas = 1
	selectMap := make(map[string]string)
	selectMap["app"] = "oai"
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql",
			Namespace: m.Namespace,
			Labels:    selectMap,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: selectMap,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selectMap,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "mysql:latest",
						Name:  "mysql",
						Env: []corev1.EnvVar{
							{Name: "MYSQL_ROOT_PASSWORD", Value: "linux"},
						},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 3306,
							Name:          "mysql",
						}},
					}},
					Affinity: util.GenAffinity("cn"),
				},
			},
		},
	}
	// Set Mosaic5g instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep
}

// genMySQLService will generate a service so that oaicn could connect to mysql
func (r *ReconcileMosaic5g) genMySQLService(m *mosaic5gv1alpha1.Mosaic5g) *v1.Service {
	var service *v1.Service
	selectMap := make(map[string]string)
	selectMap["app"] = "oai"
	service = &v1.Service{}
	service.Spec = v1.ServiceSpec{
		Ports: []v1.ServicePort{
			{Name: "mysql", Port: 3306},
		},
		Selector:  selectMap,
		ClusterIP: "None",
	}
	service.Name = "mysql"
	service.Namespace = m.Namespace
	service.Labels = selectMap
	// Set Mosaic5g instance as the owner and controller
	controllerutil.SetControllerReference(m, service, r.scheme)
	return service
}

// deploymentForFlexRAN returns a FlexRAN Deployment object
func (r *ReconcileMosaic5g) deploymentForFlexRAN(m *mosaic5gv1alpha1.Mosaic5g) *appsv1.Deployment {
	flexRANName := m.Spec.FlexRANDomainName
	//ls := util.LabelsForMosaic5g(m.Name + cnName)
	replicas := m.Spec.Size
	labels := make(map[string]string)
	labels["app"] = "flexran"
	Annotations := make(map[string]string)
	Annotations["container.apparmor.security.beta.kubernetes.io/flexran"] = "unconfined"
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.GetName() + "-" + flexRANName,
			Namespace:   m.Namespace,
			Labels:      labels,
			Annotations: Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           m.Spec.FlexRANImage,
						Name:            "flexran",
						Command:         []string{"/sbin/init"},
						SecurityContext: &corev1.SecurityContext{Privileged: util.NewTrue()},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "cgroup",
							ReadOnly:  true,
							MountPath: "/sys/fs/cgroup/",
						}, {
							Name:      "module",
							ReadOnly:  true,
							MountPath: "/lib/modules/",
						}, {
							Name:      "mosaic5g-config",
							MountPath: "/root/config",
						}},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
							Name:          "mosaic5g-fran",
						}},
					}},
					Affinity: util.GenAffinity("flexran"),
					Volumes: []corev1.Volume{{
						Name: "cgroup",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/sys/fs/cgroup/",
								Type: util.NewHostPathType("Directory"),
							},
						}}, {
						Name: "module",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/lib/modules/",
								Type: util.NewHostPathType("Directory"),
							},
						}}, {
						Name: "mosaic5g-config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: "mosaic5g-config"},
							},
						}},
					},
				},
			},
		},
	}
	// Set Mosaic5g instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep
}

// genFlexRANService will generate a service for FlexRAN
func (r *ReconcileMosaic5g) genFlexRANService(m *mosaic5gv1alpha1.Mosaic5g) *v1.Service {
	var service *v1.Service
	selectMap := make(map[string]string)
	selectMap["app"] = "flexran"
	service = &v1.Service{}
	service.Spec = v1.ServiceSpec{
		Ports: []v1.ServicePort{
			{Name: "sbi", Port: 2210},
			{Name: "nbi", Port: 9999},
		},
		Selector:  selectMap,
		ClusterIP: "None",
	}
	service.Name = "flexran"
	service.Namespace = m.Namespace
	service.Labels = selectMap
	// Set Mosaic5g instance as the owner and controller
	controllerutil.SetControllerReference(m, service, r.scheme)
	return service
}

//deploymentForElasticsearch will deploy elasticsearch
func (r *ReconcileMosaic5g) deploymentForElasticsearch(m *mosaic5gv1alpha1.Mosaic5g) *appsv1.StatefulSet {
	labels := make(map[string]string)
	labels["app"] = "elasticsearch"
	var set *appsv1.StatefulSet
	set = &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "elasticsearch",
			Namespace: m.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: "elasticsearch",
			Replicas:    &m.Spec.Size,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:    "fix-permissions",
							Image:   "busybox",
							Command: []string{"sh", "-c", "chown -R 1000:1000 /usr/share/elasticsearch/data"},
							SecurityContext: &corev1.SecurityContext{
								Privileged: util.NewTrue(),
							},
							VolumeMounts: []corev1.VolumeMount{{
								Name:      "data",
								MountPath: "/usr/share/elasticsearch/data",
							}},
						},
						{
							Name:    "increase-vm-max-map",
							Image:   "busybox",
							Command: []string{"sysctl", "-w", "vm.max_map_count=262144"},
							SecurityContext: &corev1.SecurityContext{
								Privileged: util.NewTrue(),
							},
						},
						{
							Name:    "increase-fd-ulimit",
							Image:   "busybox",
							Command: []string{"sh", "-c", "ulimit -n 65536"},
							SecurityContext: &corev1.SecurityContext{
								Privileged: util.NewTrue(),
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "elasticsearch",
							Image: "docker.elastic.co/elasticsearch/elasticsearch:7.4.2",
							Env: []corev1.EnvVar{
								{
									Name:  "discovery.type",
									Value: "single-node",
								},
								{
									Name:  "ES_JAVA_OPTS",
									Value: "-Xms512m -Xmx512m",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "client",
									ContainerPort: 9200,
								},
								{
									Name:          "nodes",
									ContainerPort: 9300,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/usr/share/elasticsearch/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/esdata",
								},
							},
						},
					},
				},
			},
		},
	}
	// Set Elasticsearch instance as the owner and controller
	controllerutil.SetControllerReference(m, set, r.scheme)
	return set
}

//genESService create elasticsearch service for other's to use
func (r *ReconcileMosaic5g) genESService(m *mosaic5gv1alpha1.Mosaic5g) *v1.Service {
	selectMap := make(map[string]string)
	selectMap["app"] = "elasticsearch"
	var service *v1.Service
	service = &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "elasticsearch",
			Namespace: m.Namespace,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "client",
					Port: 9200,
				},
				{
					Name: "nodes",
					Port: 9300,
				},
			},
			Selector: selectMap,
		},
	}
	// Set Elasticsearch instance as the owner and controller
	controllerutil.SetControllerReference(m, service, r.scheme)
	return service
}

//deploymentForKibana will deploy kibana
func (r *ReconcileMosaic5g) deploymentForKibana(m *mosaic5gv1alpha1.Mosaic5g) *appsv1.Deployment {
	selectMap := make(map[string]string)
	selectMap["app"] = "kibana"
	var dep *appsv1.Deployment
	dep = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kibana",
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &m.Spec.Size,
			Selector: &metav1.LabelSelector{
				MatchLabels: selectMap,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selectMap,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "kibana",
							Image: "docker.elastic.co/kibana/kibana:7.4.2",
							Env: []corev1.EnvVar{
								{
									Name:  "ELASTICSEARCH_URL",
									Value: "http://elasticsearch:9200",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 5601,
								},
							},
						},
					},
				},
			},
		},
	}
	// Set Elasticsearch instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep
}

//genKibanaService create kibana service for other's to use
func (r *ReconcileMosaic5g) genKibanaService(m *mosaic5gv1alpha1.Mosaic5g) *v1.Service {
	selectMap := make(map[string]string)
	selectMap["app"] = "kibana"
	selectMaps := make(map[string]string)
	selectMaps["service"] = "kibana"
	var service *v1.Service
	service = &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kibana",
			Namespace: m.Namespace,
			Labels:    selectMaps,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port: 5601,
					Name: "webinterface",
				},
			},
			Selector: selectMap,
		},
	}
	// Set Elasticsearch instance as the owner and controller
	controllerutil.SetControllerReference(m, service, r.scheme)
	return service
}

// deploymentForDrone returns a Drone Store App Deployment object
func (r *ReconcileMosaic5g) deploymentForDrone(m *mosaic5gv1alpha1.Mosaic5g) *appsv1.Deployment {
	var replicas int32
	replicas = 1
	selectMap := make(map[string]string)
	selectMap["app"] = "store"
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "drone",
			Namespace: m.Namespace,
			Labels:    selectMap,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: selectMap,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selectMap,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "ndhfrock/store-drone:1.0",
						Name:  "drone",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8088,
							Name:          "drone",
						}},
						Command:         []string{"/sbin/init"},
						SecurityContext: &corev1.SecurityContext{Privileged: util.NewTrue()},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "cgroup",
							ReadOnly:  true,
							MountPath: "/sys/fs/cgroup/",
						}, {
							Name:      "module",
							ReadOnly:  true,
							MountPath: "/lib/modules/",
						}, {
							Name:      "mosaic5g-config",
							MountPath: "/root/config",
						}},
					}},
					Affinity: util.GenAffinity("store"),
					Volumes: []corev1.Volume{{
						Name: "cgroup",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/sys/fs/cgroup/",
								Type: util.NewHostPathType("Directory"),
							},
						}}, {
						Name: "module",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/lib/modules/",
								Type: util.NewHostPathType("Directory"),
							},
						}}, {
						Name: "mosaic5g-config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: "mosaic5g-config"},
							},
						}},
					},
				},
			},
		},
	}
	// Set Mosaic5g instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep
}

// genDroneService will generate a service so that we could access drone store app
func (r *ReconcileMosaic5g) genDroneService(m *mosaic5gv1alpha1.Mosaic5g) *v1.Service {
	var service *v1.Service
	selectMap := make(map[string]string)
	selectMap["app"] = "store"
	service = &v1.Service{}
	service.Spec = v1.ServiceSpec{
		Ports: []v1.ServicePort{
			{Name: "drone", Port: 8088},
		},
		Selector:  selectMap,
		ClusterIP: "None",
	}
	service.Name = "drone"
	service.Namespace = m.Namespace
	service.Labels = selectMap
	// Set Mosaic5g instance as the owner and controller
	controllerutil.SetControllerReference(m, service, r.scheme)
	return service
}

// deploymentForRRMKPI returns a RRMKPI Store App Deployment object
func (r *ReconcileMosaic5g) deploymentForRRMKPI(m *mosaic5gv1alpha1.Mosaic5g) *appsv1.Deployment {
	var replicas int32
	replicas = 1
	selectMap := make(map[string]string)
	selectMap["app"] = "store"
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rrmkpi",
			Namespace: m.Namespace,
			Labels:    selectMap,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: selectMap,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selectMap,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "ndhfrock/store-rrm_kpi:1.0",
						Name:  "rrmkpi",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8088,
							Name:          "rrmkpi",
						}},
						Command:         []string{"/sbin/init"},
						SecurityContext: &corev1.SecurityContext{Privileged: util.NewTrue()},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "cgroup",
							ReadOnly:  true,
							MountPath: "/sys/fs/cgroup/",
						}, {
							Name:      "module",
							ReadOnly:  true,
							MountPath: "/lib/modules/",
						}, {
							Name:      "mosaic5g-config",
							MountPath: "/root/config",
						}},
					}},
					Affinity: util.GenAffinity("store"),
					Volumes: []corev1.Volume{{
						Name: "cgroup",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/sys/fs/cgroup/",
								Type: util.NewHostPathType("Directory"),
							},
						}}, {
						Name: "module",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/lib/modules/",
								Type: util.NewHostPathType("Directory"),
							},
						}}, {
						Name: "mosaic5g-config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: "mosaic5g-config"},
							},
						}},
					},
				},
			},
		},
	}
	// Set Mosaic5g instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep
}

// genRRMKPIService will generate a service so that we could access rrmkpi store app
func (r *ReconcileMosaic5g) genRRMKPIService(m *mosaic5gv1alpha1.Mosaic5g) *v1.Service {
	var service *v1.Service
	selectMap := make(map[string]string)
	selectMap["app"] = "store"
	service = &v1.Service{}
	service.Spec = v1.ServiceSpec{
		Ports: []v1.ServicePort{
			{Name: "rrmkpi", Port: 8088},
		},
		Selector:  selectMap,
		ClusterIP: "None",
	}
	service.Name = "rrmkpi"
	service.Namespace = m.Namespace
	service.Labels = selectMap
	// Set Mosaic5g instance as the owner and controller
	controllerutil.SetControllerReference(m, service, r.scheme)
	return service
}
