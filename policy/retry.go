package policy

import (
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"time"
)

// RetryableFunc 定义支持任意参数的函数类型
type RetryableFunc func(args ...any) (any, error)

// Policy 定义限流和重试的策略配置
type Policy struct {
	RateLimit       rate.Limit       // 每秒允许的请求数
	Burst           int              // 突发容量
	MaxRetries      int              // 最大重试次数
	InitialInterval time.Duration    // 初始重试间隔
	MaxInterval     time.Duration    // 最大重试间隔
	Multiplier      float64          // 间隔时间乘数
	IsRetriable     func(error) bool // 判断错误是否可重试
}

// DefaultPolicy 默认策略配置
var DefaultPolicy = Policy{
	RateLimit:       rate.Every(time.Second),                    // 每秒1次请求
	Burst:           2,                                          // 突发容量1
	MaxRetries:      5,                                          // 最多重试3次
	InitialInterval: 500 * time.Millisecond,                     // 初始间隔500ms
	MaxInterval:     2 * time.Second,                            // 最大间隔2s
	Multiplier:      2.0,                                        // 指数退避乘数
	IsRetriable:     func(err error) bool { return err != nil }, // 默认所有错误都重试
}

// LimiterRetry 限流重试执行器
type LimiterRetry struct {
	policy  Policy
	limiter *rate.Limiter
}

// NewLR 创建限流重试执行器
func NewLR(policy Policy) *LimiterRetry {
	// 设置默认值
	if policy.RateLimit == 0 {
		policy.RateLimit = DefaultPolicy.RateLimit
	}
	if policy.Burst == 0 {
		policy.Burst = DefaultPolicy.Burst
	}
	if policy.MaxRetries == 0 {
		policy.MaxRetries = DefaultPolicy.MaxRetries
	}
	if policy.InitialInterval == 0 {
		policy.InitialInterval = DefaultPolicy.InitialInterval
	}
	if policy.MaxInterval == 0 {
		policy.MaxInterval = DefaultPolicy.MaxInterval
	}
	if policy.Multiplier == 0 {
		policy.Multiplier = DefaultPolicy.Multiplier
	}
	if policy.IsRetriable == nil {
		policy.IsRetriable = DefaultPolicy.IsRetriable
	}

	return &LimiterRetry{
		policy:  policy,
		limiter: rate.NewLimiter(policy.RateLimit, policy.Burst),
	}
}

// Execute 执行带限流和重试的函数
func (lr *LimiterRetry) Execute(ctx context.Context, fn RetryableFunc, args ...any) (any, error) {
	var lastResult interface{}
	var lastError error

	// 计算退避时间（修复幂运算的类型问题）
	calculateBackoff := func(attempt int) time.Duration {
		if attempt <= 0 {
			return 0
		}
		// 使用整数计算幂次，避免 float64 的 % 运算错误
		backoff := float64(lr.policy.InitialInterval)
		for i := 1; i < attempt; i++ {
			backoff *= lr.policy.Multiplier
			if backoff > float64(lr.policy.MaxInterval) {
				return lr.policy.MaxInterval
			}
		}
		return time.Duration(backoff)
	}

	totalAttempts := lr.policy.MaxRetries + 1 // 总尝试次数（含首次）

	for attempt := 1; attempt <= totalAttempts; attempt++ {
		// 等待限流许可
		if err := lr.limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("限流等待失败: %w", err)
		}

		// 执行目标函数
		lastResult, lastError = fn(args...)

		// 成功或遇到不可重试的错误，直接返回
		if lastError == nil || !lr.policy.IsRetriable(lastError) {
			return lastResult, lastError
		}

		// 未达最大尝试次数，等待退避后重试
		if attempt < totalAttempts {
			backoff := calculateBackoff(attempt)
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("上下文已结束: %w", ctx.Err())
			case <-time.After(backoff):
				continue // 继续重试
			}
		}
	}

	// 达到最大重试次数
	return lastResult, fmt.Errorf("达到最大重试次数（%d次），最后错误: %w", lr.policy.MaxRetries, lastError)
}
