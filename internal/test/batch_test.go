package batcher

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/shamil/proxy_track_service-1/internal/batcher"
	"github.com/shamil/proxy_track_service-1/internal/config"
	"github.com/shamil/proxy_track_service-1/internal/models"
	"github.com/shamil/proxy_track_service-1/internal/repository"
)

type MockExternalAPIClient struct {
	requests []string
	mu       sync.Mutex
}

func NewMockExternalAPIClient() *MockExternalAPIClient {
	return &MockExternalAPIClient{
		requests: make([]string, 0),
	}
}

func (m *MockExternalAPIClient) TrackPackagesBatch(ctx context.Context, trackCodes []string) (map[string]*models.TrackData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requests = append(m.requests, trackCodes...)
	result := make(map[string]*models.TrackData)
	for _, code := range trackCodes {
		result[code] = &models.TrackData{
			Countries: []string{code + " - In Transit"},
			Events: []models.Event{
				{
					Status: "Package received",
					Date:   time.Now().Format(time.RFC3339),
				},
			},
		}
	}

	return result, nil
}

func (m *MockExternalAPIClient) TrackPackage(ctx context.Context, trackCode string) (*models.TrackData, error) {
	results, err := m.TrackPackagesBatch(ctx, []string{trackCode})
	if err != nil {
		return nil, err
	}
	return results[trackCode], nil
}

func (m *MockExternalAPIClient) Health(ctx context.Context) error {
	return nil
}

func (m *MockExternalAPIClient) GetRequestCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.requests)
}

func (m *MockExternalAPIClient) GetRequests() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.requests...)
}

type MockCacheRepository struct {
	data map[string]*models.TrackData
	mu   sync.Mutex
}

func NewMockCacheRepository() *MockCacheRepository {
	return &MockCacheRepository{
		data: make(map[string]*models.TrackData),
	}
}

func (m *MockCacheRepository) GetTrackData(ctx context.Context, trackCode string) (*models.TrackData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if data, exists := m.data[trackCode]; exists {
		return data, nil
	}
	return nil, repository.ErrTrackDataNotFound
}

func (m *MockCacheRepository) SetTrackData(ctx context.Context, trackCode string, data *models.TrackData, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[trackCode] = data
	return nil
}

func (m *MockCacheRepository) Health(ctx context.Context) error {
	return nil
}

func (m *MockCacheRepository) Get(ctx context.Context, key string) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if data, exists := m.data[key]; exists {
		return data, nil
	}
	return nil, repository.ErrTrackDataNotFound
}

func (m *MockCacheRepository) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if data, ok := value.(*models.TrackData); ok {
		m.data[key] = data
	}
	return nil
}

func (m *MockCacheRepository) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}

func (m *MockCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.data[key]
	return exists, nil
}

func (m *MockCacheRepository) Close() error {
	return nil
}

// TestBatcherBatchSize - тест накопления 50 трек-кодов
func TestBatcherBatchSize(t *testing.T) {
	config := config.BatcherConfig{
		BatchSize:    5,
		BatchTimeout: 10 * time.Second,
		Workers:      1,
	}

	mockClient := NewMockExternalAPIClient()
	mockCache := NewMockCacheRepository()

	batcher := batcher.NewBatcher(config, mockCache, mockClient)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := batcher.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start batcher: %v", err)
	}
	defer batcher.Stop()

	trackCodes := []string{"TEST001", "TEST002", "TEST003", "TEST004", "TEST005"}

	var wg sync.WaitGroup
	responses := make([]models.TrackResponse, len(trackCodes))

	for i, trackCode := range trackCodes {
		wg.Add(1)
		go func(i int, code string) {
			defer wg.Done()
			responseChan := batcher.AddRequest(ctx, code)
			select {
			case response := <-responseChan:
				responses[i] = response
			case <-ctx.Done():
				t.Errorf("Request %d timed out", i)
			}
		}(i, trackCode)
	}
	wg.Wait()

	for i, response := range responses {
		if !response.Status {
			t.Errorf("Request %d failed: %s", i, response.Error)
		}
		if response.Data == nil {
			t.Errorf("Request %d has no data", i)
		}
	}

	requestCount := mockClient.GetRequestCount()
	if requestCount != 5 {
		t.Errorf("Expected 5 requests to external API, got %d", requestCount)
	}

	requests := mockClient.GetRequests()
	if len(requests) != 5 {
		t.Errorf("Expected 5 track codes, got %d", len(requests))
	}

	expectedCodes := map[string]bool{
		"TEST001": true,
		"TEST002": true,
		"TEST003": true,
		"TEST004": true,
		"TEST005": true,
	}

	for _, code := range requests {
		if !expectedCodes[code] {
			t.Errorf("Unexpected track code: %s", code)
		}
	}
}

