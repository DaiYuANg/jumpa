package domain

import "time"

type Gateway struct {
	ID              string
	NodeKey         string
	NodeName        string
	RuntimeType     string
	AdvertiseAddr   string
	SSHListenAddr   string
	Zone            string
	Tags            []string
	State           string
	EffectiveStatus string
	RegisteredAt    time.Time
	LastSeenAt      time.Time
	UpdatedAt       time.Time
}
