package api

import (
	"net/http"
	"strconv"

	"queue-manager/internal/queue"
	"queue-manager/internal/reconciliation"
	"queue-manager/internal/repository"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// QueueDetailResponse represents a queue with its assignment details for API responses
type QueueDetailResponse struct {
	QueueName     string                 `json:"queue_name"`
	UUID          string                 `json:"uuid"`
	Durable       bool                   `json:"durable"`
	AutoDelete    bool                   `json:"auto_delete"`
	Arguments     map[string]interface{} `json:"arguments"`
	Description   string                 `json:"description"`
	PrefetchCount int                    `json:"prefetch_count"`
	MaxInflight   int                    `json:"max_inflight"`
	Notes         string                 `json:"notes"`
	Meta          map[string]interface{} `json:"meta"`
}

func RegisterRoutes(r *gin.Engine, repo *repository.Repository, qp queue.Provider) {
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, APIResponse{Success: true, Data: map[string]string{"status": "ok"}})
	})

	r.GET("/services/:service_name/queues", getServiceQueues(repo))
	r.POST("/sync", syncTopology(repo, qp))
}

// getServiceQueues returns all queues assigned to a service
func getServiceQueues(repo *repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if repo == nil {
			c.JSON(http.StatusServiceUnavailable, APIResponse{
				Success: false,
				Error: &APIError{
					Code:    "SERVICE_UNAVAILABLE",
					Message: "Database connection not available",
				},
			})
			return
		}

		serviceName := c.Param("service_name")
		if serviceName == "" {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error: &APIError{
					Code:    "INVALID_PARAMETER",
					Message: "service_name parameter is required",
				},
			})
			return
		}

		queues, err := repo.GetQueuesByServiceName(serviceName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Error: &APIError{
					Code:    "DATABASE_ERROR",
					Message: "Failed to retrieve queues for service",
				},
			})
			return
		}

		// Convert to response format
		queueDetails := make([]QueueDetailResponse, len(queues))
		for i, qwa := range queues {
			queueDetails[i] = QueueDetailResponse{
				QueueName:     qwa.Queue.QueueName,
				UUID:          qwa.Queue.UUID,
				Durable:       qwa.Queue.Durable,
				AutoDelete:    qwa.Queue.AutoDelete,
				Arguments:     qwa.Queue.Arguments,
				Description:   qwa.Queue.Description,
				PrefetchCount: qwa.PrefetchCount,
				MaxInflight:   qwa.MaxInflight,
				Notes:         qwa.Notes,
				Meta:          qwa.AssignmentMeta,
			}
		}

		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Data:    queueDetails,
		})
	}
}

// syncTopology handles the POST /sync endpoint for manual synchronization
func syncTopology(repo *repository.Repository, qp queue.Provider) gin.HandlerFunc {
	return func(c *gin.Context) {
		if qp == nil {
			c.JSON(http.StatusServiceUnavailable, APIResponse{
				Success: false,
				Error: &APIError{
					Code:    "SERVICE_UNAVAILABLE",
					Message: "Queue provider not available",
				},
			})
			return
		}

		if repo == nil {
			c.JSON(http.StatusServiceUnavailable, APIResponse{
				Success: false,
				Error: &APIError{
					Code:    "SERVICE_UNAVAILABLE",
					Message: "Database connection not available",
				},
			})
			return
		}

		// Parse dryRun query parameter
		dryRun := false
		if dryRunStr := c.Query("dryRun"); dryRunStr != "" {
			var err error
			dryRun, err = strconv.ParseBool(dryRunStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, APIResponse{
					Success: false,
					Error: &APIError{
						Code:    "INVALID_PARAMETER",
						Message: "dryRun must be a boolean value",
					},
				})
				return
			}
		}

		// Perform reconciliation
		result, err := reconciliation.ReconcileTopology(qp, repo, dryRun)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Error: &APIError{
					Code:    "RECONCILIATION_ERROR",
					Message: err.Error(),
				},
			})
			return
		}

		// Format response according to API spec
		responseData := map[string]interface{}{
			"actions": map[string]interface{}{
				"toCreate": map[string]interface{}{
					"exchanges": result.CreatedExchanges,
					"queues":    result.CreatedQueues,
					"bindings":  result.CreatedBindings,
				},
				"toDelete": map[string]interface{}{
					"exchanges": result.DeletedExchanges,
					"queues":    result.DeletedQueues,
					"bindings":  result.DeletedBindings,
				},
			},
		}

		response := APIResponse{
			Success: true,
			Data:    responseData,
		}

		if dryRun {
			c.JSON(http.StatusOK, response)
		} else {
			// For actual sync, return 200 OK with results
			c.JSON(http.StatusOK, response)
		}
	}
}
