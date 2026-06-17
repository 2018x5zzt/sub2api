package repository

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const (
	openAIImageJobKeyPrefix  = "openai:image_job:"
	openAIImageJobActiveZSet = "openai:image_jobs:active"
)

type openAIImageJobStore struct {
	rdb *redis.Client
}

func NewOpenAIImageJobStore(rdb *redis.Client) service.OpenAIImageJobStore {
	return &openAIImageJobStore{rdb: rdb}
}

func (s *openAIImageJobStore) Create(ctx context.Context, record *service.OpenAIImageJobRecord, ttl time.Duration) error {
	if s == nil || s.rdb == nil || record == nil || strings.TrimSpace(record.ID) == "" {
		return nil
	}
	rec := cloneOpenAIImageJobRecord(record)
	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	pipe := s.rdb.TxPipeline()
	pipe.Set(ctx, openAIImageJobKey(rec.ID), raw, ttl)
	if isOpenAIImageJobActiveStatus(rec.Status) {
		pipe.ZAdd(ctx, openAIImageJobActiveZSet, redis.Z{Score: openAIImageJobScore(rec.UpdatedAt), Member: rec.ID})
	} else {
		pipe.ZRem(ctx, openAIImageJobActiveZSet, rec.ID)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (s *openAIImageJobStore) Get(ctx context.Context, id string, owner service.OpenAIImageJobOwner) (*service.OpenAIImageJobRecord, bool, error) {
	if s == nil || s.rdb == nil || strings.TrimSpace(id) == "" {
		return nil, false, nil
	}
	rec, ok, err := s.getByID(ctx, id)
	if err != nil || !ok {
		return nil, ok, err
	}
	if rec.Owner.UserID != owner.UserID || rec.Owner.APIKeyID != owner.APIKeyID {
		return nil, false, nil
	}
	return rec, true, nil
}

func (s *openAIImageJobStore) SetRunning(ctx context.Context, id string, updatedAt time.Time, ttl time.Duration) (bool, error) {
	if s == nil || s.rdb == nil || strings.TrimSpace(id) == "" {
		return false, nil
	}
	return s.updateActiveJobCAS(ctx, id, ttl, func(rec *service.OpenAIImageJobRecord) bool {
		rec.Status = service.OpenAIImageJobStatusRunning
		rec.UpdatedAt = updatedAt
		return true
	})
}

func (s *openAIImageJobStore) Complete(
	ctx context.Context,
	id string,
	status service.OpenAIImageJobStatus,
	statusCode int,
	headers map[string][]string,
	body []byte,
	completedAt time.Time,
	ttl time.Duration,
) error {
	if s == nil || s.rdb == nil || strings.TrimSpace(id) == "" {
		return nil
	}
	_, err := s.updateActiveJobCAS(ctx, id, ttl, func(rec *service.OpenAIImageJobRecord) bool {
		rec.Status = status
		rec.StatusCode = statusCode
		rec.Headers = cloneOpenAIImageJobHeaders(headers)
		rec.Body = append([]byte(nil), body...)
		rec.UpdatedAt = completedAt
		rec.CompletedAt = completedAt
		return true
	})
	return err
}

func (s *openAIImageJobStore) MarkStaleTimeouts(ctx context.Context, now time.Time, timeout time.Duration, limit int, ttl time.Duration) (int64, error) {
	if s == nil || s.rdb == nil || timeout <= 0 {
		return 0, nil
	}
	if limit <= 0 {
		limit = 100
	}
	cutoff := now.Add(-timeout)
	ids, err := s.rdb.ZRangeByScore(ctx, openAIImageJobActiveZSet, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    floatString(openAIImageJobScore(cutoff)),
		Offset: 0,
		Count:  int64(limit),
	}).Result()
	if err != nil {
		return 0, err
	}
	var changed int64
	for _, id := range ids {
		rec, ok, getErr := s.getByID(ctx, id)
		if getErr != nil {
			return changed, getErr
		}
		if !ok {
			if remErr := s.rdb.ZRem(ctx, openAIImageJobActiveZSet, id).Err(); remErr != nil {
				return changed, remErr
			}
			continue
		}
		if !isOpenAIImageJobActiveStatus(rec.Status) {
			if remErr := s.rdb.ZRem(ctx, openAIImageJobActiveZSet, id).Err(); remErr != nil {
				return changed, remErr
			}
			continue
		}
		if rec.UpdatedAt.After(cutoff) {
			continue
		}
		updated, err := s.updateActiveJobCAS(ctx, id, ttl, func(latest *service.OpenAIImageJobRecord) bool {
			if latest.UpdatedAt.After(cutoff) {
				return false
			}
			latest.Status = service.OpenAIImageJobStatusTimeout
			latest.StatusCode = http.StatusGatewayTimeout
			latest.Headers = service.OpenAIImageJobTimeoutHeaders()
			latest.Body = openAIImageJobRedisTimeoutBody()
			latest.UpdatedAt = now
			latest.CompletedAt = now
			return true
		})
		if err != nil {
			return changed, err
		}
		if updated {
			changed++
		}
	}
	return changed, nil
}

func (s *openAIImageJobStore) updateActiveJobCAS(
	ctx context.Context,
	id string,
	ttl time.Duration,
	mutate func(*service.OpenAIImageJobRecord) bool,
) (bool, error) {
	key := openAIImageJobKey(id)
	for attempt := 0; attempt < 3; attempt++ {
		applied := false
		err := s.rdb.Watch(ctx, func(tx *redis.Tx) error {
			raw, err := tx.Get(ctx, key).Bytes()
			if err == redis.Nil {
				return tx.ZRem(ctx, openAIImageJobActiveZSet, id).Err()
			}
			if err != nil {
				return err
			}
			var rec service.OpenAIImageJobRecord
			if err := json.Unmarshal(raw, &rec); err != nil {
				return err
			}
			if !isOpenAIImageJobActiveStatus(rec.Status) || mutate == nil || !mutate(&rec) {
				return nil
			}
			applied = true
			next, err := json.Marshal(cloneOpenAIImageJobRecord(&rec))
			if err != nil {
				return err
			}
			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Set(ctx, key, next, ttl)
				if isOpenAIImageJobActiveStatus(rec.Status) {
					pipe.ZAdd(ctx, openAIImageJobActiveZSet, redis.Z{Score: openAIImageJobScore(rec.UpdatedAt), Member: rec.ID})
				} else {
					pipe.ZRem(ctx, openAIImageJobActiveZSet, rec.ID)
				}
				return nil
			})
			return err
		}, key)
		if err == redis.TxFailedErr {
			continue
		}
		return applied && err == nil, err
	}
	return false, redis.TxFailedErr
}

