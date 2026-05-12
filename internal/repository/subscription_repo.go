package repository

import (
	"fmt"
	"time"

	"github.com/svetilka/EffectiveMobileProject/internal/models"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(subscription *models.Subscription) error {
	log.WithFields(log.Fields{
		"service_name": subscription.ServiceName,
		"user_id":      subscription.UserID,
	}).Info("Creating new subscription")

	if err := r.db.Create(subscription).Error; err != nil {
		log.WithError(err).Error("Failed to create subscription")
		return err
	}

	log.WithFields(log.Fields{
		"id":           subscription.ID,
		"service_name": subscription.ServiceName,
	}).Info("Subscription created successfully")

	return nil
}

func (r *SubscriptionRepository) GetByID(id uint) (*models.Subscription, error) {
	var subscription models.Subscription

	if err := r.db.First(&subscription, id).Error; err != nil {
		log.WithFields(log.Fields{
			"id":    id,
			"error": err,
		}).Error("Failed to get subscription by ID")
		return nil, err
	}

	return &subscription, nil
}

func (r *SubscriptionRepository) List(filter models.SubscriptionFilter, limit, offset int) ([]models.Subscription, int64, error) {
	var subscriptions []models.Subscription
	var total int64

	query := r.db.Model(&models.Subscription{})

	if filter.UserID != "" {
		userUUID, err := uuid.Parse(filter.UserID)
		if err == nil {
			query = query.Where("user_id = ?", userUUID)
		}
	}

	if filter.ServiceName != "" {
		query = query.Where("service_name = ?", filter.ServiceName)
	}

	if err := query.Count(&total).Error; err != nil {
		log.WithError(err).Error("Failed to count subscriptions")
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&subscriptions).Error; err != nil {
		log.WithError(err).Error("Failed to list subscriptions")
		return nil, 0, err
	}

	log.WithFields(log.Fields{
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"filter": filter,
	}).Info("Listed subscriptions")

	return subscriptions, total, nil
}

func (r *SubscriptionRepository) Update(id uint, updates map[string]interface{}) error {
	log.WithFields(log.Fields{
		"id":      id,
		"updates": updates,
	}).Info("Updating subscription")

	if err := r.db.Model(&models.Subscription{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		log.WithError(err).Error("Failed to update subscription")
		return err
	}

	log.WithField("id", id).Info("Subscription updated successfully")
	return nil
}

func (r *SubscriptionRepository) Delete(id uint) error {
	log.WithField("id", id).Info("Deleting subscription")

	if err := r.db.Delete(&models.Subscription{}, id).Error; err != nil {
		log.WithError(err).Error("Failed to delete subscription")
		return err
	}

	log.WithField("id", id).Info("Subscription deleted successfully")
	return nil
}

func (r *SubscriptionRepository) CalculateTotalCost(startDate, endDate time.Time, userID string, serviceName string) (int, error) {
	log.WithFields(log.Fields{
		"start_date":   startDate,
		"end_date":     endDate,
		"user_id":      userID,
		"service_name": serviceName,
	}).Info("Calculating total cost")

	query := r.db.Model(&models.Subscription{}).
		Where("start_date <= ?", endDate).
		Where("end_date IS NULL OR end_date >= ?", startDate)

	if userID != "" {
		userUUID, err := uuid.Parse(userID)
		if err == nil {
			query = query.Where("user_id = ?", userUUID)
		}
	}

	if serviceName != "" {
		query = query.Where("service_name = ?", serviceName)
	}

	var totalCost int64
	if err := query.Select("COALESCE(SUM(price), 0)").Scan(&totalCost).Error; err != nil {
		log.WithError(err).Error("Failed to calculate total cost")
		return 0, fmt.Errorf("failed to calculate total cost: %w", err)
	}

	log.WithFields(log.Fields{
		"total_cost": totalCost,
		"period":     fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
	}).Info("Total cost calculated")

	return int(totalCost), nil
}
