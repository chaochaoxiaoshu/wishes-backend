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
	WechatOpenID  string `json:"wechatOpenId,omitempty" gorm:"uniqueIndex"`
	WechatUnionID string `json:"wechatUnionId,omitempty"`
	AvatarURL     string `json:"avatarUrl,omitempty"`
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

// @Description 儿童心愿信息
type Wish struct {
	Model

	ChildName string `json:"childName"`
	Grade     string `json:"grade,omitempty"`
	Gender    Gender `json:"gender"`
	Content   string `json:"content"`
	PhotoURL  string `json:"photoUrl,omitempty"`

	IsDone bool `json:"isDone" gorm:"default:false"`

	DonorID *uint `json:"donorId,omitempty" gorm:"index"`
	Donor   *User `json:"donor,omitempty" gorm:"foreignKey:DonorID"`

	DonorName    string `json:"donorName,omitempty"`
	DonorMobile  string `json:"donorMobile,omitempty"`
	DonorAddress string `json:"donorAddress,omitempty"`
	DonorComment string `json:"donorComment,omitempty"`
}
