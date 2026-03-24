package repo

import "fmt"

func principalIDByUser(userID int64) string {
	return fmt.Sprintf("user:%d", userID)
}
