package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wishes/models"
	"wishes/services"
	"wishes/utils"
)

type WishController struct {
	wishService   *services.WishService
	recordService *services.RecordService
	userService   *services.UserService
}

func NewWishController(
	wishService *services.WishService,
	recordService *services.RecordService,
	userService *services.UserService,
) *WishController {
	return &WishController{
		wishService:   wishService,
		recordService: recordService,
		userService:   userService,
	}
}

type GetWishesResponse struct {
	Items      []models.Wish    `json:"items"`
	Pagination utils.Pagination `json:"pagination"`
}

// GetWishes godoc
// @Summary      [小程序/后台]获取心愿列表
// @Description  获取心愿列表，支持分页和过滤
// @Tags         心愿
// @Accept       json
// @Produce      json
// @Param        content      query     string  false  "按心愿内容模糊搜索"
// @Param        is-done      query     bool    false  "按完成状态过滤,默认为false"  default(false)
// @Param        is-published query     bool    false  "按公开状态过滤,不传为全部"
// @Param        page-index   query     int     false  "页码，默认1"  default(1)
// @Param        page-size    query     int     false  "每页数量，默认10"  default(10)
// @Success      200  {object}  GetWishesResponse  "返回心愿列表和分页信息"
// @Failure      500  {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/wishes [get]
func (c *WishController) GetWishes(ctx *gin.Context) {
	content := ctx.Query("content")
	isDoneStr := ctx.DefaultQuery("is-done", "false")
	isPublishedStr := ctx.Query("is-published") // 不设置默认值，不传表示全部
	pageIndexStr := ctx.DefaultQuery("page-index", "1")
	pageSizeStr := ctx.DefaultQuery("page-size", "10")

	pageIndex, err := strconv.Atoi(pageIndexStr)
	if err != nil || pageIndex < 1 {
		pageIndex = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	filters := map[string]any{
		"content":     content,
		"isDone":      isDoneStr,
		"isPublished": isPublishedStr, // 添加公开状态过滤
		"pageIndex":   pageIndex,
		"pageSize":    pageSize,
	}

	wishes, total, err := c.wishService.GetWishes(filters)
	if err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "获取心愿列表失败"))
		return
	}

	ctx.JSON(200, utils.CreateResponse(GetWishesResponse{
		Items:      wishes,
		Pagination: utils.NewPagination(total, pageIndex, pageSize),
	}))
}

type CreateWishRequest struct {
	ChildName string        `json:"childName"`
	Gender    models.Gender `json:"gender"`
	Content   string        `json:"content"`
	Reason    string        `json:"reason"`
	Grade     string        `json:"grade,omitempty"`
	PhotoURL  string        `json:"photoUrl,omitempty"`

	IsPublished bool `json:"isPublished,omitempty"`
}

// CreateWish godoc
// @Summary      [后台]创建新心愿
// @Description  创建一个新的心愿
// @Tags         心愿
// @Accept       json
// @Produce      json
// @Param        request  body      CreateWishRequest  true  "心愿信息"
// @Success      201   {object}  models.Wish  "返回创建的心愿"
// @Failure      400   {object}  map[string]interface{}  "请求数据无效"
// @Failure      500   {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/wishes [post]
func (c *WishController) CreateWish(ctx *gin.Context) {
	var wish CreateWishRequest
	if err := ctx.ShouldBindJSON(&wish); err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, "无效的请求数据"))
		return
	}

	newWish := models.Wish{
		ChildName:   wish.ChildName,
		Gender:      wish.Gender,
		Content:     wish.Content,
		Reason:      wish.Reason,
		Grade:       &wish.Grade,
		PhotoURL:    &wish.PhotoURL,
		IsPublished: wish.IsPublished,
	}

	if err := c.wishService.CreateWish(&newWish); err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "无法创建心愿"))
		return
	}

	ctx.JSON(201, utils.CreateResponse(newWish))
}

// DeleteWish godoc
// @Summary      [后台]删除心愿
// @Description  删除心愿
// @Tags         心愿
// @Accept       json
// @Produce      json
// @Param        id    path    uint    true  "心愿ID"
// @Success      200   {object}  map[string]interface{}  "成功删除心愿"
// @Failure      400   {object}  map[string]interface{}  "请求数据无效"
// @Failure      500   {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/wishes/{id} [delete]
func (c *WishController) DeleteWish(ctx *gin.Context) {
	wishID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, "无效的心愿ID"))
		return
	}

	if err := c.wishService.DeleteWish(uint(wishID)); err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "无法删除心愿"))
		return
	}

	ctx.JSON(200, utils.CreateResponse(nil))
}

type UpdateWishRequest struct {
	ChildName string        `json:"childName"`
	Gender    models.Gender `json:"gender"`
	Content   string        `json:"content"`
	Reason    string        `json:"reason"`
	Grade     string        `json:"grade"`
	PhotoURL  string        `json:"photoUrl"`

	IsPublished bool `json:"isPublished"`
}

