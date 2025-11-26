package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"queue-manager/internal/repository"
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

func RegisterRoutes(r *gin.Engine, repo *repository.Repository) {
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, APIResponse{Success: true, Data: map[string]string{"status": "ok"}})
	})

	r.GET("/services/:service_name/queues", getServiceQueues(repo))
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


