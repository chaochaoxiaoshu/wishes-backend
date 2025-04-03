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
}

func NewWishController(db *gorm.DB) *WishController {
	return &WishController{
		wishService: services.NewWishService(db),
	}
}

// GetWishes godoc
// @Summary      获取心愿列表
// @Description  获取心愿列表，支持分页和过滤
// @Tags         心愿
// @Accept       json
// @Produce      json
// @Param        child-name   query     string  false  "按姓名模糊搜索"
// @Param        content      query     string  false  "按心愿内容模糊搜索"
// @Param        is-done      query     bool    false  "按完成状态过滤,默认为false"  default(false)
// @Param        page-index   query     int     false  "页码，默认1"  default(1)
// @Param        page-size    query     int     false  "每页数量，默认10"  default(10)
// @Success      200  {object}  map[string]interface{}  "返回心愿列表和分页信息"
// @Failure      500  {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/wishes [get]
func (c *WishController) GetWishes(ctx *gin.Context) {
	childName := ctx.Query("child-name")
	content := ctx.Query("content")
	isDoneStr := ctx.DefaultQuery("is-done", "false")
	pageIndexStr := ctx.DefaultQuery("page-index", "1")
	pageSizeStr := ctx.DefaultQuery("page-size", "10")

	pageIndex, err := strconv.Atoi(pageIndexStr)
	if err != nil || pageIndex < 1 {
		pageIndex = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	filters := map[string]any{
		"childName": childName,
		"content":   content,
		"isDone":    isDoneStr,
		"pageIndex": pageIndex,
		"pageSize":  pageSize,
	}

	wishes, total, err := c.wishService.FindWishes(filters)
	if err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "获取心愿列表失败"))
		return
	}

	ctx.JSON(200, utils.CreateResponse(gin.H{
		"items": wishes,
		"pagination": gin.H{
			"total":     total,
			"pageIndex": pageIndex,
			"pageSize":  pageSize,
			"pageCount": int(math.Ceil(float64(total) / float64(pageSize))),
		},
	}))
}

// CreateWish godoc
// @Summary      创建新心愿
// @Description  创建一个新的心愿
// @Tags         心愿
// @Accept       json
// @Produce      json
// @Param        wish  body      models.Wish  true  "心愿信息"
// @Success      201   {object}  map[string]interface{}  "返回创建的心愿"
// @Failure      400   {object}  map[string]interface{}  "请求数据无效"
// @Failure      500   {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/wishes [post]
func (c *WishController) CreateWish(ctx *gin.Context) {
	var wish models.Wish
	if err := ctx.ShouldBindJSON(&wish); err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, "无效的请求数据"))
		return
	}

	if err := c.wishService.CreateWish(&wish); err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "无法创建心愿"))
		return
	}

	ctx.JSON(201, utils.CreateResponse(wish))
}

// UpdateWishDonor godoc
// @Summary      更新心愿捐赠者
// @Description  为心愿绑定捐赠者并标记为已完成
// @Tags         心愿
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "心愿ID"
// @Param        donor  body      models.User  true  "捐赠者信息"
// @Success      200   {object}  map[string]interface{}  "返回更新后的心愿"
// @Failure      400   {object}  map[string]interface{}  "请求数据无效"
// @Failure      404   {object}  map[string]interface{}  "心愿不存在"
// @Failure      500   {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/wishes/{id}/donor [put]
func (c *WishController) UpdateWishDonor(ctx *gin.Context) {
	wishID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, "无效的心愿ID"))
		return
	}

	var donor models.User
	if err := ctx.ShouldBindJSON(&donor); err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, "无效的捐赠者信息"))
		return
	}

	wish, err := c.wishService.GetWishByID(uint(wishID))
	if err != nil {
		ctx.JSON(404, utils.CreateResponse(nil, "心愿不存在"))
		return
	}

	if err := c.wishService.CreateOrUpdateUser(&donor); err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "无法保存捐赠者信息"))
		return
	}

	wish.IsDone = true
	wish.DonorID = &donor.ID

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
