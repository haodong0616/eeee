package queue

import (
	"expchange-backend/database"
	"expchange-backend/models"
	"expchange-backend/services"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	_ "expchange-backend/config" // 确保config包被初始化
)

// TaskType 任务类型
type TaskType string

const (
	TaskGenerateTrades  TaskType = "generate_trades"
	TaskGenerateKlines  TaskType = "generate_klines"
	TaskVerifyDeposit   TaskType = "verify_deposit"
	TaskProcessWithdraw TaskType = "process_withdraw"
)

// Task 任务
type Task struct {
	ID         string
	Type       TaskType
	Status     string     // pending, running, completed, failed
	Symbol     string     // 交易对符号（用于交易数据生成）
	RecordID   string     // 关联记录ID（用于充值/提现）
	RecordType string     // 记录类型：deposit, withdraw
	StartTime  *time.Time // 开始时间
	EndTime    *time.Time // 结束时间
	Message    string
	CreatedAt  time.Time
	StartedAt  *time.Time
	EndedAt    *time.Time
	Error      string
}

// TaskQueue 任务队列
type TaskQueue struct {
	tasks             map[string]*Task
	queue             chan *Task
	mu                sync.RWMutex
	running           bool
	maxWorkers        int
	workers           int // 当前运行的worker数
	depositVerifier   *services.DepositVerifier
	withdrawProcessor *services.WithdrawProcessor
}

var (
	instance *TaskQueue
	once     sync.Once
)

// GetQueue 获取任务队列单例
func GetQueue() *TaskQueue {
	once.Do(func() {
		// 初始化充值验证器
		depositVerifier, err := services.NewDepositVerifier()
		if err != nil {
			log.Printf("⚠️  充值验证服务初始化失败: %v", err)
		}

		// 初始化提现处理器
		withdrawProcessor, err := services.NewWithdrawProcessor()
		if err != nil {
			log.Printf("⚠️  提现处理服务初始化失败: %v", err)
		}

		instance = &TaskQueue{
			tasks:             make(map[string]*Task),
			queue:             make(chan *Task, 100),
			maxWorkers:        0, // 动态设置
			workers:           0,
			depositVerifier:   depositVerifier,
			withdrawProcessor: withdrawProcessor,
		}
		instance.loadFromDB() // 从数据库加载未完成的任务
		instance.Start()
	})
	return instance
}

// loadFromDB 从数据库加载任务到内存（启动时）
func (q *TaskQueue) loadFromDB() {
	var dbTasks []models.Task
	// 只加载未完成的任务（pending 和 running）
	database.DB.Where("status IN ?", []string{"pending", "running"}).Find(&dbTasks)

	for _, dbTask := range dbTasks {
		task := &Task{
			ID:         dbTask.ID,
			Type:       TaskType(dbTask.Type),
			Status:     dbTask.Status,
			Symbol:     dbTask.Symbol,
			RecordID:   dbTask.RecordID,
			RecordType: dbTask.RecordType,
			StartTime:  dbTask.StartTime,
			EndTime:    dbTask.EndTime,
			Message:    dbTask.Message,
			CreatedAt:  dbTask.CreatedAt,
			StartedAt:  dbTask.StartedAt,
			EndedAt:    dbTask.EndedAt,
			Error:      dbTask.Error,
		}
		q.tasks[task.ID] = task

		// 如果是 pending 状态，重新加入队列
		if task.Status == "pending" {
			q.queue <- task
		}
	}

	if len(dbTasks) > 0 {
		log.Printf("📋 从数据库加载了 %d 个任务", len(dbTasks))
	}
}

// Start 启动任务队列
func (q *TaskQueue) Start() {
	if q.running {
		return
	}

	q.running = true

	// 动态读取worker数量
	sysConfig := database.GetSystemConfigManager()
	workers := sysConfig.GetInt("task.queue.workers", 10)
	q.maxWorkers = workers

	log.Printf("📋 任务队列已启动 (%d个数据生成worker)", workers)

	// 启动数据生成worker
	for i := 0; i < q.maxWorkers; i++ {
		go q.worker(i)
	}

	// 启动专门的充值验证worker（单独进程）
	go q.depositWorker()

	// 启动专门的提现处理worker（单独进程）
	go q.withdrawWorker()

	// 启动worker数量监控协程，支持动态调整
	go q.monitorWorkerCount()
}

