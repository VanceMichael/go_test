package service

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"

	"release-manager/internal/config"
	"release-manager/internal/model"
	"release-manager/internal/repository"

	"gopkg.in/yaml.v3"
)

type BuildService struct {
	cfg           *config.Config
	buildTaskRepo *repository.BuildTaskRepository
	opLogRepo     *repository.OperationLogRepository
	bosService    *BOSService
	logger        *config.Logger
}

func NewBuildService(
	cfg *config.Config,
	buildTaskRepo *repository.BuildTaskRepository,
	opLogRepo *repository.OperationLogRepository,
	bosService *BOSService,
	logger *config.Logger,
) *BuildService {
	svc := &BuildService{
		cfg:           cfg,
		buildTaskRepo: buildTaskRepo,
		opLogRepo:     opLogRepo,
		bosService:    bosService,
		logger:        logger,
	}

	// 启动后台任务处理器
	go svc.processTasksLoop()

	return svc
}

type SubmitBuildRequest struct {
	YamlContent string `json:"yamlContent" binding:"required"`
}

func (s *BuildService) Submit(userID uint, req *SubmitBuildRequest, ip string) (*model.BuildTask, error) {
	// 验证YAML格式
	var yamlData map[string]interface{}
	if err := yaml.Unmarshal([]byte(req.YamlContent), &yamlData); err != nil {
		return nil, fmt.Errorf("YAML格式错误: %w", err)
	}

	task := &model.BuildTask{
		UserID:      userID,
		YamlContent: req.YamlContent,
		Status:      model.BuildStatusPending,
	}

	if err := s.buildTaskRepo.Create(task); err != nil {
		return nil, err
	}

	// 记录操作日志
	s.opLogRepo.Create(&model.OperationLog{
		UserID:     userID,
		Action:     "BUILD_SUBMIT",
		TargetType: "BUILD_TASK",
		TargetID:   task.ID,
		Detail:     "提交自定义编译任务",
		IPAddress:  ip,
	})

	s.logger.Infow("Build task submitted", "taskId", task.ID, "userId", userID)

	return task, nil
}

func (s *BuildService) GetTasks(userID uint, page, pageSize int) ([]model.BuildTask, int64, error) {
	return s.buildTaskRepo.FindByUserID(userID, page, pageSize)
}

func (s *BuildService) GetTask(taskID uint) (*model.BuildTask, error) {
	return s.buildTaskRepo.FindByID(taskID)
}

func (s *BuildService) processTasksLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.processPendingTasks()
	}
}

func (s *BuildService) processPendingTasks() {
	tasks, err := s.buildTaskRepo.FindPending()
	if err != nil {
		s.logger.Errorw("Failed to find pending tasks", "error", err)
		return
	}

	for _, task := range tasks {
		s.processTask(&task)
	}
}

func (s *BuildService) processTask(task *model.BuildTask) {
	s.logger.Infow("Processing build task", "taskId", task.ID)

	// 更新状态为构建中
	task.Status = model.BuildStatusBuilding
	s.buildTaskRepo.Update(task)

	// 写入临时YAML文件
	tmpFile, err := os.CreateTemp("", "build-*.yaml")
	if err != nil {
		s.failTask(task, fmt.Sprintf("创建临时文件失败: %v", err))
		return
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(task.YamlContent); err != nil {
		s.failTask(task, fmt.Sprintf("写入YAML失败: %v", err))
		return
	}
	tmpFile.Close()

	// 执行构建工具
	outputFile := fmt.Sprintf("/tmp/build-output-%d", task.ID)
	cmd := exec.Command(s.cfg.BuildTool, "-c", tmpFile.Name(), "-o", outputFile)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		s.failTask(task, fmt.Sprintf("构建失败: %v\n%s", err, stderr.String()))
		return
	}

	// 上传结果到BOS
	resultFile, err := os.Open(outputFile)
	if err != nil {
		s.failTask(task, fmt.Sprintf("打开构建结果失败: %v", err))
		return
	}
	defer resultFile.Close()
	defer os.Remove(outputFile)

	stat, _ := resultFile.Stat()
	result, err := s.bosService.Upload(resultFile, fmt.Sprintf("custom-build-%d.zip", task.ID), stat.Size(), fmt.Sprintf("custom/%d", task.UserID))
	if err != nil {
		s.failTask(task, fmt.Sprintf("上传结果失败: %v", err))
		return
	}

	// 更新任务状态
	now := time.Now()
	task.Status = model.BuildStatusSuccess
	task.ResultPath = result.BOSPath
	task.CompletedAt = &now
	s.buildTaskRepo.Update(task)

	s.logger.Infow("Build task completed", "taskId", task.ID, "resultPath", result.BOSPath)
}

func (s *BuildService) failTask(task *model.BuildTask, errMsg string) {
	now := time.Now()
	task.Status = model.BuildStatusFailed
	task.ErrorMessage = errMsg
	task.CompletedAt = &now
	s.buildTaskRepo.Update(task)

	s.logger.Errorw("Build task failed", "taskId", task.ID, "error", errMsg)
}
