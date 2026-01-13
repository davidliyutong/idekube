package k8s

import (
	"context"
	"fmt"

	"github.com/davidliyutong/idekube-controller/internal/models"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DeploymentManager manages Kubernetes Deployments
type DeploymentManager struct {
	clientset *kubernetes.Clientset
	namespace string
}

// NewDeploymentManager creates a new deployment manager
func NewDeploymentManager(clientset *kubernetes.Clientset, namespace string) *DeploymentManager {
	return &DeploymentManager{
		clientset: clientset,
		namespace: namespace,
	}
}

// CreateDeployment creates a Kubernetes Deployment for a workspace
func (m *DeploymentManager) CreateDeployment(ctx context.Context, workspace *models.Workspace, template *models.Template, volumes []*models.Volume) (string, error) {
	deploymentName := fmt.Sprintf("workspace-%d-%s", workspace.ID, workspace.UUID.String()[:8])

	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: m.namespace,
			Labels: map[string]string{
				"app":           "idekube",
				"workspace-id":  fmt.Sprintf("%d", workspace.ID),
				"workspace-uuid": workspace.UUID.String(),
				"owner-type":    string(workspace.OwnerType),
				"owner-id":      fmt.Sprintf("%d", workspace.OwnerID),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":          "idekube",
					"workspace-id": fmt.Sprintf("%d", workspace.ID),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":          "idekube",
						"workspace-id": fmt.Sprintf("%d", workspace.ID),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "workspace",
							Image: template.ImageRef,
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", workspace.CPUMillicores)),
									corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", workspace.MemoryMB)),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", workspace.CPUMillicores)),
									corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", workspace.MemoryMB)),
								},
							},
						},
					},
				},
			},
		},
	}

	// Note: GPU support removed as Workspace model doesn't include GPU field
	// Add it back when GPU field is added to Workspace model

	// Add volume mounts
	if len(volumes) > 0 {
		for _, vol := range volumes {
			if vol.PVCName == nil {
				continue
			}

			volumeMount := corev1.VolumeMount{
				Name:      fmt.Sprintf("vol-%d", vol.ID),
				MountPath: fmt.Sprintf("/workspace/volumes/%s", vol.Name),
			}
			deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(
				deployment.Spec.Template.Spec.Containers[0].VolumeMounts,
				volumeMount,
			)

			volume := corev1.Volume{
				Name: fmt.Sprintf("vol-%d", vol.ID),
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: *vol.PVCName,
					},
				},
			}
			deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volume)
		}
	}

	_, err := m.clientset.AppsV1().Deployments(m.namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create deployment: %w", err)
	}

	return deploymentName, nil
}

// DeleteDeployment deletes a Kubernetes Deployment
func (m *DeploymentManager) DeleteDeployment(ctx context.Context, deploymentName string) error {
	err := m.clientset.AppsV1().Deployments(m.namespace).Delete(ctx, deploymentName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}
	return nil
}

// GetDeploymentStatus retrieves the status of a Deployment
func (m *DeploymentManager) GetDeploymentStatus(ctx context.Context, deploymentName string) (models.WorkspaceStatus, error) {
	deployment, err := m.clientset.AppsV1().Deployments(m.namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return models.WorkspaceStatusFailed, fmt.Errorf("failed to get deployment: %w", err)
	}

	if deployment.Status.ReadyReplicas > 0 {
		return models.WorkspaceStatusRunning, nil
	} else if deployment.Status.Replicas > 0 {
		return models.WorkspaceStatusStarting, nil
	} else {
		return models.WorkspaceStatusStopped, nil
	}
}

// ScaleDeployment scales a Deployment to the specified number of replicas
func (m *DeploymentManager) ScaleDeployment(ctx context.Context, deploymentName string, replicas int32) error {
	deployment, err := m.clientset.AppsV1().Deployments(m.namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	deployment.Spec.Replicas = &replicas

	_, err = m.clientset.AppsV1().Deployments(m.namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}

	return nil
}

// UpdateDeploymentResources updates CPU and memory resources of a Deployment
func (m *DeploymentManager) UpdateDeploymentResources(ctx context.Context, deploymentName string, cpuMillicores, memoryMB int) error {
	deployment, err := m.clientset.AppsV1().Deployments(m.namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	deployment.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU] = resource.MustParse(fmt.Sprintf("%dm", cpuMillicores))
	deployment.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory] = resource.MustParse(fmt.Sprintf("%dMi", memoryMB))
	deployment.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU] = resource.MustParse(fmt.Sprintf("%dm", cpuMillicores))
	deployment.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] = resource.MustParse(fmt.Sprintf("%dMi", memoryMB))

	_, err = m.clientset.AppsV1().Deployments(m.namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update deployment resources: %w", err)
	}

	return nil
}
