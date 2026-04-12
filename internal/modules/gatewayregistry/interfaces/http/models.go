package http

import "time"

type gatewayDTO struct {
	ID              string    `json:"id"`
	NodeKey         string    `json:"nodeKey"`
	NodeName        string    `json:"nodeName"`
	RuntimeType     string    `json:"runtimeType"`
	AdvertiseAddr   string    `json:"advertiseAddr"`
	SSHListenAddr   string    `json:"sshListenAddr"`
	Zone            string    `json:"zone"`
	Tags            []string  `json:"tags"`
	State           string    `json:"state"`
	EffectiveStatus string    `json:"effectiveStatus"`
	RegisteredAt    time.Time `json:"registeredAt"`
	LastSeenAt      time.Time `json:"lastSeenAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
