package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/svetilka/EffectiveMobileProject/internal/models"
	"github.com/svetilka/EffectiveMobileProject/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo *repository.SubscriptionRepository
	log  *logrus.Logger
}

func NewHandler(repo *repository.SubscriptionRepository) *Handler {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	return &Handler{
		repo: repo,
		log:  log,
	}
}

// CreateSubscription godoc
// @Summary Create a new subscription
// @Description Create a new subscription record
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.Subscription true "Subscription object"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [post]
func (h *Handler) CreateSubscription(c *gin.Context) {
	var sub models.Subscription
	if err := c.ShouldBindJSON(&sub); err != nil {
		h.log.WithError(err).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Create(&sub); err != nil {
		h.log.WithError(err).Error("Failed to create subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
		return
	}

	h.log.WithField("id", sub.ID).Info("Subscription created successfully")
	c.JSON(http.StatusCreated, sub)
}

// GetSubscription godoc
// @Summary Get subscription by ID
// @Description Get a subscription by its uint ID
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 200 {object} models.Subscription
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [get]
func (h *Handler) GetSubscription(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 0 {
		h.log.WithError(err).Error("Invalid ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	sub, err := h.repo.GetByID(uint(id))
	if err != nil {
		h.log.WithError(err).WithField("id", id).Error("Subscription not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// UpdateSubscription godoc
// @Summary Update a subscription
// @Description Update an existing subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "Subscription ID"
// @Param subscription body models.Subscription true "Subscription object"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [put]
func (h *Handler) UpdateSubscription(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	var sub models.Subscription
	if err := c.ShouldBindJSON(&sub); err != nil {
		h.log.WithError(err).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updateData map[string]interface{}
	subBytes, err := json.Marshal(sub)
	if err != nil {
		h.log.WithError(err).Error("Failed to marshal subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	if err := json.Unmarshal(subBytes, &updateData); err != nil {
		h.log.WithError(err).Error("Failed to unmarshal subscription to map")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Передаем полученный map в метод Update
	if err := h.repo.Update(uint(id), updateData); err != nil {
		h.log.WithError(err).WithField("id", id).Error("Failed to update subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	h.log.WithField("id", id).Info("Subscription updated successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Subscription updated successfully"})
}

// DeleteSubscription godoc
// @Summary Delete a subscription
// @Description Delete a subscription by ID
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 204 {object} nil
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [delete]
func (h *Handler) DeleteSubscription(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		h.log.WithError(err).WithField("id", id).Error("Failed to delete subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscription"})
		return
	}

	h.log.WithField("id", id).Info("Subscription deleted successfully")
	c.JSON(http.StatusNoContent, nil)
}

// ListSubscriptions godoc
// @Summary      Список всех подписок
// @Description  Возвращает список подписок с возможностью фильтрации
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        user_id query string false "ID пользователя (UUID)"
// @Param        service_name query string false "Название сервиса"
// @Param        start_date query string false "Дата начала (YYYY-MM-DD)"
// @Param        end_date query string false "Дата окончания (YYYY-MM-DD)"
// @Success      200 {array} models.Subscription
// @Failure      400 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /api/v1/subscriptions [get]
func (h *Handler) ListSubscriptions(c *gin.Context) {
	limit := 10
	offset := 0

	if l, ok := c.GetQuery("limit"); ok {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if o, ok := c.GetQuery("offset"); ok {
		if parsedOffset, err := strconv.Atoi(o); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	var filter models.SubscriptionFilter

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err == nil {
			userIDStrClean := userID
			filter.UserID = &userIDStrClean
		}
	}

	if serviceName := c.Query("service_name"); serviceName != "" {
		filter.ServiceName = &serviceName
	}

	subs, total, err := h.repo.List(filter, limit, offset)
	if err != nil {
		h.log.WithError(err).Error("Failed to list subscriptions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   subs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// CalculateTotalCost godoc
// @Summary      Расчет общей стоимости подписок
// @Description  Рассчитывает общую стоимость подписок за период
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        start_date query string true "Дата начала периода (YYYY-MM-DD)"
// @Param        end_date query string true "Дата окончания периода (YYYY-MM-DD)"
// @Param        user_id query string false "ID пользователя (UUID)"
// @Param        service_name query string false "Название сервиса"
// @Success      200 {object} models.TotalCostResponse
// @Failure      400 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /api/v1/subscriptions/total-cost [get]
func (h *Handler) CalculateTotalCost(c *gin.Context) {
	var filter models.SubscriptionFilter

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err == nil {
			// Преобразовано в строку с помощью .String(), так как фильтр ожидает string (строка 195:20)
			userIDStrClean := userID
			filter.UserID = &userIDStrClean
		}
	}

	if serviceName := c.Query("service_name"); serviceName != "" {
		filter.ServiceName = &serviceName
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			filter.EndDate = &endDate
		}
	}

	// Удален контекст из вызова метода
	totalCost, err := h.repo.CalculateTotalCost(&filter)
	if err != nil {
		h.log.WithError(err).Error("Failed to calculate statistics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate statistics"})
		return
	}

	h.log.WithFields(logrus.Fields{
		"total_cost": totalCost,
		"filter":     filter,
	}).Info("Statistics calculated successfully")

	c.JSON(http.StatusOK, models.TotalCostResponse{TotalCost: totalCost})
}
