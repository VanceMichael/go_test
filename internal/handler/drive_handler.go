package handler

import (
	"net/http"
	"strconv"

	"release-manager/internal/config"
	"release-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type DriveHandler struct {
	driveSvc *service.DriveService
	logger   *config.Logger
}

func NewDriveHandler(driveSvc *service.DriveService, logger *config.Logger) *DriveHandler {
	return &DriveHandler{
		driveSvc: driveSvc,
		logger:   logger,
	}
}

func (h *DriveHandler) GetPersonalFiles(c *gin.Context) {
	userID := c.GetUint("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	files, total, err := h.driveSvc.GetPersonalFiles(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  files,
		"total": total,
		"page":  page,
	})
}

func (h *DriveHandler) GetPublicFiles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	files, total, err := h.driveSvc.GetPublicFiles(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  files,
		"total": total,
		"page":  page,
	})
}

func (h *DriveHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文件"})
		return
	}
	defer file.Close()

	isPublic := c.PostForm("isPublic") == "true"
	userID := c.GetUint("userID")
	ip := c.ClientIP()

	req := &service.UploadFileRequest{
		FileName: header.Filename,
		FileSize: header.Size,
		IsPublic: isPublic,
		Reader:   file,
	}

	result, err := h.driveSvc.Upload(userID, req, ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *DriveHandler) Delete(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	userID := c.GetUint("userID")
	ip := c.ClientIP()

	if err := h.driveSvc.Delete(userID, uint(fileID), ip); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *DriveHandler) GetFileURL(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	userID := c.GetUint("userID")

	resp, err := h.driveSvc.GetFileURL(userID, uint(fileID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
