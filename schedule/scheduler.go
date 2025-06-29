package cron

import (
	"bossfi-backend/app/services"
	"bossfi-backend/config"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type Scheduler struct {
	cron              *cron.Cron
	blockchainService *services.BlockchainService
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		cron: cron.New(
			cron.WithChain(
				cron.SkipIfStillRunning(cron.DefaultLogger),
				cron.Recover(cron.DefaultLogger),
			),
		),
		blockchainService: services.NewBlockchainService(),
	}
}

// Start 启动定时任务调度器
func (s *Scheduler) Start() error {
	if !config.AppConfig.Cron.Enabled {
		logrus.Info("Cron scheduler is disabled")
		return nil
	}

	logrus.Info("Starting cron scheduler...")

	// 添加区块链数据同步任务
	if config.AppConfig.Cron.BlockchainSyncInterval != "" {
		_, err := s.cron.AddFunc(config.AppConfig.Cron.BlockchainSyncInterval, func() {
			logrus.Info("Running blockchain data synchronization job...")
			if err := s.blockchainService.SyncBlockchainData(); err != nil {
				logrus.Errorf("Blockchain sync job failed: %v", err)
			}
		})
		if err != nil {
			logrus.Errorf("Failed to add blockchain sync job: %v", err)
			return err
		}
		logrus.Infof("Added blockchain sync job with interval: %s", config.AppConfig.Cron.BlockchainSyncInterval)
	}

	// 可以添加更多定时任务
	// 例如：清理过期的 nonce、统计分析、数据备份等

	// 添加清理过期 Redis 数据的任务（每小时执行一次）
	_, err := s.cron.AddFunc("0 * * * *", func() {
		logrus.Info("Running Redis cleanup job...")
		s.cleanupExpiredData()
	})
	if err != nil {
		logrus.Errorf("Failed to add Redis cleanup job: %v", err)
		return err
	}
	logrus.Info("Added Redis cleanup job (hourly)")

	// 启动定时任务
	s.cron.Start()
	logrus.Info("Cron scheduler started successfully")

	return nil
}

// Stop 停止定时任务调度器
func (s *Scheduler) Stop() {
	if s.cron != nil {
		ctx := s.cron.Stop()
		<-ctx.Done()
		logrus.Info("Cron scheduler stopped")
	}

	if s.blockchainService != nil {
		s.blockchainService.Close()
	}
}

// cleanupExpiredData 清理过期数据
func (s *Scheduler) cleanupExpiredData() {
	// 这里可以实现清理逻辑
	// 例如：删除过期的 nonce、清理过期的会话等
	logrus.Debug("Cleanup job completed")
}

// AddCustomJob 添加自定义定时任务
func (s *Scheduler) AddCustomJob(spec string, cmd func()) (cron.EntryID, error) {
	return s.cron.AddFunc(spec, cmd)
}

// RemoveJob 移除定时任务
func (s *Scheduler) RemoveJob(id cron.EntryID) {
	s.cron.Remove(id)
}
