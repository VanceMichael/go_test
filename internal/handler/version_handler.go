package handler

import (
	"net/http"
	"strconv"

	"release-manager/internal/config"
	"release-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type VersionHandler struct {
	versionSvc *service.VersionService
	logger     *config.Logger
}

func NewVersionHandler(versionSvc *service.VersionService, logger *config.Logger) *VersionHandler {
	return &VersionHandler{
		versionSvc: versionSvc,
		logger:     logger,
	}
}

func (h *VersionHandler) GetDirectories(c *gin.Context) {
	tree, err := h.versionSvc.GetDirectoryTree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取目录失败"})
		return
	}
	c.JSON(http.StatusOK, tree)
}

func (h *VersionHandler) GetVersions(c *gin.Context) {
	dirID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的目录ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	versions, total, err := h.versionSvc.GetVersions(uint(dirID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取版本列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  versions,
		"total": total,
		"page":  page,
	})
}

func (h *VersionHandler) GetDownloadURL(c *gin.Context) {
	versionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的版本ID"})
		return
	}

	resp, err := h.versionSvc.GetDownloadURL(uint(versionID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取下载链接失败"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *VersionHandler) CompareVersions(c *gin.Context) {
	versionID1, err := strconv.ParseUint(c.Query("version1"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的版本ID1"})
		return
	}

	versionID2, err := strconv.ParseUint(c.Query("version2"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的版本ID2"})
		return
	}

	if versionID1 == versionID2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能对比同一个版本"})
		return
	}

	resp, err := h.versionSvc.CompareVersions(uint(versionID1), uint(versionID2))
	if err != nil {
		h.logger.Errorw("Failed to compare versions", "error", err, "version1", versionID1, "version2", versionID2)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
