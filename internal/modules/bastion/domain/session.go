package domain

import "time"

type Session struct {
	ID            string     `json:"id"`
	HostName      string     `json:"hostName"`
	HostAccount   string     `json:"hostAccount"`
	PrincipalName string     `json:"principalName"`
	Protocol      string     `json:"protocol"`
	Status        string     `json:"status"`
	StartedAt     time.Time  `json:"startedAt"`
	EndedAt       *time.Time `json:"endedAt,omitempty"`
}
