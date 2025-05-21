package controllers

import (
	"sort"
	"strconv"
	"wishes/models"
	"wishes/services"
	"wishes/utils"

	"github.com/gin-gonic/gin"
)

type RecordController struct {
	recordService *services.RecordService
}

func NewRecordController(
	recordService *services.RecordService,
) *RecordController {
	return &RecordController{
		recordService: recordService,
	}
}

type GetWishRecordsResponse struct {
	Items      []models.WishRecord `json:"items"`
	Pagination utils.Pagination    `json:"pagination"`
}

// GetWishRecords godoc
// @Summary      [小程序]获取用户点亮心愿的记录
// @Description  获取当前登录用户点亮心愿的记录
// @Tags         记录
// @Accept       json
// @Produce      json
// @Param        pageIndex    query    int     false  "页码，默认1"
// @Param        pageSize     query    int     false  "每页数量，默认10"
// @Success      200  {object}  controllers.GetWishesResponse  "返回用户点亮的心愿列表"
// @Failure      401  {object}  map[string]interface{}  "用户未登录"
// @Failure      500  {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/user/records [get]
func (c *RecordController) GetWishRecords(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(401, utils.CreateResponse(nil, "登录已过期"))
		return
	}

	userType, exists := ctx.Get("userType")
	if !exists || userType != "user" {
		ctx.JSON(401, utils.CreateResponse(nil, "只有用户可以查看点亮的心愿"))
		return
	}

	pageIndexStr := ctx.DefaultQuery("pageIndex", "1")
	pageIndex, err := strconv.Atoi(pageIndexStr)
	if err != nil || pageIndex < 1 {
		pageIndex = 1
	}

	pageSizeStr := ctx.DefaultQuery("pageSize", "10")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	status := ctx.Query("status")

	records, total, err := c.recordService.GetRecordsByUserID(userID.(uint), pageIndex, pageSize, status)
	if err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "获取心愿列表失败"))
		return
	}

	response := GetWishRecordsResponse{
		Items:      records,
		Pagination: utils.NewPagination(total, pageIndex, pageSize),
	}

	ctx.JSON(200, utils.CreateResponse(response))
}

// GetAllRecords godoc
// @Summary      [后台]获取所有心愿认领记录
// @Description  获取系统中所有心愿认领记录，支持分页和状态过滤
// @Tags         记录
// @Accept       json
// @Produce      json
// @Param        pageIndex    query    int     false  "页码，默认1"
// @Param        pageSize     query    int     false  "每页数量，默认10"
// @Param        status        query    string  false  "状态过滤，可选值：pending_shipment, pending_confirmation等"
// @Success      200  {object}  controllers.GetWishRecordsResponse  "返回记录列表"
// @Failure      401  {object}  map[string]interface{}  "用户未登录或无权限"
// @Failure      500  {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/admin/records [get]
func (c *RecordController) GetAllRecords(ctx *gin.Context) {
	userType, exists := ctx.Get("userType")
	if !exists || userType != "admin" {
		ctx.JSON(401, utils.CreateResponse(nil, "只有管理员可以查看所有记录"))
		return
	}

	pageIndexStr := ctx.DefaultQuery("pageIndex", "1")
	pageIndex, err := strconv.Atoi(pageIndexStr)
	if err != nil || pageIndex < 1 {
		pageIndex = 1
	}

	pageSizeStr := ctx.DefaultQuery("pageSize", "10")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	status := ctx.DefaultQuery("status", "")

	records, total, err := c.recordService.GetAllRecords(pageIndex, pageSize, status)
	if err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "获取记录列表失败"))
		return
	}

	response := GetWishRecordsResponse{
		Items:      records,
		Pagination: utils.NewPagination(total, pageIndex, pageSize),
	}

	ctx.JSON(200, utils.CreateResponse(response))
}

// 进度项结构体
type ProgressItem struct {
	Type           string `json:"type"`                     // 进度类型：creation, shipping, confirmation, delivery, receipt, cancellation
	Status         string `json:"status"`                   // 对应的状态值
	Timestamp      int64  `json:"timestamp"`                // 时间戳
	Message        string `json:"message,omitempty"`        // 信息，如有
	Photos         string `json:"photos,omitempty"`         // 照片，如有
	TrackingNumber string `json:"trackingNumber,omitempty"` // 单号，如有
}

