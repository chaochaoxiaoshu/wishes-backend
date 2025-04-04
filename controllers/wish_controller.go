package controllers

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"wishes/models"
	"wishes/services"
	"wishes/utils"
)

type WishController struct {
	wishService *services.WishService
	userService *services.UserService
}

func NewWishController(db *gorm.DB, wishService *services.WishService, userService *services.UserService) *WishController {
	return &WishController{
		wishService: wishService,
		userService: userService,
	}
}

type Pagination struct {
	Total     int `json:"total"`
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	PageCount int `json:"pageTotal"`
}

type GetWishesResponse struct {
	Items      []models.Wish `json:"items"`
	Pagination Pagination    `json:"pagination"`
}

// GetWishes godoc
// @Summary      获取心愿列表
// @Description  获取心愿列表，支持分页和过滤
// @Tags         心愿
// @Accept       json
// @Produce      json
// @Param        content      query     string  false  "按心愿内容模糊搜索"
// @Param        is-done      query     bool    false  "按完成状态过滤,默认为false"  default(false)
// @Param        page-index   query     int     false  "页码，默认1"  default(1)
// @Param        page-size    query     int     false  "每页数量，默认10"  default(10)
// @Success      200  {object}  GetWishesResponse  "返回心愿列表和分页信息"
// @Failure      500  {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/wishes [get]
func (c *WishController) GetWishes(ctx *gin.Context) {
	content := ctx.Query("content")
	isDoneStr := ctx.DefaultQuery("is-done", "false")
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
		"content":   content,
		"isDone":    isDoneStr,
		"pageIndex": pageIndex,
		"pageSize":  pageSize,
	}

	wishes, total, err := c.wishService.GetWishes(filters)
	if err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "获取心愿列表失败"))
		return
	}

	ctx.JSON(200, utils.CreateResponse(GetWishesResponse{
		Items: wishes,
		Pagination: Pagination{
			Total:     int(total),
			PageIndex: pageIndex,
			PageSize:  pageSize,
			PageCount: int(math.Ceil(float64(total) / float64(pageSize))),
		},
	}))
}

type CreateWishRequest struct {
	ChildName string        `json:"childName"`
	Grade     string        `json:"grade"`
	Gender    models.Gender `json:"gender"`
	Content   string        `json:"content"`
	PhotoURL  string        `json:"photoUrl"`
}

// CreateWish godoc
// @Summary      创建新心愿
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
		ChildName: wish.ChildName,
		Grade:     wish.Grade,
		Gender:    wish.Gender,
		Content:   wish.Content,
		PhotoURL:  wish.PhotoURL,
	}

	if err := c.wishService.CreateWish(&newWish); err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "无法创建心愿"))
		return
	}

	ctx.JSON(201, utils.CreateResponse(newWish))
}

// DeleteWish godoc
// @Summary      删除心愿
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
	Grade     string        `json:"grade"`
	Gender    models.Gender `json:"gender"`
	Content   string        `json:"content"`
	PhotoURL  string        `json:"photoUrl"`
}

// UpdateWish godoc
// @Summary      更新心愿
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
	wish.Grade = wishInfo.Grade
	wish.Gender = wishInfo.Gender
	wish.Content = wishInfo.Content
	wish.PhotoURL = wishInfo.PhotoURL

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

// UpdateWishDonor godoc
// @Summary      点亮心愿
// @Description  为心愿绑定捐赠者并标记为已完成
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
func (c *WishController) UpdateWishDonor(ctx *gin.Context) {
	donorID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(401, utils.CreateResponse(nil, "登录已过期"))
		return
	}
	userType, exists := ctx.Get("userType")
	if !exists || userType != "user" {
		ctx.JSON(401, utils.CreateResponse(nil, "只有用户可以点亮心愿"))
		return
	}

	donor, err := c.userService.GetUserByID(donorID.(uint))
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
		ctx.JSON(400, utils.CreateResponse(nil, "无效的捐赠者信息"))
		return
	}

	wish.DonorName = donorInfo.Name
	wish.DonorMobile = donorInfo.Mobile
	wish.DonorAddress = donorInfo.Address
	wish.DonorComment = donorInfo.Comment

	wish.DonorID = &donor.ID
	wish.Donor = donor

	wish.IsDone = true

	if err := c.wishService.UpdateWish(wish); err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "无法更新心愿状态"))
		return
	}

	updatedWish, err := c.wishService.GetWishByID(uint(wishID))
	if err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "无法获取更新后的心愿信息"))
		return
	}

	ctx.JSON(200, utils.CreateResponse(updatedWish))
}

// GetUserDonatedWishes godoc
// @Summary      获取用户点亮的心愿
// @Description  获取当前登录用户点亮的所有心愿
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        page-index    query    int     false  "页码，默认1"
// @Param        page-size     query    int     false  "每页数量，默认10"
// @Success      200  {object}  controllers.GetWishesResponse  "返回用户点亮的心愿列表"
// @Failure      401  {object}  map[string]interface{}  "用户未登录"
// @Failure      500  {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/user/wishes [get]
func (c *WishController) GetUserDonatedWishes(ctx *gin.Context) {
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

	pageIndexStr := ctx.DefaultQuery("page-index", "1")
	pageIndex, err := strconv.Atoi(pageIndexStr)
	if err != nil || pageIndex < 1 {
		pageIndex = 1
	}

	pageSizeStr := ctx.DefaultQuery("page-size", "10")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	wishes, total, err := c.wishService.GetWishesByDonorID(userID.(uint), pageIndex, pageSize)
	if err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "获取心愿列表失败"))
		return
	}

	response := GetWishesResponse{
		Items: wishes,
		Pagination: Pagination{
			Total:     int(total),
			PageIndex: pageIndex,
			PageSize:  pageSize,
		},
	}

	ctx.JSON(200, utils.CreateResponse(response))
}