// Stop 停止任务队列
func (q *TaskQueue) Stop() {
	q.running = false
	close(q.queue)
	log.Println("🛑 任务队列已停止")
}

// monitorWorkerCount 监控并动态调整worker数量
func (q *TaskQueue) monitorWorkerCount() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	for {
		if !q.running {
			return
		}

		select {
		case <-ticker.C:
			sysConfig := database.GetSystemConfigManager()
			newWorkerCount := sysConfig.GetInt("task.queue.workers", 10)

			q.mu.Lock()
			currentWorkers := q.maxWorkers
			q.mu.Unlock()

			if newWorkerCount != currentWorkers {
				log.Printf("🔄 任务队列worker数量变更: %d -> %d", currentWorkers, newWorkerCount)

				q.mu.Lock()
				if newWorkerCount > currentWorkers {
					// 增加worker
					for i := currentWorkers; i < newWorkerCount; i++ {
						go q.worker(i)
					}
				}
				// 注意：减少worker需要worker自然结束，这里只更新maxWorkers
				q.maxWorkers = newWorkerCount
				q.mu.Unlock()
			}
		}
	}
}

// worker 工作协程（只处理数据生成任务）
func (q *TaskQueue) worker(id int) {
	log.Printf("🔧 数据生成 Worker %d 已启动", id)

	for task := range q.queue {
		if !q.running {
			break
		}

		// 只处理数据生成类型的任务
		if task.Type == TaskGenerateTrades || task.Type == TaskGenerateKlines {
			q.processTask(task)
		} else {
			// 其他类型的任务重新放回队列，等待专门的worker处理
			log.Printf("⏭️  Worker %d 跳过非数据生成任务: %s (Type: %s)", id, task.ID, task.Type)
			q.queue <- task
			time.Sleep(100 * time.Millisecond) // 避免忙等待
		}
	}

	log.Printf("🔧 数据生成 Worker %d 已停止", id)
}

// processTask 处理任务
func (q *TaskQueue) processTask(task *Task) {
	q.logTask(task.ID, "info", "task_started", fmt.Sprintf("🚀 开始执行任务: %s", task.Type), "")
	log.Printf("🚀 开始执行任务: %s (ID: %s)", task.Type, task.ID)

	// 更新任务状态为运行中
	now := time.Now()
	task.Status = "running"
	task.StartedAt = &now
	task.Message = "任务执行中..."
	q.updateTask(task)
	q.logTask(task.ID, "info", "status_updated", "任务状态已更新为: running", "")

	// 执行任务
	var err error
	switch task.Type {
	case TaskGenerateTrades:
		q.logTask(task.ID, "info", "execution_started", "开始生成交易数据", fmt.Sprintf("交易对: %s", task.Symbol))
		err = q.executeSeedTrades(task)
	case TaskGenerateKlines:
		q.logTask(task.ID, "info", "execution_started", "开始生成K线数据", fmt.Sprintf("交易对: %s", task.Symbol))
		err = q.executeSeedKlines(task)
	case TaskVerifyDeposit:
		q.logTask(task.ID, "info", "execution_started", "开始验证充值", fmt.Sprintf("充值记录ID: %s", task.RecordID))
		err = q.executeVerifyDeposit(task)
	case TaskProcessWithdraw:
		q.logTask(task.ID, "info", "execution_started", "开始处理提现", fmt.Sprintf("提现记录ID: %s", task.RecordID))
		err = q.executeProcessWithdraw(task)
	default:
		err = fmt.Errorf("unknown task type: %s", task.Type)
		q.logTask(task.ID, "error", "execution_error", "未知的任务类型", string(task.Type))
	}

	// 更新任务状态
	endTime := time.Now()
	task.EndedAt = &endTime
	duration := endTime.Sub(*task.StartedAt)

	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		task.Message = "任务执行失败"
		q.logTask(task.ID, "error", "task_failed",
			fmt.Sprintf("❌ 任务执行失败: %v", err),
			fmt.Sprintf("耗时: %.2f秒", duration.Seconds()))
		log.Printf("❌ 任务执行失败: %s (ID: %s) - %v", task.Type, task.ID, err)
	} else {
		task.Status = "completed"
		task.Message = "任务执行完成"
		q.logTask(task.ID, "info", "task_completed",
			"✅ 任务执行完成",
			fmt.Sprintf("耗时: %.2f秒", duration.Seconds()))
		log.Printf("✅ 任务执行完成: %s (ID: %s), 耗时: %.2f秒", task.Type, task.ID, duration.Seconds())
	}

	q.updateTask(task)
	q.logTask(task.ID, "info", "task_finalized", "任务状态已保存", fmt.Sprintf("最终状态: %s", task.Status))
}