// 详细记录响应结构体
type RecordDetailResponse struct {
	// 记录基本信息
	ID        uint                    `json:"id"`
	CreatedAt int64                   `json:"createdAt"`
	UpdatedAt int64                   `json:"updatedAt"`
	DeletedAt int64                   `json:"deletedAt,omitempty"`
	Status    models.WishRecordStatus `json:"status"`

	// 进度数组
	Progress []ProgressItem `json:"progress"`

	ChildName   string `json:"childName"`
	WishContent string `json:"wishContent"`
	WishReason  string `json:"wishReason"`
	ClaimedAt   int64  `json:"claimedAt"`

	DonorName    string `json:"donorName,omitempty"`
	DonorMobile  string `json:"donorMobile,omitempty"`
	DonorAddress string `json:"donorAddress,omitempty"`
}

// GetRecordByID godoc
// @Summary      [小程序/后台]获取单个心愿认领记录详情
// @Description  根据ID获取单个心愿认领记录的详细信息
// @Tags         记录
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "记录ID"
// @Success      200  {object}  controllers.RecordDetailResponse  "返回记录详情"
// @Failure      400  {object}  map[string]interface{}  "无效的ID"
// @Failure      401  {object}  map[string]interface{}  "未登录或无权限"
// @Failure      404  {object}  map[string]interface{}  "记录不存在"
// @Failure      500  {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/records/{id} [get]
func (c *RecordController) GetRecordByID(ctx *gin.Context) {
	// 获取路径参数中的ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, "无效的记录ID"))
		return
	}

	// 获取记录详情
	record, err := c.recordService.GetRecordByIDWithoutRecursion(uint(id))
	if err != nil {
		ctx.JSON(404, utils.CreateResponse(nil, "记录不存在"))
		return
	}

	// 检查访问权限（管理员可以查看任何记录，用户只能查看自己的记录）
	userID, exists := ctx.Get("userID")
	userType, _ := ctx.Get("userType")

	if userType == "admin" || (exists && userType == "user" && record.DonorID == userID.(uint)) {
		// 构建进度数组
		var progressItems []ProgressItem

		// 取消
		if record.CancellationTime != nil && *record.CancellationTime > 0 {
			progressItems = append(progressItems, ProgressItem{
				Type:      "cancellation",
				Status:    string(models.StatusCancelled),
				Timestamp: *record.CancellationTime,
			})
		}

		// // 心愿主人回礼
		// if record.OwnerGiftTime > 0 {
		// 	progressItems = append(progressItems, ProgressItem{
		// 		Type:      "ownerGift",
		// 		Status:    string(models.StatusGiftReturned),
		// 		Timestamp: record.OwnerGiftTime,
		// 		Message:   record.OwnerGiftMessage,
		// 		Photos:    record.OwnerGiftPhotos,
		// 	})
		// }

		// // 平台回礼
		// if record.PlatformGiftTime > 0 {
		// 	progressItems = append(progressItems, ProgressItem{
		// 		Type:      "platformGift",
		// 		Status:    string(models.StatusGiftReturned),
		// 		Timestamp: record.PlatformGiftTime,
		// 		Message:   record.PlatformGiftMessage,
		// 		Photos:    record.PlatformGiftPhotos,
		// 	})
		// }

		// 签收
		if record.ReceiptTime != nil && *record.ReceiptTime > 0 {
			progressItems = append(progressItems, ProgressItem{
				Type:      "receipt",
				Status:    string(models.StatusCompleted),
				Timestamp: *record.ReceiptTime,
				Message:   *record.ReceiptMessage,
				Photos:    *record.ReceiptPhotos,
			})
		}

		// 发货
		if record.DeliveryTime != nil && *record.DeliveryTime > 0 {
			progressItems = append(progressItems, ProgressItem{
				Type:           "delivery",
				Status:         string(models.StatusAwaitingReceipt),
				Timestamp:      *record.DeliveryTime,
				TrackingNumber: *record.DeliveryNumber,
			})
		}

		// 确认
		if record.ConfirmationTime != nil && *record.ConfirmationTime > 0 {
			progressItems = append(progressItems, ProgressItem{
				Type:      "confirmation",
				Status:    string(models.StatusConfirmed),
				Timestamp: *record.ConfirmationTime,
				Message:   *record.ConfirmationMessage,
				Photos:    *record.ConfirmationPhotos,
			})
		}

		// 寄送
		if record.ShippingTime != nil && *record.ShippingTime > 0 {
			progressItems = append(progressItems, ProgressItem{
				Type:           "shipping",
				Status:         string(models.StatusPendingConfirmation),
				Timestamp:      *record.ShippingTime,
				TrackingNumber: *record.ShippingNumber,
			})
		}

		// 创建（认领）
		progressItems = append(progressItems, ProgressItem{
			Type:      "creation",
			Status:    string(models.StatusPendingShipment),
			Timestamp: record.CreatedAt,
		})

		// 按时间降序排序（虽然我们是按这个顺序添加的，但为了保险起见）
		sort.Slice(progressItems, func(i, j int) bool {
			return progressItems[i].Timestamp > progressItems[j].Timestamp
		})

		response := RecordDetailResponse{
			// 记录基本信息
			ID:        record.ID,
			CreatedAt: record.CreatedAt,
			UpdatedAt: record.UpdatedAt,
			DeletedAt: record.DeletedAt,
			Status:    record.Status,

			// 进度数组
			Progress: progressItems,

			ChildName:   record.Wish.ChildName,
			WishContent: record.Wish.Content,
			WishReason:  record.Wish.Reason,
			ClaimedAt:   record.CreatedAt,

			DonorName:    record.DonorName,
			DonorMobile:  record.DonorMobile,
			DonorAddress: record.DonorAddress,
		}

		ctx.JSON(200, utils.CreateResponse(response))
		return
	}

	// 其他情况视为无权限
	ctx.JSON(401, utils.CreateResponse(nil, "无权查看此记录"))
}

