package k8s

import (
	"context"
	"fmt"

	"github.com/davidliyutong/idekube-controller/internal/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// ServiceManager manages Kubernetes Services
type ServiceManager struct {
	clientset *kubernetes.Clientset
	namespace string
}

// NewServiceManager creates a new service manager
func NewServiceManager(clientset *kubernetes.Clientset, namespace string) *ServiceManager {
	return &ServiceManager{
		clientset: clientset,
		namespace: namespace,
	}
}

// CreateService creates a Kubernetes Service for a workspace
func (m *ServiceManager) CreateService(ctx context.Context, workspace *models.Workspace) (string, error) {
	serviceName := fmt.Sprintf("workspace-%d-%s", workspace.ID, workspace.UUID.String()[:8])

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: m.namespace,
			Labels: map[string]string{
				"app":           "idekube",
				"workspace-id":  fmt.Sprintf("%d", workspace.ID),
				"workspace-uuid": workspace.UUID.String(),
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app":          "idekube",
				"workspace-id": fmt.Sprintf("%d", workspace.ID),
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
				{
					Name:       "ssh",
					Protocol:   corev1.ProtocolTCP,
					Port:       22,
					TargetPort: intstr.FromInt(22),
				},
			},
		},
	}

	_, err := m.clientset.CoreV1().Services(m.namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create service: %w", err)
	}

	return serviceName, nil
}

// DeleteService deletes a Kubernetes Service
func (m *ServiceManager) DeleteService(ctx context.Context, serviceName string) error {
	err := m.clientset.CoreV1().Services(m.namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}
	return nil
}

// GetServiceEndpoint retrieves the endpoint of a Service
func (m *ServiceManager) GetServiceEndpoint(ctx context.Context, serviceName string) (string, error) {
	service, err := m.clientset.CoreV1().Services(m.namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get service: %w", err)
	}

	// For ClusterIP services, return the internal DNS name
	endpoint := fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, service.Namespace)

	// If it's a LoadBalancer service and has an external IP, use that
	if service.Spec.Type == corev1.ServiceTypeLoadBalancer && len(service.Status.LoadBalancer.Ingress) > 0 {
		ingress := service.Status.LoadBalancer.Ingress[0]
		if ingress.IP != "" {
			endpoint = ingress.IP
		} else if ingress.Hostname != "" {
			endpoint = ingress.Hostname
		}
	}

	return endpoint, nil
}
