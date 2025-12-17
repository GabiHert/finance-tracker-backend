// Package aicategorization contains AI categorization-related use cases.
package aicategorization

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestInMemoryProcessingTracker_ErrorTracking(t *testing.T) {
	tracker := NewInMemoryProcessingTracker()
	userID := uuid.New()

	// Test HasError returns false when no error exists.
	t.Run("HasError returns false when no error exists", func(t *testing.T) {
		if tracker.HasError(userID) {
			t.Error("expected HasError to return false for non-existent error")
		}
	})

	// Test GetError returns nil when no error exists.
	t.Run("GetError returns nil when no error exists", func(t *testing.T) {
		if tracker.GetError(userID) != nil {
			t.Error("expected GetError to return nil for non-existent error")
		}
	})

	// Test SetError stores the error.
	t.Run("SetError stores the error", func(t *testing.T) {
		testError := &ProcessingError{
			Code:      ErrCodeAIRateLimited,
			Message:   errorMessages[ErrCodeAIRateLimited],
			Retryable: true,
			Timestamp: time.Now(),
		}

		tracker.SetError(userID, testError)

		if !tracker.HasError(userID) {
			t.Error("expected HasError to return true after SetError")
		}

		retrieved := tracker.GetError(userID)
		if retrieved == nil {
			t.Fatal("expected GetError to return non-nil error")
		}

		if retrieved.Code != testError.Code {
			t.Errorf("expected code %s, got %s", testError.Code, retrieved.Code)
		}

		if retrieved.Message != testError.Message {
			t.Errorf("expected message %s, got %s", testError.Message, retrieved.Message)
		}

		if retrieved.Retryable != testError.Retryable {
			t.Errorf("expected retryable %v, got %v", testError.Retryable, retrieved.Retryable)
		}
	})

	// Test SetError overwrites existing error.
	t.Run("SetError overwrites existing error", func(t *testing.T) {
		newError := &ProcessingError{
			Code:      ErrCodeAIServiceUnavailable,
			Message:   errorMessages[ErrCodeAIServiceUnavailable],
			Retryable: true,
			Timestamp: time.Now(),
		}

		tracker.SetError(userID, newError)

		retrieved := tracker.GetError(userID)
		if retrieved == nil {
			t.Fatal("expected GetError to return non-nil error")
		}

		if retrieved.Code != ErrCodeAIServiceUnavailable {
			t.Errorf("expected code %s, got %s", ErrCodeAIServiceUnavailable, retrieved.Code)
		}
	})

	// Test ClearError removes the error.
	t.Run("ClearError removes the error", func(t *testing.T) {
		tracker.ClearError(userID)

		if tracker.HasError(userID) {
			t.Error("expected HasError to return false after ClearError")
		}

		if tracker.GetError(userID) != nil {
			t.Error("expected GetError to return nil after ClearError")
		}
	})

	// Test ClearError on non-existent error does not panic.
	t.Run("ClearError on non-existent error does not panic", func(t *testing.T) {
		nonExistentUserID := uuid.New()
		// This should not panic.
		tracker.ClearError(nonExistentUserID)
	})

	// Test error tracking is user-specific.
	t.Run("error tracking is user-specific", func(t *testing.T) {
		user1 := uuid.New()
		user2 := uuid.New()

		error1 := &ProcessingError{
			Code:      ErrCodeAIRateLimited,
			Message:   errorMessages[ErrCodeAIRateLimited],
			Retryable: true,
			Timestamp: time.Now(),
		}

		error2 := &ProcessingError{
			Code:      ErrCodeAIAuthError,
			Message:   errorMessages[ErrCodeAIAuthError],
			Retryable: false,
			Timestamp: time.Now(),
		}

		tracker.SetError(user1, error1)
		tracker.SetError(user2, error2)

		retrieved1 := tracker.GetError(user1)
		retrieved2 := tracker.GetError(user2)

		if retrieved1.Code != ErrCodeAIRateLimited {
			t.Errorf("user1: expected code %s, got %s", ErrCodeAIRateLimited, retrieved1.Code)
		}

		if retrieved2.Code != ErrCodeAIAuthError {
			t.Errorf("user2: expected code %s, got %s", ErrCodeAIAuthError, retrieved2.Code)
		}

		// Clear user1 error should not affect user2.
		tracker.ClearError(user1)

		if tracker.HasError(user1) {
			t.Error("user1: expected HasError to return false after ClearError")
		}

		if !tracker.HasError(user2) {
			t.Error("user2: expected HasError to still return true")
		}
	})
}

func TestInMemoryProcessingTracker_ThreadSafety(t *testing.T) {
	tracker := NewInMemoryProcessingTracker()
	userIDs := make([]uuid.UUID, 10)
	for i := range userIDs {
		userIDs[i] = uuid.New()
	}

	const goroutines = 100
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Run concurrent operations to verify no race conditions.
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()

			userID := userIDs[id%len(userIDs)]

			for j := 0; j < iterations; j++ {
				// Mix of processing and error operations.
				switch j % 8 {
				case 0:
					tracker.SetProcessing(userID, uuid.New().String())
				case 1:
					tracker.IsProcessing(userID)
				case 2:
					tracker.GetJobID(userID)
				case 3:
					tracker.ClearProcessing(userID)
				case 4:
					tracker.SetError(userID, &ProcessingError{
						Code:      ErrCodeAIRateLimited,
						Message:   errorMessages[ErrCodeAIRateLimited],
						Retryable: true,
						Timestamp: time.Now(),
					})
				case 5:
					tracker.GetError(userID)
				case 6:
					tracker.HasError(userID)
				case 7:
					tracker.ClearError(userID)
				}
			}
		}(i)
	}

	wg.Wait()
	// If we reach here without data race panic, the test passes.
}

func TestInMemoryProcessingTracker_ProcessingMethods(t *testing.T) {
	tracker := NewInMemoryProcessingTracker()
	userID := uuid.New()
	jobID := "test-job-123"

	// Test IsProcessing returns false when not processing.
	t.Run("IsProcessing returns false when not processing", func(t *testing.T) {
		if tracker.IsProcessing(userID) {
			t.Error("expected IsProcessing to return false")
		}
	})

	// Test GetJobID returns empty string when not processing.
	t.Run("GetJobID returns empty string when not processing", func(t *testing.T) {
		if tracker.GetJobID(userID) != "" {
			t.Error("expected GetJobID to return empty string")
		}
	})

	// Test SetProcessing sets the processing state.
	t.Run("SetProcessing sets the processing state", func(t *testing.T) {
		tracker.SetProcessing(userID, jobID)

		if !tracker.IsProcessing(userID) {
			t.Error("expected IsProcessing to return true after SetProcessing")
		}

		if tracker.GetJobID(userID) != jobID {
			t.Errorf("expected jobID %s, got %s", jobID, tracker.GetJobID(userID))
		}
	})

	// Test ClearProcessing clears the processing state.
	t.Run("ClearProcessing clears the processing state", func(t *testing.T) {
		tracker.ClearProcessing(userID)

		if tracker.IsProcessing(userID) {
			t.Error("expected IsProcessing to return false after ClearProcessing")
		}

		if tracker.GetJobID(userID) != "" {
			t.Error("expected GetJobID to return empty string after ClearProcessing")
		}
	})
}
