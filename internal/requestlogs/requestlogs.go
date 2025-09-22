package requestlogs

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ubcent/edge.link/internal/models"
)

// Service provides request logging functionality
type Service struct {
	db *sql.DB
}

// NewService creates a new request log service
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// Log creates a new request log entry
func (s *Service) Log(ctx context.Context, log *models.RequestLog) error {
	query := `
		INSERT INTO requests_log (tenant_id, route_id, status_code, latency_ms, cache_status, bytes_in, bytes_out, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	
	err := s.db.QueryRowContext(ctx, query,
		log.TenantID, log.RouteID, log.StatusCode, log.LatencyMs,
		log.CacheStatus, log.BytesIn, log.BytesOut, log.CreatedAt,
	).Scan(&log.ID)
	
	if err != nil {
		return fmt.Errorf("failed to create request log: %w", err)
	}
	
	return nil
}

// GetByTenant retrieves request logs for a specific tenant with pagination
func (s *Service) GetByTenant(ctx context.Context, tenantID int, limit, offset int) ([]*models.RequestLog, error) {
	query := `
		SELECT id, tenant_id, route_id, status_code, latency_ms, cache_status, bytes_in, bytes_out, created_at
		FROM requests_log
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := s.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get request logs for tenant: %w", err)
	}
	defer rows.Close()
	
	var logs []*models.RequestLog
	for rows.Next() {
		log := &models.RequestLog{}
		err := rows.Scan(
			&log.ID, &log.TenantID, &log.RouteID, &log.StatusCode, &log.LatencyMs,
			&log.CacheStatus, &log.BytesIn, &log.BytesOut, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan request log: %w", err)
		}
		logs = append(logs, log)
	}
	
	return logs, nil
}

// GetByRoute retrieves request logs for a specific route with pagination
func (s *Service) GetByRoute(ctx context.Context, routeID int, limit, offset int) ([]*models.RequestLog, error) {
	query := `
		SELECT id, tenant_id, route_id, status_code, latency_ms, cache_status, bytes_in, bytes_out, created_at
		FROM requests_log
		WHERE route_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := s.db.QueryContext(ctx, query, routeID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get request logs for route: %w", err)
	}
	defer rows.Close()
	
	var logs []*models.RequestLog
	for rows.Next() {
		log := &models.RequestLog{}
		err := rows.Scan(
			&log.ID, &log.TenantID, &log.RouteID, &log.StatusCode, &log.LatencyMs,
			&log.CacheStatus, &log.BytesIn, &log.BytesOut, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan request log: %w", err)
		}
		logs = append(logs, log)
	}
	
	return logs, nil
}

// GetStats retrieves aggregated statistics for a tenant
func (s *Service) GetStats(ctx context.Context, tenantID int, since time.Time) (*RequestStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_requests,
			AVG(latency_ms) as avg_latency,
			COUNT(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 END) as success_count,
			COUNT(CASE WHEN cache_status = 'hit' THEN 1 END) as cache_hits,
			SUM(bytes_in) as total_bytes_in,
			SUM(bytes_out) as total_bytes_out
		FROM requests_log
		WHERE tenant_id = $1 AND created_at >= $2
	`
	
	var stats RequestStats
	var avgLatency sql.NullFloat64
	var totalBytesIn, totalBytesOut sql.NullInt64
	
	err := s.db.QueryRowContext(ctx, query, tenantID, since).Scan(
		&stats.TotalRequests, &avgLatency, &stats.SuccessCount,
		&stats.CacheHits, &totalBytesIn, &totalBytesOut,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get request stats: %w", err)
	}
	
	stats.AvgLatency = int(avgLatency.Float64)
	stats.TotalBytesIn = totalBytesIn.Int64
	stats.TotalBytesOut = totalBytesOut.Int64
	
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.TotalRequests) * 100
		stats.CacheHitRate = float64(stats.CacheHits) / float64(stats.TotalRequests) * 100
	}
	
	return &stats, nil
}

// RequestStats represents aggregated request statistics
type RequestStats struct {
	TotalRequests  int64   `json:"total_requests"`
	SuccessCount   int64   `json:"success_count"`
	SuccessRate    float64 `json:"success_rate"`
	CacheHits      int64   `json:"cache_hits"`
	CacheHitRate   float64 `json:"cache_hit_rate"`
	AvgLatency     int     `json:"avg_latency_ms"`
	TotalBytesIn   int64   `json:"total_bytes_in"`
	TotalBytesOut  int64   `json:"total_bytes_out"`
}

// CleanupOldLogs removes request logs older than the specified duration
func (s *Service) CleanupOldLogs(ctx context.Context, olderThan time.Time) (int64, error) {
	result, err := s.db.ExecContext(ctx, 
		"DELETE FROM requests_log WHERE created_at < $1", olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old logs: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	return rowsAffected, nil
}

// generateVerificationToken generates a random verification token
func (s *Service) generateVerificationToken() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}