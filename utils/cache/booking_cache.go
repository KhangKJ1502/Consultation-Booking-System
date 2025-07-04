// utils/cache/booking_cache.go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisBookingCache struct {
	redis *RedisCache
}

type BookingCacheData struct {
	BookingID        string    `json:"booking_id"`
	UserID           string    `json:"user_id"`
	ExpertProfileID  string    `json:"expert_profile_id"`
	BookingDatetime  time.Time `json:"booking_datetime"`
	DurationMinutes  int       `json:"duration_minutes"`
	BookingStatus    string    `json:"booking_status"`
	ConsultationType string    `json:"consultation_type"`
}

type BookingCache interface {
	// Kiểm tra expert có available không trong khoảng thời gian
	IsExpertAvailable(ctx context.Context, expertID string, startTime, endTime time.Time) (bool, error)
	// Kiểm tra user có booking trùng lặp không
	HasConflictingBooking(ctx context.Context, userID string, startTime, endTime time.Time) (bool, error)
	// Lưu booking vào cache
	CacheBooking(ctx context.Context, booking *BookingCacheData) error
	// Lấy booking từ cache
	GetBooking(ctx context.Context, bookingID string) (*BookingCacheData, error)
	// Xóa booking khỏi cache
	DeleteBooking(ctx context.Context, bookingID string) error
	// Lấy tất cả booking của expert trong ngày
	GetExpertBookingsForDay(ctx context.Context, expertID string, date time.Time) ([]*BookingCacheData, error)
	// Lấy tất cả booking của user trong ngày
	GetUserBookingsForDay(ctx context.Context, userID string, date time.Time) ([]*BookingCacheData, error)
}

type redisBookingCache struct {
	redis  *RedisCache
	logger *zap.Logger
}

func NewRedisBookingCache(redis *RedisCache, logger *zap.Logger) BookingCache {
	return &redisBookingCache{
		redis:  redis,
		logger: logger,
	}
}

// Redis key patterns
const (
	bookingKeyPrefix        = "booking:"
	expertScheduleKeyPrefix = "expert_schedule:"
	userScheduleKeyPrefix   = "user_schedule:"
	dailyBookingKeyPrefix   = "daily_bookings:"
)

func (r *redisBookingCache) IsExpertAvailable(ctx context.Context, expertID string, startTime, endTime time.Time) (bool, error) {
	// Tạo key để kiểm tra lịch expert
	// scheduleKey := fmt.Sprintf("%s%s", expertScheduleKeyPrefix, expertID)

	// Kiểm tra các booking hiện tại của expert
	bookings, err := r.GetExpertBookingsForDay(ctx, expertID, startTime)
	if err != nil {
		r.logger.Error("Failed to get expert bookings", zap.Error(err))
		return false, err
	}

	// Kiểm tra xung đột thời gian
	for _, booking := range bookings {
		bookingStart := booking.BookingDatetime
		bookingEnd := bookingStart.Add(time.Duration(booking.DurationMinutes) * time.Minute)

		// Kiểm tra overlap
		if r.isTimeOverlap(startTime, endTime, bookingStart, bookingEnd) {
			return false, nil
		}
	}

	return true, nil
}

func (r *redisBookingCache) HasConflictingBooking(ctx context.Context, userID string, startTime, endTime time.Time) (bool, error) {
	// Lấy tất cả booking của user trong ngày
	bookings, err := r.GetUserBookingsForDay(ctx, userID, startTime)
	if err != nil {
		r.logger.Error("Failed to get user bookings", zap.Error(err))
		return false, err
	}

	// Kiểm tra xung đột thời gian
	for _, booking := range bookings {
		bookingStart := booking.BookingDatetime
		bookingEnd := bookingStart.Add(time.Duration(booking.DurationMinutes) * time.Minute)

		// Kiểm tra overlap
		if r.isTimeOverlap(startTime, endTime, bookingStart, bookingEnd) {
			return true, nil
		}
	}

	return false, nil
}

func (r *redisBookingCache) CacheBooking(ctx context.Context, booking *BookingCacheData) error {
	// Serialize booking data
	bookingJSON, err := json.Marshal(booking)
	if err != nil {
		r.logger.Error("Failed to marshal booking", zap.Error(err))
		return err
	}

	// Cache với multiple keys để query dễ dàng
	pipe := r.redis.Client.Pipeline()

	// 1. Cache booking chính
	bookingKey := fmt.Sprintf("%s%s", bookingKeyPrefix, booking.BookingID)
	pipe.Set(ctx, bookingKey, bookingJSON, 24*time.Hour) // Cache 24h

	// 2. Cache trong danh sách expert schedule
	expertDailyKey := fmt.Sprintf("%s%s:%s", dailyBookingKeyPrefix, booking.ExpertProfileID, booking.BookingDatetime.Format("2006-01-02"))
	pipe.SAdd(ctx, expertDailyKey, booking.BookingID)
	pipe.Expire(ctx, expertDailyKey, 48*time.Hour)

	// 3. Cache trong danh sách user schedule
	userDailyKey := fmt.Sprintf("%s%s:%s", dailyBookingKeyPrefix, booking.UserID, booking.BookingDatetime.Format("2006-01-02"))
	pipe.SAdd(ctx, userDailyKey, booking.BookingID)
	pipe.Expire(ctx, userDailyKey, 48*time.Hour)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to cache booking", zap.Error(err))
		return err
	}

	r.logger.Info("Booking cached successfully", zap.String("bookingID", booking.BookingID))
	return nil
}

