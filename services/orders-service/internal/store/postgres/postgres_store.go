package postgres

import (
	"errors"

	"github.com/ametowartem/orderflow/orders-service/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresStore struct {
	db *gorm.DB
}

func NewPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&OrderModel{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&OrderItemModel{}); err != nil {
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Create(order domain.Order) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		model := toOrderModel(order)
		if err := tx.Create(&model).Error; err != nil {
			return err
		}

		itemModels := make([]OrderItemModel, 0, len(order.Items))
		for _, i := range order.Items {
			itemModels = append(itemModels, toOrderItemModel(i, order.ID))
		}

		if len(itemModels) > 0 {
			if err := tx.Create(&itemModels).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *PostgresStore) GetByID(id string) (domain.Order, error) {
	var model OrderModel

	err := s.db.Preload("Items").First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Order{}, domain.ErrNotFound
		}
		return domain.Order{}, err
	}
	return toDomainOrder(model), nil
}

func (s *PostgresStore) List() ([]domain.Order, error) {
	var models []OrderModel
	if err := s.db.Preload("Items").Find(&models).Error; err != nil {
		return nil, err
	}

	orders := make([]domain.Order, 0, len(models))
	for _, m := range models {
		orders = append(orders, toDomainOrder(m))
	}
	return orders, nil
}
