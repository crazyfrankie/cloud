package ratelimit

import (
	"context"
	"io"
	"time"

	"golang.org/x/time/rate"

	"github.com/crazyfrankie/cloud/pkg/conf"
	"github.com/crazyfrankie/cloud/pkg/consts"
)

// RangeInfo HTTP Range 请求信息
type RangeInfo struct {
	Start  int64
	End    int64
	Length int64
	Total  int64
}

// RateLimitedWriter 速率限制的 Writer 包装器
type RateLimitedWriter struct {
	writer          io.Writer
	userLimiter     *rate.Limiter
	globalLimiter   *rate.Limiter
	ctx             context.Context
	fileSize        int64
	downloadedBytes int64
	startTime       time.Time
	config          RateLimitConfig
}

type RateLimitConfig struct {
	GlobalBytesPerSec     int64                         // 全局每秒下载字节数限制 (默认: 100MB/s)
	DynamicRateCalculator func(requestSize int64) int64 // 动态速率计算函数
	VipLevel              consts.VIPType                // VIP等级 (0=普通用户, 1=VIP, 2=SVIP)
	EnableDynamicRate     bool                          // 是否启用动态调整
}

// NewRateLimitConfig 创建默认速率限制配置
func NewRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		GlobalBytesPerSec:     conf.GetConf().System.GlobalBytesLimitPerSec,
		DynamicRateCalculator: calculateDynamicRate,
	}
}

// NewVipRateLimitConfig 创建VIP用户的速率限制配置
func NewVipRateLimitConfig(vipLevel consts.VIPType) RateLimitConfig {
	config := NewRateLimitConfig()
	config.VipLevel = vipLevel
	config.EnableDynamicRate = true
	return config
}

// calculateDynamicRate 根据请求大小计算固定速率 - 专为大文件设计
func calculateDynamicRate(requestSize int64) int64 {
	const MB = 1024 * 1024

	// 确保稳定的下载体验，避免速度波动
	switch {
	case requestSize <= 500*MB:
		// 500MB以下：5MB/s
		return 5 * MB
	case requestSize <= 2*1024*MB: // 2GB
		// 500MB-2GB：4MB/s
		return 4 * MB
	default:
		// 2GB以上：3MB/s
		return 3 * MB
	}
}

// NewRateLimitedWriter 创建速率限制的Writer
func NewRateLimitedWriter(ctx context.Context, writer io.Writer, config RateLimitConfig, fileSize int64, rangeInfo *RangeInfo) *RateLimitedWriter {
	// 计算实际传输大小（基于Range头）
	var actualTransferSize = fileSize
	if rangeInfo != nil {
		actualTransferSize = rangeInfo.End - rangeInfo.Start + 1
	}
	// 为大文件下载设置稳定的初始速率
	// 直接使用目标速率，避免速度波动
	var initialRate int64
	const (
		KB = 1024
		MB = 1024 * KB
	)

	// 根据文件大小直接设置最终速率，确保稳定性
	switch {
	case actualTransferSize <= 500*MB:
		initialRate = 5 * MB // 500MB以下：5MB/s
	case actualTransferSize <= 2*1024*MB: // 2GB
		initialRate = 4 * MB // 500MB-2GB：4MB/s
	default:
		initialRate = 3 * MB // 2GB以上：3MB/s
	}

	// 计算合适的burst大小 - 平衡稳定性和兼容性
	// 设置适中的burst，避免初始速度过高，同时确保能处理常见缓冲区
	userBurstSize := int(initialRate / 2) // 0.5倍速率作为burst
	const minBurstSize = 4 * MB           // 最小4MB burst，兼容大多数缓冲区
	const maxBurstSize = 8 * MB           // 最大8MB burst，限制突发但兼容大缓冲区

	if userBurstSize < minBurstSize {
		userBurstSize = minBurstSize
	}
	if userBurstSize > maxBurstSize {
		userBurstSize = maxBurstSize
	}

	globalBurstSize := int(config.GlobalBytesPerSec / 10) // 全局burst也控制在更小范围
	if globalBurstSize < minBurstSize {
		globalBurstSize = minBurstSize
	}

	// 创建速率限制器
	userLimiter := rate.NewLimiter(rate.Limit(initialRate), userBurstSize)
	globalLimiter := rate.NewLimiter(rate.Limit(config.GlobalBytesPerSec), globalBurstSize)

	return &RateLimitedWriter{
		writer:          writer,
		userLimiter:     userLimiter,
		globalLimiter:   globalLimiter,
		ctx:             ctx,
		fileSize:        actualTransferSize, // 使用实际传输大小而非文件大小
		downloadedBytes: 0,
		startTime:       time.Now(),
		config:          config,
	}
}

