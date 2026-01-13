package queue

import (
	"time"

	"github.com/davidliyutong/idekube-controller/internal/models"
)

// Event types for workspace operations
const (
	EventTypeWorkspaceCreate = "workspace.create"
	EventTypeWorkspaceUpdate = "workspace.update"
	EventTypeWorkspaceDelete = "workspace.delete"
	EventTypeWorkspaceStart  = "workspace.start"
	EventTypeWorkspaceStop   = "workspace.stop"
	
	EventTypeVolumeCreate = "volume.create"
	EventTypeVolumeDelete = "volume.delete"
	EventTypeVolumeResize = "volume.resize"
	
	EventTypeStatusChanged = "workspace.status.changed"
	EventTypeTimeout       = "workspace.timeout"
	EventTypeK8SError      = "workspace.k8s.error"
)

// Exchange and Queue names
const (
	ExchangeIDEKube         = "idekube.events"
	ExchangeType            = "topic"
	
	QueueHousekeeperWorkspace = "housekeeper.workspace.events"
	QueueHousekeeperVolume    = "housekeeper.volume.events"
	QueueControllerStatus     = "controller.status.events"
	
	// Routing key patterns
	RoutingKeyWorkspace = "workspace.*"
	RoutingKeyVolume    = "volume.*"
	RoutingKeyStatus    = "workspace.status.*"
)

// WorkspaceEvent represents a workspace lifecycle event
type WorkspaceEvent struct {
	Type        string             `json:"type"`
	WorkspaceID int64              `json:"workspace_id"`
	Workspace   *models.Workspace  `json:"workspace,omitempty"`
	Template    *models.Template   `json:"template,omitempty"`
	Volumes     []*models.Volume   `json:"volumes,omitempty"`
	Timestamp   time.Time          `json:"timestamp"`
	
	// Additional context
	UserID       int64  `json:"user_id,omitempty"`
	Reason       string `json:"reason,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// VolumeEvent represents a volume lifecycle event
type VolumeEvent struct {
	Type      string          `json:"type"`
	VolumeID  int64           `json:"volume_id"`
	Volume    *models.Volume  `json:"volume,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	
	// Additional context
	UserID       int64  `json:"user_id,omitempty"`
	Reason       string `json:"reason,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// StatusChangeEvent represents a workspace status change event (HouseKeeper -> Controller)
type StatusChangeEvent struct {
	Type         string                    `json:"type"`
	WorkspaceID  int64                     `json:"workspace_id"`
	OldStatus    models.WorkspaceStatus    `json:"old_status"`
	NewStatus    models.WorkspaceStatus    `json:"new_status"`
	Reason       string                    `json:"reason,omitempty"`
	ErrorMessage string                    `json:"error_message,omitempty"`
	Timestamp    time.Time                 `json:"timestamp"`
	
	// K8S resource information
	K8SResourceStatus *K8SResourceStatus `json:"k8s_resource_status,omitempty"`
}

// K8SResourceStatus contains detailed K8S resource information
type K8SResourceStatus struct {
	DeploymentReady bool              `json:"deployment_ready"`
	PodsReady       int32             `json:"pods_ready"`
	PodsTotal       int32             `json:"pods_total"`
	ServiceIP       string            `json:"service_ip,omitempty"`
	Conditions      []ResourceCondition `json:"conditions,omitempty"`
}

// ResourceCondition represents a K8S resource condition
type ResourceCondition struct {
	Type    string    `json:"type"`
	Status  string    `json:"status"`
	Reason  string    `json:"reason,omitempty"`
	Message string    `json:"message,omitempty"`
	LastTransitionTime time.Time `json:"last_transition_time"`
}

// NewWorkspaceEvent creates a new workspace event
func NewWorkspaceEvent(eventType string, workspace *models.Workspace, template *models.Template, volumes []*models.Volume) *WorkspaceEvent {
	return &WorkspaceEvent{
		Type:        eventType,
		WorkspaceID: workspace.ID,
		Workspace:   workspace,
		Template:    template,
		Volumes:     volumes,
		Timestamp:   time.Now(),
	}
}

// NewVolumeEvent creates a new volume event
func NewVolumeEvent(eventType string, volume *models.Volume) *VolumeEvent {
	return &VolumeEvent{
		Type:      eventType,
		VolumeID:  volume.ID,
		Volume:    volume,
		Timestamp: time.Now(),
	}
}

// NewStatusChangeEvent creates a new status change event
func NewStatusChangeEvent(workspaceID int64, oldStatus, newStatus models.WorkspaceStatus, reason string) *StatusChangeEvent {
	return &StatusChangeEvent{
		Type:        EventTypeStatusChanged,
		WorkspaceID: workspaceID,
		OldStatus:   oldStatus,
		NewStatus:   newStatus,
		Reason:      reason,
		Timestamp:   time.Now(),
	}
}
