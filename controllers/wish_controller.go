package controllers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

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
// @Param        isDone      query     bool    false  "按完成状态过滤,默认为false"  default(false)
// @Param        isPublished query     bool    false  "按公开状态过滤,不传为全部"
// @Param        pageIndex   query     int     false  "页码，默认1"  default(1)
// @Param        pageSize    query     int     false  "每页数量，默认10"  default(10)
// @Success      200  {object}  GetWishesResponse  "返回心愿列表和分页信息"
// @Failure      500  {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/wishes [get]
func (c *WishController) GetWishes(ctx *gin.Context) {
	content := ctx.Query("content")
	isDoneStr := ctx.DefaultQuery("isDone", "false")
	isPublishedStr := ctx.Query("isPublished") // 不设置默认值，不传表示全部
	pageIndexStr := ctx.DefaultQuery("pageIndex", "1")
	pageSizeStr := ctx.DefaultQuery("pageSize", "10")

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
// @Description  批量导入多个心愿，支持JSON和XLSX文件
// @Tags         心愿
// @Accept       multipart/form-data
// @Accept       json
// @Produce      json
// @Param        request  body   BatchCreateWishRequest  false  "JSON格式的心愿信息数组"
// @Param        file     formData  file  false  "Excel文件，支持 .xlsx 和 .xls"
// @Success      201   {object}  map[string]interface{}  "返回导入结果"
// @Failure      400   {object}  map[string]interface{}  "请求数据无效"
// @Failure      500   {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/wishes/batch [post]
func (c *WishController) BatchCreateWishes(ctx *gin.Context) {
	contentType := ctx.GetHeader("Content-Type")

	var wishes []*models.Wish

	if strings.Contains(contentType, "application/json") {
		var wishRequest BatchCreateWishRequest
		if err := ctx.ShouldBindJSON(&wishRequest); err != nil {
			ctx.JSON(400, utils.CreateResponse(nil, "无效的请求数据"))
			return
		}

		if len(wishRequest.Data) == 0 {
			ctx.JSON(400, utils.CreateResponse(nil, "导入列表不能为空"))
			return
		}

		wishes = make([]*models.Wish, 0, len(wishRequest.Data))
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
	} else if strings.Contains(contentType, "multipart/form-data") {
		file, err := ctx.FormFile("file")
		if err != nil {
			ctx.JSON(400, utils.CreateResponse(nil, "无法获取上传的文件"))
			return
		}

		fileName := strings.ToLower(file.Filename)
		if !strings.HasSuffix(fileName, ".xlsx") && !strings.HasSuffix(fileName, ".xls") {
			ctx.JSON(400, utils.CreateResponse(nil, "仅支持Excel文件格式(.xlsx或.xls)"))
			return
		}

		fileContent, err := file.Open()
		if err != nil {
			ctx.JSON(500, utils.CreateResponse(nil, "无法打开上传文件"))
			return
		}
		defer fileContent.Close()

		xlsx, err := excelize.OpenReader(fileContent)
		if err != nil {
			ctx.JSON(500, utils.CreateResponse(nil, "解析Excel文件失败"))
			return
		}
		defer xlsx.Close()

		wishes = make([]*models.Wish, 0)
		sheetList := xlsx.GetSheetList()

		for _, sheetName := range sheetList {
			rows, err := xlsx.GetRows(sheetName)
			if err != nil {
				ctx.JSON(500, utils.CreateResponse(nil, fmt.Sprintf("读取sheet '%s' 失败", sheetName)))
				return
			}

			// 跳过空sheet
			if len(rows) <= 1 {
				continue
			}

			// 查找标题行并确定列索引
			headerRow := rows[0]
			nameColIndex := -1
			genderColIndex := -1
			contentColIndex := -1
			reasonColIndex := -1

			// 寻找必要的列
			for i, cell := range headerRow {
				cell = strings.TrimSpace(strings.ToLower(cell))
				switch {
				case cell == "姓名" || cell == "name" || cell == "学生姓名" || cell == "儿童姓名" || cell == "childname":
					nameColIndex = i
				case cell == "性别" || cell == "gender" || cell == "sex":
					genderColIndex = i
				case cell == "心愿" || cell == "wish" || cell == "愿望" || cell == "心愿内容" || cell == "content" || cell == "wishcontent":
					contentColIndex = i
				case cell == "理由" || cell == "原因" || cell == "reason" || cell == "wish reason" || cell == "心愿理由":
					reasonColIndex = i
				}
			}

			// 检查必要的列是否都找到了
			if nameColIndex == -1 || genderColIndex == -1 || contentColIndex == -1 || reasonColIndex == -1 {
				ctx.JSON(400, utils.CreateResponse(nil, fmt.Sprintf(
					"sheet '%s' 中未找到必要的列。需要包含姓名、性别、心愿和理由列", sheetName)))
				return
			}

			// 处理数据行
			for i := 1; i < len(rows); i++ {
				row := rows[i]
				// 确保行有足够的数据
				if len(row) <= max(nameColIndex, genderColIndex, contentColIndex, reasonColIndex) {
					continue
				}

				childName := strings.TrimSpace(row[nameColIndex])
				genderStr := strings.TrimSpace(row[genderColIndex])
				content := strings.TrimSpace(row[contentColIndex])
				reason := strings.TrimSpace(row[reasonColIndex])

				// 跳过空行
				if childName == "" || genderStr == "" || content == "" || reason == "" {
					continue
				}

				// 解析性别
				var gender models.Gender
				if genderStr == "男" {
					gender = "male"
				} else if genderStr == "女" {
					gender = "female"
				}

				wish := &models.Wish{
					ChildName:   childName,
					Gender:      gender,
					Content:     content,
					Reason:      reason,
					IsPublished: true,
				}
				wishes = append(wishes, wish)
			}
		}

		if len(wishes) == 0 {
			ctx.JSON(400, utils.CreateResponse(nil, "Excel文件中没有有效的心愿数据"))
			return
		}
	} else {
		ctx.JSON(400, utils.CreateResponse(nil, "不支持的Content-Type，请使用application/json或multipart/form-data"))
		return
	}

	if err := c.wishService.BatchCreateWishes(wishes); err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "批量导入心愿失败"))
		return
	}

	ctx.JSON(201, utils.CreateResponse("批量导入心愿成功"))
}

// 辅助函数，用于找出最大的索引值
func max(values ...int) int {
	maxVal := values[0]
	for _, v := range values[1:] {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}
