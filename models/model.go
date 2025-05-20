package models

// @Description 基础模型结构，包含ID、创建时间、更新时间和删除时间
type Model struct {
	ID        uint  `gorm:"primaryKey" json:"id"`
	CreatedAt int64 `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt int64 `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt int64 `json:"deletedAt" gorm:"autoDeleteTime"`
}

// @Description 微信小程序用户信息
type User struct {
	Model
	WechatOpenID  string `json:"wechatOpenId,omitempty" gorm:"uniqueIndex;column:wechat_openid"`
	WechatUnionID string `json:"wechatUnionId,omitempty" gorm:"column:wechat_unionid"`
	AvatarURL     string `json:"avatarUrl,omitempty" gorm:"column:avatar_url"`
	Nickname      string `json:"nickname,omitempty"`
}

// @Description 系统管理员信息
type Admin struct {
	Model
	Username string `json:"username" gorm:"uniqueIndex"`
	Password string `json:"password,omitempty" gorm:"not null"`
}

// @Description 用户性别类型
type Gender string

const (
	Male   Gender = "male"
	Female Gender = "female"
)

// @Description 心愿信息
type Wish struct {
	Model

	ChildName string  `json:"childName"`
	Gender    Gender  `json:"gender"`
	Content   string  `json:"content"`
	Reason    string  `json:"reason"`
	Grade     *string `json:"grade,omitempty"`
	PhotoURL  *string `json:"photoUrl,omitempty"`

	IsPublished bool `json:"isPublished" gorm:"default:false"`

	ActiveRecordID *uint       `json:"activeRecordId,omitempty"`
	ActiveRecord   *WishRecord `json:"activeRecord,omitempty" gorm:"foreignKey:ActiveRecordID"`
}

// @Description 心愿认领记录状态
type WishRecordStatus string

const (
	StatusPendingShipment     WishRecordStatus = "pending_shipment"
	StatusPendingConfirmation WishRecordStatus = "pending_confirmation"
	StatusConfirmed           WishRecordStatus = "confirmed"
	StatusAwaitingReceipt     WishRecordStatus = "awaiting_receipt"
	StatusCompleted           WishRecordStatus = "completed"
	StatusGiftReturned        WishRecordStatus = "gift_returned"
	StatusCancelled           WishRecordStatus = "cancelled"
)

// @Description 心愿认领记录
type WishRecord struct {
	Model

	Status WishRecordStatus `json:"status" gorm:"default:'pending_shipment'"`

	WishID uint  `json:"wishId" gorm:"index"`
	Wish   *Wish `json:"wish,omitempty" gorm:"foreignKey:WishID"`

	DonorID uint  `json:"donorId,omitempty" gorm:"index"`
	Donor   *User `json:"donor,omitempty" gorm:"foreignKey:DonorID"`

	DonorName    string `json:"donorName"`
	DonorMobile  string `json:"donorMobile"`
	DonorAddress string `json:"donorAddress"`
	DonorComment string `json:"donorComment"`

	ShippingNumber *string `json:"shippingNumber,omitempty"` // 寄送单号
	ShippingTime   *int64  `json:"shippingTime,omitempty"`   // 寄送时间

	ConfirmationMessage *string `json:"confirmationMessage,omitempty"` // 确认信息
	ConfirmationPhotos  *string `json:"confirmationPhotos,omitempty"`  // 确认照片数组
	ConfirmationTime    *int64  `json:"confirmationTime,omitempty"`    // 确认时间

	DeliveryNumber *string `json:"deliveryNumber,omitempty"` // 发货单号
	DeliveryTime   *int64  `json:"deliveryTime,omitempty"`   // 发货时间

	ReceiptMessage *string `json:"receiptMessage,omitempty"` // 签收信息
	ReceiptPhotos  *string `json:"receiptPhotos,omitempty"`  // 签收照片数组
	ReceiptTime    *int64  `json:"receiptTime,omitempty"`    // 签收时间

	PlatformGiftMessage *string `json:"platformGiftMessage,omitempty"` // 平台回礼信息
	PlatformGiftPhotos  *string `json:"platformGiftPhotos,omitempty"`  // 平台回礼照片数组
	PlatformGiftTime    *int64  `json:"platformGiftTime,omitempty"`    // 平台回礼时间
	OwnerGiftMessage    *string `json:"ownerGiftMessage,omitempty"`    // 心愿主人回礼信息
	OwnerGiftPhotos     *string `json:"ownerGiftPhotos,omitempty"`     // 心愿主人回礼照片数组
	OwnerGiftTime       *int64  `json:"ownerGiftTime,omitempty"`       // 心愿主人回礼时间

	CancellationTime *int64 `json:"cancellationTime,omitempty"` // 取消时间
}
