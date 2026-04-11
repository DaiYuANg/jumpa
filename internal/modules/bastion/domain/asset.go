package domain

import "time"

type Host struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Address         string    `json:"address"`
	Port            int       `json:"port"`
	Protocol        string    `json:"protocol"`
	Environment     string    `json:"environment,omitempty"`
	Platform        string    `json:"platform,omitempty"`
	Authentication  string    `json:"authentication"`
	JumpEnabled     bool      `json:"jumpEnabled"`
	RecordingPolicy string    `json:"recordingPolicy"`
	CreatedAt       time.Time `json:"createdAt"`
}

type HostAccount struct {
	ID                 string    `json:"id"`
	HostID             string    `json:"hostId"`
	AccountName        string    `json:"accountName"`
	AuthenticationType string    `json:"authenticationType"`
	CredentialRef      *string   `json:"credentialRef,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
}
