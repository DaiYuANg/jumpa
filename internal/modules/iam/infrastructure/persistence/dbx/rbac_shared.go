package dbx

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

func normalizeIDs(ids []string) []string {
	return lo.Uniq(lo.FilterMap(ids, func(id string, _ int) (string, bool) {
		v := strings.TrimSpace(id)
		return v, v != ""
	}))
}

func principalIDByUser(userID int64) string { return fmt.Sprintf("user:%d", userID) }

