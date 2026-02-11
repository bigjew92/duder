package main

import (
	"os"
	"time"
)

func init() {
	Duder.Scheduler = new(SchedulerManager)
}

// SchedulerManager handles scheduled tasks
type SchedulerManager struct {
	done         chan struct{}
	lastPostDate string
}

// Start begins the scheduler loop
func (manager *SchedulerManager) Start() {
	manager.done = make(chan struct{})
	Duder.Log(LogGeneral, "Starting scheduler...")

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				manager.checkAndPostFridayVideo()
			case <-manager.done:
				return
			}
		}
	}()
}

// checkAndPostFridayVideo checks if it's Friday at 9 AM PST and posts the video
func (manager *SchedulerManager) checkAndPostFridayVideo() {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		Duder.Logf(LogWarning, "Failed to load timezone: %s", err.Error())
		return
	}

	now := time.Now().In(loc)
	today := now.Format("2006-01-02")

	// Check if it's Friday at 9 AM and we haven't posted today
	if now.Weekday() == time.Friday && now.Hour() == 9 && manager.lastPostDate != today {
		Duder.Log(LogGeneral, "It's Friday at 9 AM PST! Posting video...")
		manager.postFridayVideo()
		manager.lastPostDate = today
	}
}

// postFridayVideo sends the configured video to the configured channel
func (manager *SchedulerManager) postFridayVideo() {
	videoPath := Duder.Config.FridayVideoPath()
	channelID := Duder.Config.FridayVideoChannelID()

	if len(videoPath) == 0 || len(channelID) == 0 {
		Duder.Log(LogWarning, "Friday video path or channel ID not configured")
		return
	}

	// Check if the video file exists
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		Duder.Logf(LogWarning, "Friday video file not found: %s", videoPath)
		return
	}

	err := Duder.Discord.SendVideoToChannel(channelID, videoPath, "Happy Friday! 🎉")
	if err != nil {
		Duder.Logf(LogWarning, "Failed to post Friday video: %s", err.Error())
	} else {
		Duder.Log(LogGeneral, "Successfully posted Friday video!")
	}
}

// teardown gracefully stops the scheduler
func (manager *SchedulerManager) teardown() {
	if manager.done != nil {
		close(manager.done)
	}
}