// executeSeedTrades 执行生成交易数据
func (q *TaskQueue) executeSeedTrades(task *Task) error {
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("生成交易数据 panic: %v", r)
			q.logTask(task.ID, "error", "panic_recovered", errMsg, "")
			log.Printf("❌ %s", errMsg)
		}
	}()

	if task.Symbol == "" {
		q.logTask(task.ID, "error", "validation_failed", "交易对符号为空", "")
		return fmt.Errorf("symbol is required for trade data generation")
	}

	// 使用时间范围生成数据
	startTime := task.StartTime
	endTime := task.EndTime
	if startTime == nil {
		// 默认6-12个月前
		now := time.Now()
		monthsBack := 6 + (time.Now().UnixNano() % 7)
		defaultStart := now.AddDate(0, -int(monthsBack), 0)
		startTime = &defaultStart
		q.logTask(task.ID, "info", "default_time_range",
			"使用默认时间范围",
			fmt.Sprintf("开始时间: %s", startTime.Format("2006-01-02")))
	}
	if endTime == nil {
		now := time.Now()
		endTime = &now
		q.logTask(task.ID, "info", "default_time_range",
			"使用默认结束时间",
			fmt.Sprintf("结束时间: %s", endTime.Format("2006-01-02")))
	}

	q.logTask(task.ID, "info", "data_generation_started",
		"开始生成交易数据",
		fmt.Sprintf("交易对: %s, 时间范围: %s ~ %s",
			task.Symbol,
			startTime.Format("2006-01-02"),
			endTime.Format("2006-01-02")))

	database.SeedTradesForSymbolWithTimeRange(task.Symbol, *startTime, *endTime)

	q.logTask(task.ID, "info", "data_generation_completed",
		"交易数据生成完成",
		fmt.Sprintf("交易对: %s", task.Symbol))

	return nil
}

// executeSeedKlines 执行生成K线数据
func (q *TaskQueue) executeSeedKlines(task *Task) error {
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("生成K线数据 panic: %v", r)
			q.logTask(task.ID, "error", "panic_recovered", errMsg, "")
			log.Printf("❌ %s", errMsg)
		}
	}()

	if task.Symbol == "" {
		q.logTask(task.ID, "error", "validation_failed", "交易对符号为空", "")
		return fmt.Errorf("symbol is required for kline data generation")
	}

	q.logTask(task.ID, "info", "kline_generation_started",
		"开始生成K线数据",
		fmt.Sprintf("交易对: %s", task.Symbol))

	database.SeedKlinesForSymbol(task.Symbol)

	q.logTask(task.ID, "info", "kline_generation_completed",
		"K线数据生成完成",
		fmt.Sprintf("交易对: %s", task.Symbol))

	return nil
}

