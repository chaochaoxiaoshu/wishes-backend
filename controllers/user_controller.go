package controllers

import (
	"strconv"
	"wishes/models"
	"wishes/services"
	"wishes/utils"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// GetAdminUsersResponse 获取管理员用户列表响应
type GetAdminUsersResponse struct {
	Items      []models.User    `json:"items"`
	Pagination utils.Pagination `json:"pagination"`
}

// GetAdminUsers godoc
// @Summary      获取所有拥有管理员权限的用户
// @Description  获取所有具有管理员权限的用户
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        pageIndex    query    int     false  "页码，默认1"
// @Param        pageSize     query    int     false  "每页数量，默认10"
// @Success      200  {object}  controllers.GetAdminUsersResponse  "返回管理员用户列表"
// @Failure      401  {object}  map[string]interface{}  "用户未登录或无权限"
// @Failure      500  {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/users/admin [get]
func (c *UserController) GetAdminUsers(ctx *gin.Context) {
	userType, exists := ctx.Get("userType")
	if !exists || userType != "admin" {
		ctx.JSON(401, utils.CreateResponse(nil, "只有系统管理员可以查看用户列表"))
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

	users, total, err := c.userService.GetAdminUsers(pageIndex, pageSize)
	if err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "获取管理员用户列表失败"))
		return
	}

	response := GetAdminUsersResponse{
		Items:      users,
		Pagination: utils.NewPagination(total, pageIndex, pageSize),
	}

	ctx.JSON(200, utils.CreateResponse(response))
}

// GetNonAdminUsers godoc
// @Summary      获取所有不拥有管理员权限的用户
// @Description  获取所有不具有管理员权限的普通用户
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        pageIndex    query    int     false  "页码，默认1"
// @Param        pageSize     query    int     false  "每页数量，默认10"
// @Success      200  {object}  controllers.GetAdminUsersResponse  "返回普通用户列表"
// @Failure      401  {object}  map[string]interface{}  "用户未登录或无权限"
// @Failure      500  {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/users/regular [get]
func (c *UserController) GetNonAdminUsers(ctx *gin.Context) {
	userType, exists := ctx.Get("userType")
	if !exists || userType != "admin" {
		ctx.JSON(401, utils.CreateResponse(nil, "只有系统管理员可以查看用户列表"))
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

	users, total, err := c.userService.GetNonAdminUsers(pageIndex, pageSize)
	if err != nil {
		ctx.JSON(500, utils.CreateResponse(nil, "获取普通用户列表失败"))
		return
	}

	response := GetAdminUsersResponse{
		Items:      users,
		Pagination: utils.NewPagination(total, pageIndex, pageSize),
	}

	ctx.JSON(200, utils.CreateResponse(response))
}

// UpdateUserAdminRequest 更新用户管理员权限请求
type UpdateUserAdminRequest struct {
	IsAdmin bool `json:"isAdmin" binding:"required"`
}

// UpdateUserAdmin godoc
// @Summary      更新用户管理员权限
// @Description  设置或取消用户的管理员权限
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id    path    int     true  "用户ID"
// @Param        request    body    controllers.UpdateUserAdminRequest  true  "请求数据"
// @Success      200  {object}  map[string]interface{}  "更新成功"
// @Failure      400  {object}  map[string]interface{}  "请求数据错误"
// @Failure      401  {object}  map[string]interface{}  "用户未登录或无权限"
// @Failure      404  {object}  map[string]interface{}  "用户不存在"
// @Failure      500  {object}  map[string]interface{}  "服务器错误"
// @Router       /api/v1/users/{id}/admin [put]
func (c *UserController) UpdateUserAdmin(ctx *gin.Context) {
	userType, exists := ctx.Get("userType")
	if !exists || userType != "admin" {
		ctx.JSON(401, utils.CreateResponse(nil, "只有系统管理员可以更新用户权限"))
		return
	}

	userIDStr := ctx.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, "无效的用户ID"))
		return
	}

	var req UpdateUserAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, utils.CreateResponse(nil, "无效的请求数据"))
		return
	}

	if err := c.userService.UpdateUserAdminStatus(uint(userID), req.IsAdmin); err != nil {
		if err.Error() == "未找到ID为"+userIDStr+"的用户" {
			ctx.JSON(404, utils.CreateResponse(nil, "用户不存在"))
		} else {
			ctx.JSON(500, utils.CreateResponse(nil, "更新用户权限失败"))
		}
		return
	}

	message := "更新用户权限成功"
	if req.IsAdmin {
		message = "已将用户设置为管理员"
	} else {
		message = "已取消用户的管理员权限"
	}

	ctx.JSON(200, utils.CreateResponse(nil, message))
}
