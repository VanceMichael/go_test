package handler

import (
	"net/http"
	"strconv"

	"release-manager/internal/config"
	"release-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type BuildHandler struct {
	buildSvc *service.BuildService
	logger   *config.Logger
}

func NewBuildHandler(buildSvc *service.BuildService, logger *config.Logger) *BuildHandler {
	return &BuildHandler{
		buildSvc: buildSvc,
		logger:   logger,
	}
}

func (h *BuildHandler) Submit(c *gin.Context) {
	var req service.SubmitBuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	userID := c.GetUint("userID")
	ip := c.ClientIP()

	task, err := h.buildSvc.Submit(userID, &req, ip)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *BuildHandler) GetTasks(c *gin.Context) {
	userID := c.GetUint("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	tasks, total, err := h.buildSvc.GetTasks(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  tasks,
		"total": total,
		"page":  page,
	})
}

func (h *BuildHandler) GetTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	task, err := h.buildSvc.GetTask(uint(taskID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, task)
}