// depositWorker 专门处理充值验证任务的worker（单独进程）
func (q *TaskQueue) depositWorker() {
	log.Printf("💰 充值验证 Worker 已启动（独立进程）")

	for {
		if !q.running {
			break
		}

		// 从队列中获取任务
		select {
		case task := <-q.queue:
			if task.Type == TaskVerifyDeposit {
				q.processTask(task)
			} else {
				// 不是充值任务，放回队列
				q.queue <- task
			}
		case <-time.After(1 * time.Second):
			// 超时，继续循环
		}
	}

	log.Printf("💰 充值验证 Worker 已停止")
}

// withdrawWorker 专门处理提现任务的worker（单独进程，带Nonce管理）
func (q *TaskQueue) withdrawWorker() {
	log.Printf("💸 提现处理 Worker 已启动（独立进程，线程安全）")

	for {
		if !q.running {
			break
		}

		// 从队列中获取任务（串行处理，保证nonce顺序）
		select {
		case task := <-q.queue:
			if task.Type == TaskProcessWithdraw {
				q.processTask(task)
			} else {
				// 不是提现任务，放回队列
				q.queue <- task
			}
		case <-time.After(1 * time.Second):
			// 超时，继续循环
		}
	}

	log.Printf("💸 提现处理 Worker 已停止")
}

// executeVerifyDeposit 执行充值验证（调用充值验证服务）
func (q *TaskQueue) executeVerifyDeposit(task *Task) error {
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("充值验证 panic: %v", r)
			q.logTask(task.ID, "error", "panic_recovered", errMsg, "")
			log.Printf("❌ %s", errMsg)
		}
	}()

	if task.RecordID == "" {
		q.logTask(task.ID, "error", "validation_failed", "充值记录ID为空", "")
		return fmt.Errorf("record_id is required for deposit verification")
	}

	// 获取充值记录
	var deposit models.DepositRecord
	if err := database.DB.Where("id = ?", task.RecordID).First(&deposit).Error; err != nil {
		q.logTask(task.ID, "error", "record_not_found",
			fmt.Sprintf("充值记录不存在: %v", err),
			fmt.Sprintf("记录ID: %s", task.RecordID))
		return fmt.Errorf("deposit record not found: %w", err)
	}

	q.logTask(task.ID, "info", "deposit_record_loaded",
		"充值记录已加载",
		fmt.Sprintf("TxHash: %s, Amount: %s %s, Chain: %s",
			deposit.TxHash, deposit.Amount.String(), deposit.Asset, deposit.Chain))

	// 调用充值验证服务
	if q.depositVerifier == nil {
		q.logTask(task.ID, "error", "service_unavailable", "充值验证服务未初始化", "")
		return fmt.Errorf("deposit verifier not available")
	}

	q.logTask(task.ID, "info", "chain_verification_started", "开始链上验证", "")

	// 调用充值验证服务（返回error表示需要重试）
	verifyErr := q.depositVerifier.VerifyDeposit(&deposit)

	// 检查是否需要重试
	if verifyErr != nil && strings.Contains(verifyErr.Error(), "RETRY_LATER") {
		q.logTask(task.ID, "info", "retry_scheduled",
			"交易未确认，10秒后重试",
			fmt.Sprintf("原因: %s", verifyErr.Error()))

		// 10秒后重新加入队列
		go func() {
			time.Sleep(10 * time.Second)
			q.queue <- task
			log.Printf("🔄 充值验证任务已重新加入队列: TaskID=%s, RecordID=%s", task.ID, task.RecordID)
		}()

		return fmt.Errorf("RETRY_SCHEDULED: %w", verifyErr)
	}

	// 重新加载充值记录，检查验证结果
	if err := database.DB.Where("id = ?", task.RecordID).First(&deposit).Error; err != nil {
		q.logTask(task.ID, "error", "reload_failed",
			fmt.Sprintf("重新加载充值记录失败: %v", err), "")
		return fmt.Errorf("failed to reload deposit record: %w", err)
	}

	if deposit.Status == "confirmed" {
		q.logTask(task.ID, "info", "deposit_verification_completed",
			"充值验证成功",
			fmt.Sprintf("TxHash: %s, Amount: %s %s",
				deposit.TxHash, deposit.Amount.String(), deposit.Asset))
		return nil
	} else if deposit.Status == "failed" {
		q.logTask(task.ID, "error", "verification_failed",
			"充值验证失败",
			fmt.Sprintf("TxHash: %s, 原因: 请查看充值记录", deposit.TxHash))
		return fmt.Errorf("deposit verification failed")
	}

	return nil
}

