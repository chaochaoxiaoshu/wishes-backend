package controllers

import (
	"wishes/services"
	"wishes/utils"

	"github.com/gin-gonic/gin"
)

type UploadController struct {
	storageService *services.StorageService
}

// NewUploadController 创建上传控制器实例
func NewUploadController(storageService *services.StorageService) *UploadController {
	return &UploadController{
		storageService: storageService,
	}
}

// UploadImageResponse 上传图片响应数据
type UploadImageResponse struct {
	URL string `json:"url"` // 上传成功后的图片URL
}

// UploadImage godoc
// @Summary      上传图片
// @Description  上传图片到腾讯云对象存储
// @Tags         文件上传
// @Accept       multipart/form-data
// @Produce      json
// @Param        file    formData    file     true  "图片文件"
// @Param        directory  formData    string  false  "存储目录，例如: images/avatar"
// @Success      200  {object}  controllers.UploadImageResponse  "上传成功，返回图片URL"
// @Failure      400  {object}  map[string]interface{}  "请求参数错误"
// @Failure      401  {object}  map[string]interface{}  "用户未登录"
// @Failure      500  {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/upload/image [post]
func (c *UploadController) UploadImage(ctx *gin.Context) {
	// 检查是否已登录
	_, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(401, utils.CreateResponse(nil, "请先登录"))
		return
	}

	// 获取上传文件
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, "请选择要上传的图片"))
		return
	}

	// 检查文件类型
	contentType := file.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/gif" && contentType != "image/webp" {
		ctx.JSON(400, utils.CreateResponse(nil, "只支持上传JPG、PNG、GIF或WEBP格式的图片"))
		return
	}

	// 检查文件大小（限制为5MB）
	if file.Size > 5*1024*1024 {
		ctx.JSON(400, utils.CreateResponse(nil, "图片大小不能超过5MB"))
		return
	}

	// 获取存储目录参数，默认为"images"
	directory := ctx.DefaultPostForm("directory", "images")

	// 上传图片到腾讯云COS
	fileURL, err := c.storageService.UploadImage(file, directory)
	if err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "上传图片失败: "+err.Error()))
		return
	}

	// 返回成功响应和图片URL
	response := UploadImageResponse{
		URL: fileURL,
	}
	ctx.JSON(200, utils.CreateResponse(response))
}