// Write 支持动态速率调整和大数据块分批处理
func (w *RateLimitedWriter) Write(data []byte) (int, error) {
	// 检查是否需要动态调整速率
	if w.config.EnableDynamicRate {
		w.adjustRateIfNeeded()
	}

	dataLen := len(data)
	burstSize := w.userLimiter.Burst()

	// 如果数据块大于 burst 大小，分批处理
	if dataLen > burstSize {
		totalWritten := 0
		for totalWritten < dataLen {
			// 计算本次写入的大小
			chunkSize := burstSize
			if totalWritten+chunkSize > dataLen {
				chunkSize = dataLen - totalWritten
			}

			// 等待用户级别的限制
			if w.userLimiter != nil {
				if err := w.userLimiter.WaitN(w.ctx, chunkSize); err != nil {
					return totalWritten, err
				}
			}

			// 等待全局级别的限制
			if w.globalLimiter != nil {
				if err := w.globalLimiter.WaitN(w.ctx, chunkSize); err != nil {
					return totalWritten, err
				}
			}

			// 写入这一块数据
			n, err := w.writer.Write(data[totalWritten : totalWritten+chunkSize])
			totalWritten += n

			if err != nil {
				// 更新已下载字节数（即使出错也要更新已成功写入的部分）
				w.downloadedBytes += int64(totalWritten)
				return totalWritten, err
			}
		}

		// 更新已下载字节数
		w.downloadedBytes += int64(totalWritten)

		return totalWritten, nil
	}

	// 小数据块，直接处理
	tokens := dataLen

	// 等待用户级别的限制
	if w.userLimiter != nil {
		if err := w.userLimiter.WaitN(w.ctx, tokens); err != nil {
			return 0, err
		}
	}

	// 等待全局级别的限制
	if w.globalLimiter != nil {
		if err := w.globalLimiter.WaitN(w.ctx, tokens); err != nil {
			return 0, err
		}
	}

	// 写入数据
	n, err := w.writer.Write(data)

	// 更新已下载字节数
	if err == nil {
		w.downloadedBytes += int64(n)
	}

	return n, err
}

// adjustRateIfNeeded 根据VIP等级和下载进度动态调整速率
func (w *RateLimitedWriter) adjustRateIfNeeded() {
	const (
		KB = 1024
		MB = 1024 * KB
	)

	// 计算已传输时间和当前速度
	elapsed := time.Since(w.startTime).Seconds()
	if elapsed < 1.0 { // 前1秒不调整，让系统稳定
		return
	}

	// 根据VIP等级确定基础速率倍数
	var speedMultiplier float64
	var maxSpeed int64
	switch w.config.VipLevel {
	case consts.VIP: // VIP用户
		speedMultiplier = 2.0
		maxSpeed = 10 * MB
	case consts.SVIP: // SVIP 用户
		speedMultiplier = 3.0
		maxSpeed = 15 * MB
	default:
		speedMultiplier = 1.0
		maxSpeed = 5 * MB
	}

	// 根据文件大小计算基础速率
	var baseRate int64
	switch {
	case w.fileSize <= 500*MB:
		baseRate = int64(float64(5*MB) * speedMultiplier)
	case w.fileSize <= 2*1024*MB: // 2GB
		baseRate = int64(float64(4*MB) * speedMultiplier)
	default:
		baseRate = int64(float64(3*MB) * speedMultiplier)
	}

	// 限制最大速度
	if baseRate > maxSpeed {
		baseRate = maxSpeed
	}

	// 根据下载进度进行微调
	progress := float64(w.downloadedBytes) / float64(w.fileSize)
	var adjustedRate int64

	if progress < 0.1 { // 前10%，稍微保守
		adjustedRate = int64(float64(baseRate) * 0.9)
	} else if progress > 0.8 { // 后20%，可以提速完成
		adjustedRate = int64(float64(baseRate) * 1.1)
	} else { // 中间阶段，保持基础速率
		adjustedRate = baseRate
	}

	// 获取当前限制速率
	currentLimit := int64(w.userLimiter.Limit())

	// 只有当速率变化超过10%时才调整，避免频繁变动
	rateDiff := float64(adjustedRate-currentLimit) / float64(currentLimit)
	if rateDiff > 0.1 || rateDiff < -0.1 {
		w.userLimiter.SetLimit(rate.Limit(adjustedRate))
	}
}