// TestBatcherTimeout - тест таймаута батча
func TestBatcherTimeout(t *testing.T) {
	config := config.BatcherConfig{
		BatchSize:    10,                     // Большой размер батча
		BatchTimeout: 100 * time.Millisecond, // Короткий таймаут
		Workers:      1,
	}

	mockClient := NewMockExternalAPIClient()
	mockCache := NewMockCacheRepository()

	batcher := batcher.NewBatcher(config, mockCache, mockClient)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := batcher.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start batcher: %v", err)
	}
	defer batcher.Stop()

	trackCodes := []string{"TIMEOUT001", "TIMEOUT002"}

	var wg sync.WaitGroup
	responses := make([]models.TrackResponse, len(trackCodes))

	for i, trackCode := range trackCodes {
		wg.Add(1)
		go func(i int, code string) {
			defer wg.Done()
			responseChan := batcher.AddRequest(ctx, code)
			select {
			case response := <-responseChan:
				responses[i] = response
			case <-ctx.Done():
				t.Errorf("Request %d timed out", i)
			}
		}(i, trackCode)
	}
	wg.Wait()

	for i, response := range responses {
		if !response.Status {
			t.Errorf("Request %d failed: %s", i, response.Error)
		}
		if response.Data == nil {
			t.Errorf("Request %d has no data", i)
		}
	}

	requestCount := mockClient.GetRequestCount()
	if requestCount != 2 { // 2 трек-кода в одном запросе
		t.Errorf("Expected 2 requests to external API, got %d", requestCount)
	}

	requests := mockClient.GetRequests()
	if len(requests) != 2 {
		t.Errorf("Expected 2 track codes, got %d", len(requests))
	}
}

// TestBatcherMultipleBatches - тест нескольких батчей
func TestBatcherMultipleBatches(t *testing.T) {
	config := config.BatcherConfig{
		BatchSize:    3, // Маленький размер батча
		BatchTimeout: 500 * time.Millisecond,
		Workers:      1,
	}

	mockClient := NewMockExternalAPIClient()
	mockCache := NewMockCacheRepository()

	batcher := batcher.NewBatcher(config, mockCache, mockClient)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := batcher.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start batcher: %v", err)
	}
	defer batcher.Stop()

	trackCodes := []string{"BATCH001", "BATCH002", "BATCH003", "BATCH004", "BATCH005", "BATCH006", "BATCH007"}

	var wg sync.WaitGroup
	responses := make([]models.TrackResponse, len(trackCodes))

	for i, trackCode := range trackCodes {
		wg.Add(1)
		go func(i int, code string) {
			defer wg.Done()
			responseChan := batcher.AddRequest(ctx, code)
			select {
			case response := <-responseChan:
				responses[i] = response
			case <-ctx.Done():
				t.Errorf("Request %d timed out", i)
			}
		}(i, trackCode)
	}

	wg.Wait()

	for i, response := range responses {
		if !response.Status {
			t.Errorf("Request %d failed: %s", i, response.Error)
		}
		if response.Data == nil {
			t.Errorf("Request %d has no data", i)
		}
	}

	requestCount := mockClient.GetRequestCount()
	if requestCount != 7 { // 7 трек-кодов в 3 запросах
		t.Errorf("Expected 7 requests to external API, got %d", requestCount)
	}

	requests := mockClient.GetRequests()
	if len(requests) != 7 {
		t.Errorf("Expected 7 track codes, got %d", len(requests))
	}
}

// TestBatcherConcurrency - тест конкурентности
func TestBatcherConcurrency(t *testing.T) {
	config := config.BatcherConfig{
		BatchSize:    5,
		BatchTimeout: 1 * time.Second,
		Workers:      2,
	}

	mockClient := NewMockExternalAPIClient()
	mockCache := NewMockCacheRepository()

	batcher := batcher.NewBatcher(config, mockCache, mockClient)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := batcher.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start batcher: %v", err)
	}
	defer batcher.Stop()

	numRequests := 10
	trackCodes := make([]string, numRequests)
	for i := 0; i < numRequests; i++ {
		trackCodes[i] = fmt.Sprintf("CONCURRENT%03d", i)
	}

	var wg sync.WaitGroup
	responses := make([]models.TrackResponse, numRequests)

	for i, trackCode := range trackCodes {
		wg.Add(1)
		go func(i int, code string) {
			defer wg.Done()
			responseChan := batcher.AddRequest(ctx, code)
			select {
			case response := <-responseChan:
				responses[i] = response
			case <-ctx.Done():
				t.Errorf("Request %d timed out", i)
			}
		}(i, trackCode)
	}

	wg.Wait()

	successCount := 0
	for i, response := range responses {
		if response.Status {
			successCount++
		} else {
			t.Errorf("Request %d failed: %s", i, response.Error)
		}
	}

	if successCount != numRequests {
		t.Errorf("Expected %d successful requests, got %d", numRequests, successCount)
	}

	requests := mockClient.GetRequests()
	if len(requests) != numRequests {
		t.Errorf("Expected %d track codes, got %d", numRequests, len(requests))
	}
}
