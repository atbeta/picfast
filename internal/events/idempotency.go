package events

import "fmt"

func ImageUploadedKey(imageID int64) string {
	return fmt.Sprintf("image.uploaded:%d", imageID)
}

func ImageProcessedKey(imageID int64) string {
	return fmt.Sprintf("image.processed:%d", imageID)
}

func ImageDeletedKey(imageID int64) string {
	return fmt.Sprintf("image.deleted:%d", imageID)
}

func ModerationReviewedKey(imageID int64, status string) string {
	return fmt.Sprintf("moderation.reviewed:%d:%s", imageID, status)
}

func UserRegisteredKey(userID int64) string {
	return fmt.Sprintf("user.registered:%d", userID)
}
