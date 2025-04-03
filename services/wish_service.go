package services

import (
	"strconv"

	"gorm.io/gorm"

	"wishes/models"
)

type WishService struct {
	db *gorm.DB
}

func NewWishService(db *gorm.DB) *WishService {
	return &WishService{
		db: db,
	}
}

func (s *WishService) FindWishes(filters map[string]any) ([]models.Wish, int64, error) {
	query := s.db.Model(&models.Wish{})

	if childName, ok := filters["childName"].(string); ok && childName != "" {
		query = query.Where("child_name LIKE ?", "%"+childName+"%")
	}
	if content, ok := filters["content"].(string); ok && content != "" {
		query = query.Where("content LIKE ?", "%"+content+"%")
	}
	if isDoneStr, ok := filters["isDone"].(string); ok && isDoneStr != "" {
		if isBool, err := strconv.ParseBool(isDoneStr); err == nil {
			query = query.Where("is_done = ?", isBool)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageIndex := filters["pageIndex"].(int)
	pageSize := filters["pageSize"].(int)
	offset := (pageIndex - 1) * pageSize

	var wishes []models.Wish
	if err := query.Limit(pageSize).Offset(offset).Find(&wishes).Error; err != nil {
		return nil, 0, err
	}

	return wishes, total, nil
}

func (s *WishService) CreateWish(wish *models.Wish) error {
	return s.db.Create(wish).Error
}

func (s *WishService) GetWishByID(id uint) (*models.Wish, error) {
	var wish models.Wish
	if err := s.db.First(&wish, id).Error; err != nil {
		return nil, err
	}
	return &wish, nil
}

func (s *WishService) UpdateWish(wish *models.Wish) error {
	return s.db.Save(wish).Error
}

func (s *WishService) DeleteWish(id uint) error {
	return s.db.Delete(&models.Wish{}, id).Error
}

func (s *WishService) CreateOrUpdateUser(user *models.User) error {
	if user.ID == 0 {
		return s.db.Create(user).Error
	}

	return s.db.Save(user).Error
}
