package services

import (
	"fmt"
	"time"
	"wishes/models"

	"gorm.io/gorm"
)

type RecordService struct {
	db *gorm.DB
}

func NewRecordService(db *gorm.DB) *RecordService {
	return &RecordService{
		db: db,
	}
}

func (s *RecordService) GetAllRecords(pageIndex, pageSize int, status string) ([]models.WishRecord, int64, error) {
	query := s.db.Model(&models.WishRecord{})

	// 如果指定了状态，则按状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 预加载关联数据
	query = query.Preload("Wish").Preload("Donor")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageIndex - 1) * pageSize

	var records []models.WishRecord
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (s *RecordService) GetRecordsByUserID(userID uint, pageIndex, pageSize int) ([]models.WishRecord, int64, error) {
	query := s.db.Model(&models.WishRecord{}).Where("donor_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageIndex - 1) * pageSize

	var records []models.WishRecord
	if err := query.Limit(pageSize).Offset(offset).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (s *RecordService) CreateRecordWithWish(record *models.WishRecord, wish *models.Wish) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		wish.ActiveRecordID = &record.ID
		if err := tx.Save(wish).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *RecordService) GetRecordByIDWithoutRecursion(id uint) (*models.WishRecord, error) {
	var record models.WishRecord
	// 预加载Wish和Donor，但是不预加载Wish.ActiveRecord
	if err := s.db.Preload("Wish", func(db *gorm.DB) *gorm.DB {
		return db.Omit("ActiveRecord")
	}).Preload("Donor").First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *RecordService) UpdateRecordStatus(recordID uint, newStatus models.WishRecordStatus, params map[string]interface{}) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var record models.WishRecord
		if err := tx.First(&record, recordID).Error; err != nil {
			return err
		}

		// 检查状态转换是否合法
		if !isValidStatusTransition(record.Status, newStatus) {
			return fmt.Errorf("不允许从 %s 状态转换为 %s 状态", record.Status, newStatus)
		}

		// 更新记录状态
		record.Status = newStatus

		// 根据新状态设置相应字段
		switch newStatus {
		case models.StatusPendingConfirmation:
			if shippingNumber, ok := params["shippingNumber"].(string); ok && shippingNumber != "" {
				record.ShippingNumber = shippingNumber
				record.ShippingTime = time.Now().Unix()
			} else {
				return fmt.Errorf("转换为待确认状态需要提供寄送单号")
			}
		case models.StatusConfirmed:
			if confirmationMessage, ok := params["confirmationMessage"].(string); ok {
				record.ConfirmationMessage = confirmationMessage
			}
			if confirmationPhotos, ok := params["confirmationPhotos"].(string); ok {
				record.ConfirmationPhotos = confirmationPhotos
			}
			record.ConfirmationTime = time.Now().Unix()
		case models.StatusAwaitingReceipt:
			if deliveryNumber, ok := params["deliveryNumber"].(string); ok && deliveryNumber != "" {
				record.DeliveryNumber = deliveryNumber
				record.DeliveryTime = time.Now().Unix()
			} else {
				return fmt.Errorf("转换为待收货状态需要提供发货单号")
			}
		case models.StatusCompleted:
			if receiptMessage, ok := params["receiptMessage"].(string); ok {
				record.ReceiptMessage = receiptMessage
			}
			if receiptPhotos, ok := params["receiptPhotos"].(string); ok {
				record.ReceiptPhotos = receiptPhotos
			}
			record.ReceiptTime = time.Now().Unix()
		case models.StatusGiftReturned:
			// 更新消息和照片
			if platformGiftMessage, ok := params["platformGiftMessage"].(string); ok {
				record.PlatformGiftMessage = platformGiftMessage
				if platformGiftMessage != "" {
					record.PlatformGiftTime = time.Now().Unix()
				}
			}
			if platformGiftPhotos, ok := params["platformGiftPhotos"].(string); ok {
				record.PlatformGiftPhotos = platformGiftPhotos
				if platformGiftPhotos != "" && record.PlatformGiftTime == 0 {
					record.PlatformGiftTime = time.Now().Unix()
				}
			}
			if ownerGiftMessage, ok := params["ownerGiftMessage"].(string); ok {
				record.OwnerGiftMessage = ownerGiftMessage
				if ownerGiftMessage != "" {
					record.OwnerGiftTime = time.Now().Unix()
				}
			}
			if ownerGiftPhotos, ok := params["ownerGiftPhotos"].(string); ok {
				record.OwnerGiftPhotos = ownerGiftPhotos
				if ownerGiftPhotos != "" && record.OwnerGiftTime == 0 {
					record.OwnerGiftTime = time.Now().Unix()
				}
			}
		case models.StatusCancelled:
			record.CancellationTime = time.Now().Unix()
		}

		return tx.Save(&record).Error
	})
}

// 判断状态转换是否合法
func isValidStatusTransition(currentStatus, newStatus models.WishRecordStatus) bool {
	// 根据业务流程定义合法的状态转换
	validTransitions := map[models.WishRecordStatus][]models.WishRecordStatus{
		models.StatusPendingShipment:     {models.StatusPendingConfirmation, models.StatusCancelled},
		models.StatusPendingConfirmation: {models.StatusConfirmed, models.StatusCancelled},
		models.StatusConfirmed:           {models.StatusAwaitingReceipt, models.StatusCancelled},
		models.StatusAwaitingReceipt:     {models.StatusCompleted, models.StatusCancelled},
		models.StatusCompleted:           {models.StatusGiftReturned},
		models.StatusGiftReturned:        {},
		models.StatusCancelled:           {},
	}

	for _, validStatus := range validTransitions[currentStatus] {
		if validStatus == newStatus {
			return true
		}
	}
	return false
}