// executeProcessWithdraw 执行提现处理（调用提现处理服务，使用Nonce管理器）
func (q *TaskQueue) executeProcessWithdraw(task *Task) error {
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("提现处理 panic: %v", r)
			q.logTask(task.ID, "error", "panic_recovered", errMsg, "")
			log.Printf("❌ %s", errMsg)
		}
	}()

	if task.RecordID == "" {
		q.logTask(task.ID, "error", "validation_failed", "提现记录ID为空", "")
		return fmt.Errorf("record_id is required for withdraw processing")
	}

	// 获取提现记录
	var withdrawal models.WithdrawRecord
	if err := database.DB.Where("id = ?", task.RecordID).First(&withdrawal).Error; err != nil {
		q.logTask(task.ID, "error", "record_not_found",
			fmt.Sprintf("提现记录不存在: %v", err),
			fmt.Sprintf("记录ID: %s", task.RecordID))
		return fmt.Errorf("withdraw record not found: %w", err)
	}

	q.logTask(task.ID, "info", "withdraw_record_loaded",
		"提现记录已加载",
		fmt.Sprintf("Address: %s, Amount: %s %s, Chain: %s",
			withdrawal.Address, withdrawal.Amount.String(), withdrawal.Asset, withdrawal.Chain))

	// 调用提现处理服务
	if q.withdrawProcessor == nil {
		q.logTask(task.ID, "error", "service_unavailable", "提现处理服务未初始化", "")
		return fmt.Errorf("withdraw processor not available")
	}

	q.logTask(task.ID, "info", "withdraw_processing_started",
		"开始处理提现转账",
		fmt.Sprintf("使用Nonce管理器确保线程安全"))

	// 调用提现处理服务（内部使用Nonce管理器）
	q.withdrawProcessor.ProcessWithdrawal(&withdrawal)

	// 重新加载提现记录，检查处理结果
	if err := database.DB.Where("id = ?", task.RecordID).First(&withdrawal).Error; err != nil {
		q.logTask(task.ID, "error", "reload_failed",
			fmt.Sprintf("重新加载提现记录失败: %v", err), "")
		return fmt.Errorf("failed to reload withdrawal record: %w", err)
	}

	if withdrawal.Status == "completed" {
		q.logTask(task.ID, "info", "withdraw_processing_completed",
			"提现处理成功",
			fmt.Sprintf("TxHash: %s, Amount: %s %s",
				withdrawal.TxHash, withdrawal.Amount.String(), withdrawal.Asset))
		return nil
	} else if withdrawal.Status == "failed" {
		q.logTask(task.ID, "error", "withdraw_processing_failed",
			"提现处理失败",
			fmt.Sprintf("原因: 请查看提现记录"))
		return fmt.Errorf("withdrawal processing failed")
	}

	return nil
}

