package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"wishes/middleware"
	"wishes/models"
	"wishes/services"
	"wishes/utils"
)

type AuthController struct {
	DB            *gorm.DB
	WechatService *services.WechatService
}

func NewAuthController(db *gorm.DB, wechatService *services.WechatService) *AuthController {
	return &AuthController{
		DB:            db,
		WechatService: wechatService,
	}
}

type AdminRegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminRegister godoc
// @Summary 管理员注册
// @Description 创建新管理员账号
// @Tags 管理员
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body AdminRegisterRequest true "管理员注册信息"
// @Success 201 {object} models.Admin
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 403 {object} map[string]interface{} "禁止访问"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/v1/admin/register [post]
func (c *AuthController) AdminRegister(ctx *gin.Context) {
	var req AdminRegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.CreateResponse(nil, "无效的请求参数"))
		return
	}

	var existingAdmin models.Admin
	if result := c.DB.Where("username = ?", req.Username).First(&existingAdmin); result.Error == nil {
		ctx.JSON(http.StatusBadRequest, utils.CreateResponse(nil, "管理员用户名已存在"))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.CreateResponse(nil, "密码加密失败"))
		return
	}

	admin := models.Admin{
		Username: req.Username,
		Password: string(hashedPassword),
	}

	if result := c.DB.Create(&admin); result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, utils.CreateResponse(nil, "创建管理员失败"))
		return
	}

	admin.Password = ""
	ctx.JSON(http.StatusCreated, utils.CreateResponse(admin))
}

type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminLoginResponse struct {
	Token string       `json:"token"`
	Admin models.Admin `json:"admin"`
}

// AdminLogin godoc
// @Summary 管理员登录
// @Description 管理员登录并获取认证令牌
// @Tags 管理员
// @Accept json
// @Produce json
// @Param request body AdminLoginRequest true "管理员登录信息"
// @Success 200 {object} AdminLoginResponse
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/v1/admin/login [post]
func (c *AuthController) AdminLogin(ctx *gin.Context) {
	var req AdminLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.CreateResponse(nil, "无效的请求参数"))
		return
	}

	var admin models.Admin
	if result := c.DB.Where("username = ?", req.Username).First(&admin); result.Error != nil {
		ctx.JSON(http.StatusUnauthorized, utils.CreateResponse(nil, "用户名或密码错误"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.CreateResponse(nil, "用户名或密码错误"))
		return
	}

	token, err := middleware.GenerateAdminToken(admin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.CreateResponse(nil, "生成令牌失败"))
		return
	}

	admin.Password = ""
	ctx.JSON(http.StatusOK, utils.CreateResponse(AdminLoginResponse{
		Token: token,
		Admin: admin,
	}))
}

type WechatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

type WechatLoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// WechatLogin godoc
// @Summary 微信小程序登录
// @Description 通过微信小程序临时登录凭证code进行登录
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body WechatLoginRequest true "微信登录请求"
// @Success 200 {object} WechatLoginResponse
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/v1/user/login [post]
func (c *AuthController) WechatLogin(ctx *gin.Context) {
	var req WechatLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.CreateResponse(nil, "无效的请求参数"))
		return
	}

	token, user, err := c.WechatService.Login(req.Code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.CreateResponse(nil, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.CreateResponse(WechatLoginResponse{
		Token: token,
		User:  *user,
	}))
}

type WechatUserInfoRequest struct {
	Nickname  string `json:"nickName"`
	AvatarURL string `json:"avatarUrl"`
}

// UpdateWechatUserInfo godoc
// @Summary 更新微信用户信息
// @Description 更新微信用户的昵称和头像
// @Tags 用户
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body WechatUserInfoRequest true "微信用户信息"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/v1/user/userinfo [put]
func (c *AuthController) UpdateWechatUserInfo(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, utils.CreateResponse(nil, "未认证"))
		return
	}

	var req WechatUserInfoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.CreateResponse(nil, "无效的请求参数"))
		return
	}

	if err := c.WechatService.UpdateUserInfo(userID.(uint), req.Nickname, req.AvatarURL); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.CreateResponse(nil, "更新用户信息失败"))
		return
	}

	var user models.User
	if err := c.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.CreateResponse(nil, "获取用户信息失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.CreateResponse(user))
}
