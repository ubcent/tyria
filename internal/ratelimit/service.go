package ratelimit

// Service provides a unified rate limiting service that can use either Redis or in-memory storage
type Service struct {
	limiter PolicyLimiter
}

// ServiceConfig holds configuration for the rate limiting service
type ServiceConfig struct {
	UseRedis      bool
	RedisConfig   RedisConfig
	InMemoryConfig Config
}

// NewService creates a new rate limiting service
func NewService(config ServiceConfig) *Service {
	var limiter PolicyLimiter
	
	if config.UseRedis {
		limiter = NewRedisLimiter(config.RedisConfig)
	} else {
		limiter = NewLimiter(config.InMemoryConfig)
	}
	
	return &Service{
		limiter: limiter,
	}
}

// Allow checks if a request is allowed for the given key
func (s *Service) Allow(key string) bool {
	return s.limiter.Allow(key)
}

// AllowN checks if n requests are allowed for the given key
func (s *Service) AllowN(key string, n int) bool {
	return s.limiter.AllowN(key, n)
}

// AllowWithPolicy checks if a request is allowed with the given policy
func (s *Service) AllowWithPolicy(key string, requestsPerMinute, burst int) (allowed bool, retryAfter int) {
	return s.limiter.AllowWithPolicy(key, requestsPerMinute, burst)
}

// Stats returns statistics about the rate limiter
func (s *Service) Stats() Stats {
	return s.limiter.Stats()
}

// Reset removes all buckets
func (s *Service) Reset() {
	s.limiter.Reset()
}

// Close closes the underlying limiter (useful for Redis connections)
func (s *Service) Close() error {
	if closer, ok := s.limiter.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}