type UpdateRecordStatusRequest struct {
	Status              models.WishRecordStatus `json:"status"`
	ShippingNumber      string                  `json:"shippingNumber,omitempty"`
	ConfirmationMessage string                  `json:"confirmationMessage,omitempty"`
	ConfirmationPhotos  string                  `json:"confirmationPhotos,omitempty"`
	DeliveryNumber      string                  `json:"deliveryNumber,omitempty"`
	ReceiptMessage      string                  `json:"receiptMessage,omitempty"`
	ReceiptPhotos       string                  `json:"receiptPhotos,omitempty"`
	PlatformGiftMessage string                  `json:"platformGiftMessage,omitempty"`
	PlatformGiftPhotos  string                  `json:"platformGiftPhotos,omitempty"`
	OwnerGiftMessage    string                  `json:"ownerGiftMessage,omitempty"`
	OwnerGiftPhotos     string                  `json:"ownerGiftPhotos,omitempty"`
}

// UpdateRecordStatus godoc
// @Summary      [小程序/后台]更新心愿认领记录状态
// @Description  更新记录状态并提供相应所需信息
// @Tags         记录
// @Accept       json
// @Produce      json
// @Param        id      path      int                        true  "记录ID"
// @Param        params  body      UpdateRecordStatusRequest  true  "更新参数"
// @Success      200     {object}  models.WishRecord          "返回更新后的记录"
// @Failure      400     {object}  map[string]interface{}     "参数错误"
// @Failure      401     {object}  map[string]interface{}     "未登录或无权限"
// @Failure      404     {object}  map[string]interface{}     "记录不存在"
// @Failure      500     {object}  map[string]interface{}     "服务器错误"
// @Router       /api/v1/records/{id}/status [put]
func (c *RecordController) UpdateRecordStatus(ctx *gin.Context) {
	// 获取路径参数中的ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, "无效的记录ID"))
		return
	}

	// 获取记录详情，检查是否存在
	_, err = c.recordService.GetRecordByIDWithoutRecursion(uint(id))
	if err != nil {
		ctx.JSON(404, utils.CreateResponse(nil, "记录不存在"))
		return
	}

	// 检查权限
	userType, _ := ctx.Get("userType")
	if userType != "admin" {
		ctx.JSON(401, utils.CreateResponse(nil, "只有管理员可以更新记录状态"))
		return
	}

	// 解析请求体
	var req UpdateRecordStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, "无效的请求参数"))
		return
	}

	// 构造更新参数
	params := map[string]any{
		"shippingNumber":      req.ShippingNumber,
		"confirmationMessage": req.ConfirmationMessage,
		"confirmationPhotos":  req.ConfirmationPhotos,
		"deliveryNumber":      req.DeliveryNumber,
		"receiptMessage":      req.ReceiptMessage,
		"receiptPhotos":       req.ReceiptPhotos,
		"platformGiftMessage": req.PlatformGiftMessage,
		"platformGiftPhotos":  req.PlatformGiftPhotos,
		"ownerGiftMessage":    req.OwnerGiftMessage,
		"ownerGiftPhotos":     req.OwnerGiftPhotos,
	}

	// 更新状态
	if err := c.recordService.UpdateRecordStatus(uint(id), req.Status, params); err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, err.Error()))
		return
	}

	// 获取更新后的记录
	updatedRecord, err := c.recordService.GetRecordByIDWithoutRecursion(uint(id))
	if err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "获取更新后的记录失败"))
		return
	}

	ctx.JSON(200, utils.CreateResponse(updatedRecord))
}