// UpdateWish godoc
// @Summary      [后台]更新心愿
// @Description  更新心愿
// @Tags         心愿
// @Accept       json
// @Produce      json
// @Param        id    path    uint    true  "心愿ID"
// @Param        request  body      UpdateWishRequest  true  "心愿信息"
// @Success      200   {object}  models.Wish  "返回更新后的心愿"
// @Failure      400   {object}  map[string]interface{}  "请求数据无效"
// @Failure      500   {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/wishes/{id} [put]
func (c *WishController) UpdateWish(ctx *gin.Context) {
	wishID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, "无效的心愿ID"))
		return
	}

	wish, err := c.wishService.GetWishByID(uint(wishID))
	if err != nil {
		ctx.JSON(404, utils.CreateResponse(nil, "心愿不存在"))
		return
	}

	var wishInfo UpdateWishRequest
	if err := ctx.ShouldBindJSON(&wishInfo); err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, "无效的请求数据"))
		return
	}

	wish.ChildName = wishInfo.ChildName
	wish.Gender = wishInfo.Gender
	wish.Content = wishInfo.Content
	wish.Reason = wishInfo.Reason
	wish.Grade = &wishInfo.Grade
	wish.PhotoURL = &wishInfo.PhotoURL
	wish.IsPublished = wishInfo.IsPublished

	if err := c.wishService.UpdateWish(wish); err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "无法更新心愿"))
		return
	}

	ctx.JSON(200, utils.CreateResponse(wish))
}

type UpdateWishDonorRequest struct {
	Name    string `json:"donorName"`
	Mobile  string `json:"donorMobile"`
	Address string `json:"address"`
	Comment string `json:"comment"`
}

// ClaimWish godoc
// @Summary      [小程序]点亮心愿
// @Description  创建一条认领记录
// @Tags         心愿
// @Accept       json
// @Produce      json
// @Param        id    path    uint    true  "心愿ID"
// @Param        request  body      UpdateWishDonorRequest  true  "捐赠者信息"
// @Success      200   {object}  models.Wish  "返回更新后的心愿"
// @Failure      400   {object}  map[string]interface{}  "请求数据无效"
// @Failure      404   {object}  map[string]interface{}  "心愿不存在"
// @Failure      500   {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/wishes/{id}/donor [put]
func (c *WishController) ClaimWish(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(401, utils.CreateResponse(nil, "登录已过期"))
		return
	}
	userType, exists := ctx.Get("userType")
	if !exists || userType != "user" {
		ctx.JSON(401, utils.CreateResponse(nil, "只有用户可以点亮心愿"))
		return
	}

	donor, err := c.userService.GetUserByID(userID.(uint))
	if err != nil {
		ctx.JSON(404, utils.CreateResponse(nil, "用户不存在"))
		return
	}

	wishID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, "无效的心愿ID"))
		return
	}

	wish, err := c.wishService.GetWishByID(uint(wishID))
	if err != nil {
		ctx.JSON(404, utils.CreateResponse(nil, "心愿不存在"))
		return
	}

	var donorInfo UpdateWishDonorRequest
	if err := ctx.ShouldBindJSON(&donorInfo); err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, "无效的认领人信息"))
		return
	}

	if wish.ActiveRecordID != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "该心愿已被认领"))
		return
	}

	newRecord := models.WishRecord{
		DonorName:    donorInfo.Name,
		DonorMobile:  donorInfo.Mobile,
		DonorAddress: donorInfo.Address,
		DonorComment: donorInfo.Comment,

		WishID:  wish.ID,
		DonorID: donor.ID,
	}
	if err := c.recordService.CreateRecordWithWish(&newRecord, wish); err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "创建认领记录失败"))
		return
	}

	// 查询完整的记录信息（可选，如果您需要返回关联的对象）
	createdRecord, err := c.recordService.GetRecordByIDWithoutRecursion(newRecord.ID)
	if err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "无法获取创建的记录信息"))
		return
	}

	// 返回新创建的记录
	ctx.JSON(200, utils.CreateResponse(createdRecord))
}

type BatchCreateWishItem struct {
	ChildName string        `json:"childName"`
	Gender    models.Gender `json:"gender"`
	Content   string        `json:"content"`
	Reason    string        `json:"reason"`
	Grade     string        `json:"grade,omitempty"`
	PhotoURL  string        `json:"photoUrl,omitempty"`
}

type BatchCreateWishRequest struct {
	Data []BatchCreateWishItem `json:"data"`
}

// BatchCreateWishes godoc
// @Summary      [后台]批量导入心愿
// @Description  批量导入多个心愿
// @Tags         心愿
// @Accept       json
// @Produce      json
// @Param        request  body   BatchCreateWishRequest  true  "心愿信息数组"
// @Success      201   {object}  map[string]interface{}  "返回导入结果"
// @Failure      400   {object}  map[string]interface{}  "请求数据无效"
// @Failure      500   {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/wishes/batch [post]
func (c *WishController) BatchCreateWishes(ctx *gin.Context) {
	var wishRequest BatchCreateWishRequest
	if err := ctx.ShouldBindJSON(&wishRequest); err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, "无效的请求数据"))
		return
	}

	// 检查请求数据
	if len(wishRequest.Data) == 0 {
		ctx.JSON(400, utils.CreateResponse(nil, "导入列表不能为空"))
		return
	}

	// 转换为模型数组
	wishes := make([]*models.Wish, 0, len(wishRequest.Data))
	for _, item := range wishRequest.Data {
		grade := item.Grade
		photoURL := item.PhotoURL

		wish := &models.Wish{
			ChildName: item.ChildName,
			Gender:    item.Gender,
			Content:   item.Content,
			Reason:    item.Reason,
			Grade:     &grade,
			PhotoURL:  &photoURL,
			// 默认设置为公开
			IsPublished: true,
		}
		wishes = append(wishes, wish)
	}

	// 批量创建心愿
	if err := c.wishService.BatchCreateWishes(wishes); err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "批量导入心愿失败"))
		return
	}

	ctx.JSON(201, utils.CreateResponse(map[string]interface{}{
		"success": true,
		"count":   len(wishes),
		"message": "批量导入心愿成功",
	}))
}
