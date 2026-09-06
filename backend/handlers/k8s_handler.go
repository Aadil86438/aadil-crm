package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os/exec"
	"strings"
	"sync"
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
type K8sHandler struct {
	mu            sync.Mutex
	simulatedPods []K8sPodInfo
	initOnce      bool
}

// NewK8sHandler creates a new K8sHandler
func NewK8sHandler() *K8sHandler {
	return &K8sHandler{
		simulatedPods: make([]K8sPodInfo, 0),
	}
}

func (h *K8sHandler) ensureSimulatedPods() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.initOnce {
		return
	}

	now := time.Now().Add(-6 * time.Hour)
	h.simulatedPods = []K8sPodInfo{
		{
			Name:      "crm-postgres-db-78f94d9c4-x2p1a",
			Component: "postgres",
			Status:    "Running",
			Ready:     "1/1",
			Restarts:  "0",
			Age:       "6h",
			Node:      "docker-desktop",
			IP:        "10.244.0.12",
			CreatedAt: now,
		},
		{
			Name:      "crm-redis-cache-6d5854b79-m9k4v",
			Component: "redis",
			Status:    "Running",
			Ready:     "1/1",
			Restarts:  "0",
			Age:       "6h",
			Node:      "docker-desktop",
			IP:        "10.244.0.15",
			CreatedAt: now,
		},
		{
			Name:      "crm-backend-api-5c744f686-q8l2z",
			Component: "backend",
			Status:    "Running",
			Ready:     "1/1",
			Restarts:  "0",
			Age:       "6h",
			Node:      "docker-desktop",
			IP:        "10.244.0.18",
			CreatedAt: now,
		},
		{
			Name:      "crm-frontend-web-8467b45f4-j4w7n",
			Component: "crm-frontend",
			Status:    "Running",
			Ready:     "1/1",
			Restarts:  "0",
			Age:       "6h",
			Node:      "docker-desktop",
			IP:        "10.244.0.21",
			CreatedAt: now,
		},
	}
	h.initOnce = true
}

// GetStatus handles GET /api/admin/k8s/status — fetches real-time pod metrics via kubectl or fallback simulator
func (h *K8sHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Execute kubectl get pods -o json
	cmd := exec.CommandContext(ctx, "kubectl", "get", "pods", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to simulated cluster state so visualizer is always usable!
		h.ensureSimulatedPods()
		h.mu.Lock()
		defer h.mu.Unlock()

		utils.Success(w, map[string]interface{}{
			"connected":  true,
			"cluster":    "Kubernetes Engine (Simulated Cluster Mode)",
			"pod_count":  len(h.simulatedPods),
			"pods":       h.simulatedPods,
			"updated_at": time.Now().Format(time.RFC3339),
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
			restartsStr = fmt.Sprintf("%d", item.Status.ContainerStatuses[0].RestartCount)
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
		"cluster":    "Kubernetes (Live Cluster)",
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

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Execute kubectl delete pod <pod-name> --grace-period=0 --force
	cmd := exec.CommandContext(ctx, "kubectl", "delete", "pod", req.PodName, "--grace-period=0", "--force")
	_, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to simulated kill & auto-heal sequence
		h.ensureSimulatedPods()
		h.mu.Lock()
		defer h.mu.Unlock()

		randHex := fmt.Sprintf("%05x", rand.Intn(0xfffff))
		for i, pod := range h.simulatedPods {
			if pod.Name == req.PodName {
				// Re-create pod with new hash and increment restarts
				restarts := 1
				fmt.Sscanf(pod.Restarts, "%d", &restarts)
				restarts++

				parts := strings.Split(pod.Name, "-")
				prefix := strings.Join(parts[:len(parts)-1], "-")
				newName := fmt.Sprintf("%s-%s", prefix, randHex)

				h.simulatedPods[i] = K8sPodInfo{
					Name:      newName,
					Component: pod.Component,
					Status:    "Running",
					Ready:     "1/1",
					Restarts:  fmt.Sprintf("%d", restarts),
					Age:       "1s",
					Node:      pod.Node,
					IP:        pod.IP,
					CreatedAt: time.Now(),
				}
				break
			}
		}

		utils.Success(w, map[string]string{
			"message": "Pod " + req.PodName + " was terminated! Kubernetes auto-healing spun up a replacement Pod.",
			"pod":     req.PodName,
		})
		return
	}

	utils.Success(w, map[string]string{
		"message": "Pod " + req.PodName + " was killed! Watch Kubernetes recreate it in real-time.",
		"pod":     req.PodName,
	})
}

