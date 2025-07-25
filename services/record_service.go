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

func (s *RecordService) GetRecordsByUserID(userID uint, pageIndex, pageSize int, status string, isAdmin bool) ([]models.WishRecord, int64, error) {
	query := s.db.Model(&models.WishRecord{})

	// 如果是管理员，就查全部的数据，否则只查当前用户的
	if !isAdmin {
		query = query.Where("donor_id = ?", userID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageIndex - 1) * pageSize

	var records []models.WishRecord
	if err := query.Preload("Wish").Limit(pageSize).Offset(offset).Find(&records).Error; err != nil {
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

func (s *RecordService) UpdateRecordStatus(recordID uint, newStatus models.WishRecordStatus, params map[string]any) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var record models.WishRecord
		if err := tx.First(&record, recordID).Error; err != nil {
			return err
		}

		// // 检查状态转换是否合法
		// if !isValidStatusTransition(record.Status, newStatus) {
		// 	return fmt.Errorf("不允许从 %s 状态转换为 %s 状态", record.Status, newStatus)
		// }

		// 更新记录状态
		record.Status = newStatus

		// 根据新状态设置相应字段
		switch newStatus {
		case models.StatusPendingConfirmation:
			if shippingNumber, ok := params["shippingNumber"].(string); ok && shippingNumber != "" {
				record.ShippingNumber = &shippingNumber
				now := time.Now().Unix()
				record.ShippingTime = &now
			} else {
				return fmt.Errorf("转换为待确认状态需要提供寄送单号")
			}
		case models.StatusConfirmed:
			if confirmationMessage, ok := params["confirmationMessage"].(string); ok {
				record.ConfirmationMessage = &confirmationMessage
			}
			if confirmationPhotos, ok := params["confirmationPhotos"].(string); ok {
				record.ConfirmationPhotos = &confirmationPhotos
			}
			now := time.Now().Unix()
			record.ConfirmationTime = &now
		case models.StatusAwaitingReceipt:
			if deliveryNumber, ok := params["deliveryNumber"].(string); ok && deliveryNumber != "" {
				record.DeliveryNumber = &deliveryNumber
				now := time.Now().Unix()
				record.DeliveryTime = &now
			} else {
				return fmt.Errorf("转换为待收货状态需要提供发货单号")
			}
		case models.StatusCompleted:
			if receiptMessage, ok := params["receiptMessage"].(string); ok {
				record.ReceiptMessage = &receiptMessage
			}
			if receiptPhotos, ok := params["receiptPhotos"].(string); ok {
				record.ReceiptPhotos = &receiptPhotos
			}
			now := time.Now().Unix()
			record.ReceiptTime = &now
		case models.StatusGiftReturned:
			// 更新消息和照片
			if platformGiftMessage, ok := params["platformGiftMessage"].(string); ok {
				record.PlatformGiftMessage = &platformGiftMessage
				if platformGiftMessage != "" {
					now := time.Now().Unix()
					record.PlatformGiftTime = &now
				}
			}
			if platformGiftPhotos, ok := params["platformGiftPhotos"].(string); ok {
				record.PlatformGiftPhotos = &platformGiftPhotos
				// 需要先检查指针是否为 nil
				if platformGiftPhotos != "" && (record.PlatformGiftTime == nil || *record.PlatformGiftTime == 0) {
					now := time.Now().Unix()
					record.PlatformGiftTime = &now
				}
			}
			if ownerGiftMessage, ok := params["ownerGiftMessage"].(string); ok {
				record.OwnerGiftMessage = &ownerGiftMessage
				if ownerGiftMessage != "" {
					now := time.Now().Unix()
					record.OwnerGiftTime = &now
				}
			}
			if ownerGiftPhotos, ok := params["ownerGiftPhotos"].(string); ok {
				record.OwnerGiftPhotos = &ownerGiftPhotos
				// 需要先检查指针是否为 nil
				if ownerGiftPhotos != "" && (record.OwnerGiftTime == nil || *record.OwnerGiftTime == 0) {
					now := time.Now().Unix()
					record.OwnerGiftTime = &now
				}
			}
		case models.StatusCancelled:
			now := time.Now().Unix()
			record.CancellationTime = &now
		}

		return tx.Save(&record).Error
	})
}

// UpdateShippingInfo 更新收货信息
func (s *RecordService) UpdateShippingInfo(recordID uint, donorName, donorMobile, donorAddress string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var record models.WishRecord
		if err := tx.First(&record, recordID).Error; err != nil {
			return err
		}

		// 更新收货信息
		record.DonorName = donorName
		record.DonorMobile = donorMobile
		record.DonorAddress = donorAddress

		return tx.Save(&record).Error
	})
}

// // 判断状态转换是否合法
// func isValidStatusTransition(currentStatus, newStatus models.WishRecordStatus) bool {
// 	// 根据业务流程定义合法的状态转换
// 	validTransitions := map[models.WishRecordStatus][]models.WishRecordStatus{
// 		models.StatusPendingShipment:     {models.StatusPendingConfirmation, models.StatusCancelled},
// 		models.StatusPendingConfirmation: {models.StatusConfirmed, models.StatusCancelled},
// 		models.StatusConfirmed:           {models.StatusAwaitingReceipt, models.StatusCancelled},
// 		models.StatusAwaitingReceipt:     {models.StatusCompleted, models.StatusCancelled},
// 		models.StatusCompleted:           {models.StatusGiftReturned},
// 		models.StatusGiftReturned:        {},
// 		models.StatusCancelled:           {},
// 	}

// 	for _, validStatus := range validTransitions[currentStatus] {
// 		if validStatus == newStatus {
// 			return true
// 		}
// 	}
// 	return false
// }
