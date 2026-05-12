package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/svetilka/EffectiveMobileProject/internal/models"
	"github.com/svetilka/EffectiveMobileProject/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type SubscriptionHandler struct {
	repo *repository.SubscriptionRepository
}

func NewSubscriptionHandler(repo *repository.SubscriptionRepository) *SubscriptionHandler {
	return &SubscriptionHandler{repo: repo}
}

// CreateSubscription создает новую подписку
// @Summary Create a new subscription
// @Description Create a new subscription record
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var req models.CreateSubscriptionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Warn("Invalid request body for create subscription")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse UserID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		log.WithError(err).Warn("Invalid user_id format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format"})
		return
	}

	// Parse start date
	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		log.WithError(err).Warn("Invalid start_date format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use MM-YYYY"})
		return
	}

	// Parse end date if provided
	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		parsedEnd, err := time.Parse("01-2006", *req.EndDate)
		if err != nil {
			log.WithError(err).Warn("Invalid end_date format")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use MM-YYYY"})
			return
		}
		endDate = &parsedEnd
	}

	subscription := &models.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if err := h.repo.Create(subscription); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

// GetSubscription возвращает подписку по ID
// @Summary Get subscription by ID
// @Description Get a single subscription by its ID
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 200 {object} models.Subscription
// @Failure 404 {object} map[string]interface{}
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	subscription, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// ListSubscriptions возвращает список подписок с фильтрацией
// @Summary List subscriptions
// @Description Get a list of subscriptions with optional filtering
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param service_name query string false "Filter by service name"
// @Param limit query int false "Page limit" default(10)
// @Param offset query int false "Page offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	filter := models.SubscriptionFilter{
		UserID:      c.Query("user_id"),
		ServiceName: c.Query("service_name"),
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	subscriptions, total, err := h.repo.List(filter, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   subscriptions,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateSubscription обновляет существующую подписку
// @Summary Update subscription
// @Description Update an existing subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "Subscription ID"
// @Param subscription body models.UpdateSubscriptionRequest true "Update data"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req models.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})

	if req.ServiceName != nil {
		updates["service_name"] = *req.ServiceName
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.StartDate != nil {
		startDate, err := time.Parse("01-2006", *req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format"})
			return
		}
		updates["start_date"] = startDate
	}
	if req.EndDate != nil {
		var endDate *time.Time
		if *req.EndDate != "" {
			parsedEnd, err := time.Parse("01-2006", *req.EndDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format"})
				return
			}
			endDate = &parsedEnd
		}
		updates["end_date"] = endDate
	}

	if err := h.repo.Update(uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription"})
		return
	}

	subscription, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found after update"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// DeleteSubscription удаляет подписку
// @Summary Delete subscription
// @Description Delete a subscription by ID
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 204 {object} nil
// @Failure 404 {object} map[string]interface{}
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// CalculateTotalCost рассчитывает суммарную стоимость подписок за период
// @Summary Calculate total cost
// @Description Calculate total cost of subscriptions for a given period with optional filters
// @Tags subscriptions
// @Produce json
// @Param start_date query string true "Start date (MM-YYYY)"
// @Param end_date query string true "End date (MM-YYYY)"
// @Param user_id query string false "Filter by user ID"
// @Param service_name query string false "Filter by service name"
// @Success 200 {object} models.TotalCostResponse
// @Failure 400 {object} map[string]interface{}
// @Router /subscriptions/total-cost [get]
func (h *SubscriptionHandler) CalculateTotalCost(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	userID := c.Query("user_id")
	serviceName := c.Query("service_name")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	startDate, err := time.Parse("01-2006", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use MM-YYYY"})
		return
	}

	endDate, err := time.Parse("01-2006", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use MM-YYYY"})
		return
	}

	// Ensure start date is before end date
	if startDate.After(endDate) {
		startDate, endDate = endDate, startDate
	}

	// Set to first day of month for start and last day for end
	startDate = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate = time.Date(endDate.Year(), endDate.Month()+1, 0, 23, 59, 59, 0, time.UTC)

	totalCost, err := h.repo.CalculateTotalCost(startDate, endDate, userID, serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate total cost"})
		return
	}

	c.JSON(http.StatusOK, models.TotalCostResponse{TotalCost: totalCost})
}
