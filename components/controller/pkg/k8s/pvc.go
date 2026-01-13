package k8s

import (
	"context"
	"fmt"

	"github.com/davidliyutong/idekube-controller/internal/models"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PVCManager manages Kubernetes PersistentVolumeClaims
type PVCManager struct {
	clientset *kubernetes.Clientset
	namespace string
}

// NewPVCManager creates a new PVC manager
func NewPVCManager(clientset *kubernetes.Clientset, namespace string) *PVCManager {
	return &PVCManager{
		clientset: clientset,
		namespace: namespace,
	}
}

// CreatePVC creates a PersistentVolumeClaim for a volume
func (m *PVCManager) CreatePVC(ctx context.Context, volume *models.Volume) (string, error) {
	pvcName := fmt.Sprintf("volume-%d-%s", volume.ID, volume.UUID.String()[:8])

	// Convert access mode
	accessMode := corev1.ReadWriteOnce
	switch volume.AccessMode {
	case models.VolumeAccessModeReadWriteOnce:
		accessMode = corev1.ReadWriteOnce
	case models.VolumeAccessModeReadWriteMany:
		accessMode = corev1.ReadWriteMany
	case models.VolumeAccessModeReadOnlyMany:
		accessMode = corev1.ReadOnlyMany
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: m.namespace,
			Labels: map[string]string{
				"app":              "idekube",
				"volume-id":        fmt.Sprintf("%d", volume.ID),
				"volume-uuid":      volume.UUID.String(),
				"owner-type":       string(volume.OwnerType),
				"owner-id":         fmt.Sprintf("%d", volume.OwnerID),
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{accessMode},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(fmt.Sprintf("%dMi", volume.SizeMB)),
				},
			},
		},
	}

	// Set storage class if specified
	if volume.StorageClass != nil && *volume.StorageClass != "" {
		pvc.Spec.StorageClassName = volume.StorageClass
	}

	_, err := m.clientset.CoreV1().PersistentVolumeClaims(m.namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create PVC: %w", err)
	}

	return pvcName, nil
}

// DeletePVC deletes a PersistentVolumeClaim
func (m *PVCManager) DeletePVC(ctx context.Context, pvcName string) error {
	err := m.clientset.CoreV1().PersistentVolumeClaims(m.namespace).Delete(ctx, pvcName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete PVC: %w", err)
	}
	return nil
}

// GetPVCStatus retrieves the status of a PersistentVolumeClaim
func (m *PVCManager) GetPVCStatus(ctx context.Context, pvcName string) (models.VolumeStatus, error) {
	pvc, err := m.clientset.CoreV1().PersistentVolumeClaims(m.namespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err != nil {
		return models.VolumeStatusFailed, fmt.Errorf("failed to get PVC status: %w", err)
	}

	switch pvc.Status.Phase {
	case corev1.ClaimPending:
		return models.VolumeStatusPending, nil
	case corev1.ClaimBound:
		return models.VolumeStatusBound, nil
	case corev1.ClaimLost:
		return models.VolumeStatusFailed, nil
	default:
		return models.VolumeStatusPending, nil
	}
}

// ResizePVC resizes a PersistentVolumeClaim
func (m *PVCManager) ResizePVC(ctx context.Context, pvcName string, newSizeMB int) error {
	pvc, err := m.clientset.CoreV1().PersistentVolumeClaims(m.namespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get PVC: %w", err)
	}

	// Update storage request
	pvc.Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse(fmt.Sprintf("%dMi", newSizeMB))

	_, err = m.clientset.CoreV1().PersistentVolumeClaims(m.namespace).Update(ctx, pvc, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to resize PVC: %w", err)
	}

	return nil
}
