package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"crm/utils"
)

// K8sPodInfo represents detailed status of a Pod in Kubernetes
type K8sPodInfo struct {
	Name      string    `json:"name"`
	Component string    `json:"component"`
	Status    string    `json:"status"`
	Ready     string    `json:"ready"`
	Restarts  string    `json:"restarts"`
	Age       string    `json:"age"`
	Node      string    `json:"node"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"createdAt"`
}

// K8sHandler handles Kubernetes status and pod crash simulation
type K8sHandler struct{}

// NewK8sHandler creates a new K8sHandler
func NewK8sHandler() *K8sHandler {
	return &K8sHandler{}
}

// GetStatus handles GET /api/admin/k8s/status — fetches real-time pod metrics via kubectl
func (h *K8sHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Execute kubectl get pods -o json
	cmd := exec.CommandContext(ctx, "kubectl", "get", "pods", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		utils.Success(w, map[string]interface{}{
			"connected": false,
			"error":     "Failed to execute kubectl or K8s not reachable",
			"pods":      []K8sPodInfo{},
		})
		return
	}

	var k8sOutput struct {
		Items []struct {
			Metadata struct {
				Name              string            `json:"name"`
				CreationTimestamp time.Time         `json:"creationTimestamp"`
				Labels            map[string]string `json:"labels"`
			} `json:"metadata"`
			Spec struct {
				NodeName string `json:"nodeName"`
			} `json:"spec"`
			Status struct {
				Phase             string `json:"phase"`
				PodIP             string `json:"podIP"`
				ContainerStatuses []struct {
					Ready        bool `json:"ready"`
					RestartCount int  `json:"restartCount"`
				} `json:"containerStatuses"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &k8sOutput); err != nil {
		utils.InternalServerError(w, "Failed to parse Kubernetes cluster JSON")
		return
	}

	pods := make([]K8sPodInfo, 0, len(k8sOutput.Items))
	for _, item := range k8sOutput.Items {
		readyStr := "0/1"
		restartsStr := "0"
		if len(item.Status.ContainerStatuses) > 0 {
			if item.Status.ContainerStatuses[0].Ready {
				readyStr = "1/1"
			}
			restartsStr = string(rune(item.Status.ContainerStatuses[0].RestartCount + '0'))
		}

		comp := item.Metadata.Labels["app"]
		if comp == "" {
			comp = "Application"
		}

		pods = append(pods, K8sPodInfo{
			Name:      item.Metadata.Name,
			Component: comp,
			Status:    item.Status.Phase,
			Ready:     readyStr,
			Restarts:  restartsStr,
			Age:       time.Since(item.Metadata.CreationTimestamp).Round(time.Second).String(),
			Node:      item.Spec.NodeName,
			IP:        item.Status.PodIP,
			CreatedAt: item.Metadata.CreationTimestamp,
		})
	}

	utils.Success(w, map[string]interface{}{
		"connected":  true,
		"cluster":    "Kubernetes (Docker Desktop)",
		"pod_count":  len(pods),
		"pods":       pods,
		"updated_at": time.Now().Format(time.RFC3339),
	})
}

// KillPod handles POST /api/admin/k8s/kill-pod — force-deletes a pod to demonstrate K8s auto-healing
func (h *K8sHandler) KillPod(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PodName string `json:"podName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.PodName) == "" {
		utils.BadRequest(w, "Pod name is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Execute kubectl delete pod <pod-name> --grace-period=0 --force
	cmd := exec.CommandContext(ctx, "kubectl", "delete", "pod", req.PodName, "--grace-period=0", "--force")
	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.InternalServerError(w, "Failed to kill pod: "+string(output))
		return
	}

	utils.Success(w, map[string]string{
		"message": "Pod " + req.PodName + " was killed! Watch Kubernetes recreate it in real-time.",
		"pod":     req.PodName,
	})
}
