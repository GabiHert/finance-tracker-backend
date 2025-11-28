package mock

import "time"

type Time struct {
	currentStartTime time.Time
	updatedAt        time.Time
}

func NewTime() *Time {
	return &Time{
		currentStartTime: time.Now(),
		updatedAt:        time.Now(),
	}
}

func (t *Time) SetCurrentTime(currentTime time.Time) {
	t.currentStartTime = currentTime
	t.updatedAt = time.Now()
}

func (t *Time) Now() time.Time {
	elapsed := t.updatedAt.Sub(time.Now())
	return t.currentStartTime.Add(elapsed)
}