func (r *redisBookingCache) GetBooking(ctx context.Context, bookingID string) (*BookingCacheData, error) {
	bookingKey := fmt.Sprintf("%s%s", bookingKeyPrefix, bookingID)

	result, err := r.redis.Client.Get(ctx, bookingKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Booking not found in cache
		}
		r.logger.Error("Failed to get booking from cache", zap.Error(err))
		return nil, err
	}

	var booking BookingCacheData
	if err := json.Unmarshal([]byte(result), &booking); err != nil {
		r.logger.Error("Failed to unmarshal booking", zap.Error(err))
		return nil, err
	}

	return &booking, nil
}

func (r *redisBookingCache) DeleteBooking(ctx context.Context, bookingID string) error {
	// Lấy booking để có thông tin cho việc xóa khỏi daily sets
	booking, err := r.GetBooking(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking == nil {
		return nil // Booking not found
	}

	pipe := r.redis.Client.Pipeline()

	// 1. Xóa booking chính
	bookingKey := fmt.Sprintf("%s%s", bookingKeyPrefix, bookingID)
	pipe.Del(ctx, bookingKey)

	// 2. Xóa khỏi expert daily set
	expertDailyKey := fmt.Sprintf("%s%s:%s", dailyBookingKeyPrefix, booking.ExpertProfileID, booking.BookingDatetime.Format("2006-01-02"))
	pipe.SRem(ctx, expertDailyKey, bookingID)

	// 3. Xóa khỏi user daily set
	userDailyKey := fmt.Sprintf("%s%s:%s", dailyBookingKeyPrefix, booking.UserID, booking.BookingDatetime.Format("2006-01-02"))
	pipe.SRem(ctx, userDailyKey, bookingID)

	_, err = pipe.Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete booking from cache", zap.Error(err))
		return err
	}

	return nil
}

func (r *redisBookingCache) GetExpertBookingsForDay(ctx context.Context, expertID string, date time.Time) ([]*BookingCacheData, error) {
	dailyKey := fmt.Sprintf("%s%s:%s", dailyBookingKeyPrefix, expertID, date.Format("2006-01-02"))

	// Lấy tất cả booking IDs cho ngày
	bookingIDs, err := r.redis.Client.SMembers(ctx, dailyKey).Result()
	if err != nil {
		if err == redis.Nil {
			return []*BookingCacheData{}, nil
		}
		r.logger.Error("Failed to get expert daily bookings", zap.Error(err))
		return nil, err
	}

	// Lấy chi tiết từng booking
	var bookings []*BookingCacheData
	for _, bookingID := range bookingIDs {
		booking, err := r.GetBooking(ctx, bookingID)
		if err != nil {
			r.logger.Error("Failed to get booking detail", zap.String("bookingID", bookingID), zap.Error(err))
			continue
		}
		if booking != nil {
			bookings = append(bookings, booking)
		}
	}

	return bookings, nil
}

func (r *redisBookingCache) GetUserBookingsForDay(ctx context.Context, userID string, date time.Time) ([]*BookingCacheData, error) {
	dailyKey := fmt.Sprintf("%s%s:%s", dailyBookingKeyPrefix, userID, date.Format("2006-01-02"))

	// Lấy tất cả booking IDs cho ngày
	bookingIDs, err := r.redis.Client.SMembers(ctx, dailyKey).Result()
	if err != nil {
		if err == redis.Nil {
			return []*BookingCacheData{}, nil
		}
		r.logger.Error("Failed to get user daily bookings", zap.Error(err))
		return nil, err
	}

	// Lấy chi tiết từng booking
	var bookings []*BookingCacheData
	for _, bookingID := range bookingIDs {
		booking, err := r.GetBooking(ctx, bookingID)
		if err != nil {
			r.logger.Error("Failed to get booking detail", zap.String("bookingID", bookingID), zap.Error(err))
			continue
		}
		if booking != nil {
			bookings = append(bookings, booking)
		}
	}

	return bookings, nil
}

// Helper function để kiểm tra time overlap
func (r *redisBookingCache) isTimeOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && end1.After(start2)
}