func (s *openAIImageJobStore) getByID(ctx context.Context, id string) (*service.OpenAIImageJobRecord, bool, error) {
	raw, err := s.rdb.Get(ctx, openAIImageJobKey(id)).Bytes()
	if err == redis.Nil {
		_ = s.rdb.ZRem(ctx, openAIImageJobActiveZSet, id).Err()
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var rec service.OpenAIImageJobRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return nil, false, err
	}
	return cloneOpenAIImageJobRecord(&rec), true, nil
}

func openAIImageJobKey(id string) string {
	return openAIImageJobKeyPrefix + strings.TrimSpace(id)
}

func openAIImageJobScore(t time.Time) float64 {
	return float64(t.UnixMilli())
}

func floatString(v float64) string {
	return strconv.FormatFloat(v, 'f', 0, 64)
}

func isOpenAIImageJobActiveStatus(status service.OpenAIImageJobStatus) bool {
	return status == service.OpenAIImageJobStatusQueued || status == service.OpenAIImageJobStatusRunning
}

func cloneOpenAIImageJobRecord(in *service.OpenAIImageJobRecord) *service.OpenAIImageJobRecord {
	if in == nil {
		return nil
	}
	out := *in
	out.Headers = cloneOpenAIImageJobHeaders(in.Headers)
	out.Body = append([]byte(nil), in.Body...)
	return &out
}

func cloneOpenAIImageJobHeaders(in map[string][]string) map[string][]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string][]string, len(in))
	for key, values := range in {
		out[key] = append([]string(nil), values...)
	}
	return out
}

func openAIImageJobRedisTimeoutBody() []byte {
	body, err := json.Marshal(map[string]any{
		"error": map[string]any{
			"type":    "api_error",
			"message": "Image job timed out",
		},
	})
	if err != nil {
		return []byte(`{"error":{"type":"api_error","message":"Image job timed out"}}`)
	}
	return body
}