// AddTask 添加任务到队列
func (q *TaskQueue) AddTask(taskType TaskType, symbol string, startTime, endTime *time.Time) (*Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 检查是否已有相同类型和交易对的任务在运行或等待
	for _, t := range q.tasks {
		if t.Type == taskType && t.Symbol == symbol && (t.Status == "pending" || t.Status == "running") {
			return nil, &TaskError{Message: "Same task for this symbol is already running or pending"}
		}
	}

	// 创建新任务
	task := &Task{
		ID:        generateTaskID(),
		Type:      taskType,
		Symbol:    symbol,
		StartTime: startTime,
		EndTime:   endTime,
		Status:    "pending",
		Message:   "等待执行",
		CreatedAt: time.Now(),
	}

	// 同时保存到内存和数据库（线程安全：在锁内完成）
	q.tasks[task.ID] = task

	// 保存到数据库
	dbTask := q.taskToModel(task)
	if err := database.DB.Create(&dbTask).Error; err != nil {
		log.Printf("❌ 保存任务到数据库失败: %v", err)
		// 删除内存中的任务
		delete(q.tasks, task.ID)
		return nil, fmt.Errorf("failed to save task to database: %w", err)
	}

	// 将任务添加到队列
	q.queue <- task

	timeRangeStr := ""
	if startTime != nil && endTime != nil {
		timeRangeStr = fmt.Sprintf(" [%s ~ %s]", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
	}
	log.Printf("📝 任务已添加到队列: %s [%s]%s (ID: %s)", taskType, symbol, timeRangeStr, task.ID)

	// 记录任务创建日志
	q.logTask(task.ID, "info", "task_created",
		fmt.Sprintf("任务已创建: %s", taskType),
		fmt.Sprintf("交易对: %s, 时间范围: %s", symbol, timeRangeStr))

	return task, nil
}

// AddDepositTask 添加充值验证任务
func (q *TaskQueue) AddDepositTask(recordID string) (*Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 创建新任务
	task := &Task{
		ID:         generateTaskID(),
		Type:       TaskVerifyDeposit,
		RecordID:   recordID,
		RecordType: "deposit",
		Status:     "pending",
		Message:    "等待验证充值",
		CreatedAt:  time.Now(),
	}

	// 同时保存到内存和数据库（线程安全：在锁内完成）
	q.tasks[task.ID] = task

	// 保存到数据库
	dbTask := q.taskToModel(task)
	if err := database.DB.Create(&dbTask).Error; err != nil {
		log.Printf("❌ 保存充值任务到数据库失败: %v", err)
		delete(q.tasks, task.ID)
		return nil, fmt.Errorf("failed to save deposit task to database: %w", err)
	}

	// 将任务添加到队列
	q.queue <- task

	log.Printf("📝 充值验证任务已添加到队列: RecordID=%s (TaskID: %s)", recordID, task.ID)
	q.logTask(task.ID, "info", "task_created",
		"充值验证任务已创建",
		fmt.Sprintf("充值记录ID: %s", recordID))

	return task, nil
}

// AddWithdrawTask 添加提现处理任务
func (q *TaskQueue) AddWithdrawTask(recordID string) (*Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 创建新任务
	task := &Task{
		ID:         generateTaskID(),
		Type:       TaskProcessWithdraw,
		RecordID:   recordID,
		RecordType: "withdraw",
		Status:     "pending",
		Message:    "等待处理提现",
		CreatedAt:  time.Now(),
	}

	// 同时保存到内存和数据库（线程安全：在锁内完成）
	q.tasks[task.ID] = task

	// 保存到数据库
	dbTask := q.taskToModel(task)
	if err := database.DB.Create(&dbTask).Error; err != nil {
		log.Printf("❌ 保存提现任务到数据库失败: %v", err)
		delete(q.tasks, task.ID)
		return nil, fmt.Errorf("failed to save withdraw task to database: %w", err)
	}

	// 将任务添加到队列
	q.queue <- task

	log.Printf("📝 提现处理任务已添加到队列: RecordID=%s (TaskID: %s)", recordID, task.ID)
	q.logTask(task.ID, "info", "task_created",
		"提现处理任务已创建",
		fmt.Sprintf("提现记录ID: %s", recordID))

	return task, nil
}

// logTask 记录任务日志到数据库
func (q *TaskQueue) logTask(taskID, level, stage, message, details string) {
	taskLog := models.TaskLog{
		TaskID:  taskID,
		Level:   level,
		Stage:   stage,
		Message: message,
		Details: details,
	}

	if err := database.DB.Create(&taskLog).Error; err != nil {
		log.Printf("❌ 保存任务日志失败: %v", err)
	}
}

// GetTask 获取任务信息
func (q *TaskQueue) GetTask(id string) (*Task, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	task, exists := q.tasks[id]
	return task, exists
}

// GetAllTasks 获取所有任务（从数据库）
func (q *TaskQueue) GetAllTasks() []*Task {
	var dbTasks []models.Task
	// 从数据库读取，限制最近1000条
	database.DB.Order("created_at DESC").Limit(1000).Find(&dbTasks)

	tasks := make([]*Task, 0, len(dbTasks))
	for _, dbTask := range dbTasks {
		task := &Task{
			ID:         dbTask.ID,
			Type:       TaskType(dbTask.Type),
			Status:     dbTask.Status,
			Symbol:     dbTask.Symbol,
			RecordID:   dbTask.RecordID,
			RecordType: dbTask.RecordType,
			StartTime:  dbTask.StartTime,
			EndTime:    dbTask.EndTime,
			Message:    dbTask.Message,
			CreatedAt:  dbTask.CreatedAt,
			StartedAt:  dbTask.StartedAt,
			EndedAt:    dbTask.EndedAt,
			Error:      dbTask.Error,
		}
		tasks = append(tasks, task)
	}

	return tasks
}

// GetRunningTask 获取正在运行的任务
func (q *TaskQueue) GetRunningTask() *Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	for _, task := range q.tasks {
		if task.Status == "running" {
			return task
		}
	}
	return nil
}

