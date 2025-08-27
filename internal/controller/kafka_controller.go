package controller

import (
	"context"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kafiv1 "github.com/haunshila/kafi/api/v1"
)

// KafkaReconciler reconciles a Kafka object
type KafkaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kafi.haunshila.com,resources=kafkas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kafi.haunshila.com,resources=kafkas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *KafkaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the Kafka instance
	kafka := &kafiv1.Kafka{}
	err := r.Get(ctx, req.NamespacedName, kafka)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted.
			// Return and don't reconcile.
			log.Info("Kafka resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Kafka resource")
		return ctrl.Result{}, err
	}

	// Define a headless Service for the StatefulSet
	// This service provides stable network identities for each broker
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kafka.Name + "-headless",
			Namespace: kafka.Namespace,
			Labels:    labelsForKafka(kafka.Name),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       9092, // Kafka broker port
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(9092),
					Name:       "kafka",
				},
			},
			ClusterIP: corev1.ClusterIPNone, // Headless Service
			Selector:  labelsForKafka(kafka.Name),
		},
	}

	// Set Kafka instance as the owner of the Service
	if err := ctrl.SetControllerReference(kafka, service, r.Scheme); err != nil {
		log.Error(err, "Failed to set owner reference on Service")
		return ctrl.Result{}, err
	}

	// Check if the Service already exists
	foundService := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{Name: service.Name, Namespace: service.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new headless Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		if err := r.Create(ctx, service); err != nil {
			log.Error(err, "Failed to create new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}

	// Define the StatefulSet for the Kafka brokers
	// This uses the volumeClaimTemplates for persistent storage
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kafka.Name,
			Namespace: kafka.Namespace,
			Labels:    labelsForKafka(kafka.Name),
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: kafka.Name + "-headless",
			Replicas:    &kafka.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsForKafka(kafka.Name),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelsForKafka(kafka.Name),
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: kafka.Spec.ImagePullSecrets, // Use from CR spec					// Use from CR spec
					Containers: []corev1.Container{
						{
							Name:  "kafka",
							Image: "bitnami/kafka:" + kafka.Spec.Version,
							Ports: []corev1.ContainerPort{
								{ContainerPort: 9092, Name: "kafka"},
							},
							Env: []corev1.EnvVar{
								{Name: "KAFKA_CFG_ZOOKEEPER_CONNECT", Value: "zookeeper-service:2181"},
								{Name: "KAFKA_CFG_LISTENERS", Value: "PLAINTEXT://:9092"},
								{Name: "KAFKA_CFG_ADVERTISED_LISTENERS", Value: "PLAINTEXT://$(POD_NAME).$(SERVICE_NAME).$(NAMESPACE).svc.cluster.local:9092"},
								{Name: "KAFKA_KAFKA_JMX_OPTS", Value: "-Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.port=9999"},
								{Name: "KAFKA_CFG_BROKER_ID", ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										FieldPath: "metadata.annotations['statefulset.kubernetes.io/pod-name']",
									},
								}},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/bitnami/kafka",
								},
							},
						},
					},
					// Optional: Use a sidecar container for JMX monitoring
					// If you use a sidecar, you would also need to configure the service to expose the JMX port.
				},
			},
			// Define the volume claim template for Ceph storage
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(kafka.Spec.Storage.Size),
							},
						},
						StorageClassName: &kafka.Spec.Storage.StorageClassName,
					},
				},
			},
		},
	}

	// Set Kafka instance as the owner of the StatefulSet
	if err := ctrl.SetControllerReference(kafka, sts, r.Scheme); err != nil {
		log.Error(err, "Failed to set owner reference on StatefulSet")
		return ctrl.Result{}, err
	}

	// Check if the StatefulSet already exists
	foundSts := &appsv1.StatefulSet{}
	err = r.Get(ctx, client.ObjectKey{Name: sts.Name, Namespace: sts.Namespace}, foundSts)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new StatefulSet", "StatefulSet.Namespace", sts.Namespace, "StatefulSet.Name", sts.Name)
		if err = r.Create(ctx, sts); err != nil {
			log.Error(err, "Failed to create new StatefulSet", "StatefulSet.Namespace", sts.Namespace, "StatefulSet.Name", sts.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get StatefulSet")
		return ctrl.Result{}, err
	}

	// Update the Kafka CR's status
	kafka.Status.Replicas = foundSts.Status.ReadyReplicas
	kafka.Status.BootstrapServers = getBootstrapServers(foundSts)
	err = r.Status().Update(ctx, kafka)
	if err != nil {
		log.Error(err, "Failed to update Kafka status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// labelsForKafka returns the labels for the Kafka cluster
func labelsForKafka(name string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "kafka",
		"app.kubernetes.io/instance":   name,
		"app.kubernetes.io/created-by": "kafi-operator",
	}
}

// getBootstrapServers generates the bootstrap server address from the StatefulSet status
func getBootstrapServers(sts *appsv1.StatefulSet) string {
	// A more robust implementation would check for readiness before returning.
	// This is a simple example for demonstration purposes.

	// Format: <statefulset-name>-<0>.<service-name>.<namespace>.svc.cluster.local:9092
	// For a 3-replica cluster, this would be:
	// <sts.Name>-0.<sts.Spec.ServiceName>.<sts.Namespace>.svc.cluster.local:9092
	// and so on for each replica.

	bootstrapServers := ""
	for i := int32(0); i < *sts.Spec.Replicas; i++ {
		if bootstrapServers != "" {
			bootstrapServers += ","
		}
		// Corrected line: remove the metav1. prefix
		bootstrapServers += sts.Name + "-" + strconv.Itoa(int(i)) + "." + sts.Spec.ServiceName + "." + sts.Namespace + ".svc.cluster.local:9092"
	}
	return bootstrapServers
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kafiv1.Kafka{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
