package handler

import (
	"net/http"
	"strconv"

	"release-manager/internal/config"
	"release-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminSvc *service.AdminService
	logger   *config.Logger
}

func NewAdminHandler(adminSvc *service.AdminService, logger *config.Logger) *AdminHandler {
	return &AdminHandler{
		adminSvc: adminSvc,
		logger:   logger,
	}
}

func (h *AdminHandler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	users, total, err := h.adminSvc.GetUsers(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  users,
		"total": total,
		"page":  page,
	})
}

func (h *AdminHandler) SetAdmin(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req struct {
		IsAdmin bool `json:"isAdmin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	operatorID := c.GetUint("userID")
	ip := c.ClientIP()

	if err := h.adminSvc.SetAdmin(uint(userID), req.IsAdmin, operatorID, ip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

func (h *AdminHandler) CreateDirectory(c *gin.Context) {
	var req service.CreateDirectoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	operatorID := c.GetUint("userID")
	ip := c.ClientIP()

	dir, err := h.adminSvc.CreateDirectory(&req, operatorID, ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dir)
}

func (h *AdminHandler) UpdateDirectory(c *gin.Context) {
	dirID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的目录ID"})
		return
	}

	var req service.UpdateDirectoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	operatorID := c.GetUint("userID")
	ip := c.ClientIP()

	if err := h.adminSvc.UpdateDirectory(uint(dirID), &req, operatorID, ip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func (h *AdminHandler) DeleteDirectory(c *gin.Context) {
	dirID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的目录ID"})
		return
	}

	operatorID := c.GetUint("userID")
	ip := c.ClientIP()

	if err := h.adminSvc.DeleteDirectory(uint(dirID), operatorID, ip); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *AdminHandler) UploadVersion(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文件"})
		return
	}
	defer file.Close()

	dirID, _ := strconv.ParseUint(c.PostForm("directoryId"), 10, 32)
	versionName := c.PostForm("versionName")
	description := c.PostForm("description")

	if dirID == 0 || versionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
		return
	}

	operatorID := c.GetUint("userID")
	ip := c.ClientIP()

	req := &service.UploadVersionRequest{
		DirectoryID: uint(dirID),
		VersionName: versionName,
		Description: description,
		FileName:    header.Filename,
		FileSize:    header.Size,
		Reader:      file,
	}

	version, err := h.adminSvc.UploadVersion(req, operatorID, ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, version)
}

func (h *AdminHandler) DeleteVersion(c *gin.Context) {
	versionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的版本ID"})
		return
	}

	operatorID := c.GetUint("userID")
	ip := c.ClientIP()

	if err := h.adminSvc.DeleteVersion(uint(versionID), operatorID, ip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *AdminHandler) SetBaseline(c *gin.Context) {
	var req service.SetBaselineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	operatorID := c.GetUint("userID")
	ip := c.ClientIP()

	if err := h.adminSvc.SetBaseline(&req, operatorID, ip); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "设置成功"})
}

func (h *AdminHandler) GetLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	action := c.Query("action")

	var userID *uint
	if uid := c.Query("userId"); uid != "" {
		if id, err := strconv.ParseUint(uid, 10, 32); err == nil {
			u := uint(id)
			userID = &u
		}
	}

	logs, total, err := h.adminSvc.GetLogs(page, pageSize, userID, action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取日志失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  logs,
		"total": total,
		"page":  page,
	})
}