// RetryTask 重试失败的任务（线程安全）
func (q *TaskQueue) RetryTask(taskID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 从数据库获取任务
	var dbTask models.Task
	if err := database.DB.Where("id = ?", taskID).First(&dbTask).Error; err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// 只能重试失败的任务
	if dbTask.Status != "failed" {
		return fmt.Errorf("only failed tasks can be retried, current status: %s", dbTask.Status)
	}

	// 转换为内存任务
	task := &Task{
		ID:         dbTask.ID,
		Type:       TaskType(dbTask.Type),
		Status:     "pending",
		Symbol:     dbTask.Symbol,
		RecordID:   dbTask.RecordID,
		RecordType: dbTask.RecordType,
		StartTime:  dbTask.StartTime,
		EndTime:    dbTask.EndTime,
		Message:    "等待重新执行",
		CreatedAt:  dbTask.CreatedAt,
	}

	// 更新内存和数据库
	q.tasks[task.ID] = task

	// 更新数据库中的任务状态
	if err := database.DB.Model(&dbTask).Updates(map[string]interface{}{
		"status":     "pending",
		"message":    "等待重新执行",
		"error":      "",
		"started_at": nil,
		"ended_at":   nil,
	}).Error; err != nil {
		delete(q.tasks, task.ID)
		return fmt.Errorf("failed to update task in database: %w", err)
	}

	// 记录重试日志
	q.logTask(task.ID, "info", "task_retry", "任务已重新加入队列", "")

	// 重新加入队列
	q.queue <- task

	log.Printf("🔄 任务已重试: ID=%s, Type=%s", task.ID, task.Type)
	return nil
}

// updateTask 更新任务（内部使用，需要加锁并同步到数据库）
func (q *TaskQueue) updateTask(task *Task) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 更新内存
	q.tasks[task.ID] = task

	// 同步到数据库（线程安全：在锁内完成）
	dbTask := q.taskToModel(task)
	if err := database.DB.Save(&dbTask).Error; err != nil {
		log.Printf("❌ 更新任务到数据库失败: %v", err)
	}
}

// taskToModel 将任务转换为数据库模型
func (q *TaskQueue) taskToModel(task *Task) models.Task {
	return models.Task{
		ID:         task.ID,
		Type:       string(task.Type),
		Status:     task.Status,
		Symbol:     task.Symbol,
		RecordID:   task.RecordID,
		RecordType: task.RecordType,
		StartTime:  task.StartTime,
		EndTime:    task.EndTime,
		Message:    task.Message,
		CreatedAt:  task.CreatedAt,
		StartedAt:  task.StartedAt,
		EndedAt:    task.EndedAt,
		Error:      task.Error,
	}
}

// generateTaskID 生成任务ID
func generateTaskID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(time.Nanosecond)
	}
	return string(result)
}

// TaskError 任务错误
type TaskError struct {
	Message string
}

func (e *TaskError) Error() string {
	return e.Message
}
