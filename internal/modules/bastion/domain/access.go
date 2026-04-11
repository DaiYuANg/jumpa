package domain

import "time"

type AccessPolicy struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	SubjectType       string    `json:"subjectType"`
	SubjectName       string    `json:"subjectName"`
	TargetType        string    `json:"targetType"`
	TargetName        string    `json:"targetName"`
	AccountPattern    string    `json:"accountPattern"`
	Protocol          string    `json:"protocol"`
	ApprovalRequired  bool      `json:"approvalRequired"`
	RecordingRequired bool      `json:"recordingRequired"`
	CreatedAt         time.Time `json:"createdAt"`
}

type AccessRequest struct {
	ID                string     `json:"id"`
	PolicyID          string     `json:"policyId"`
	PrincipalName     string     `json:"principalName"`
	PrincipalEmail    string     `json:"principalEmail,omitempty"`
	HostName          string     `json:"hostName"`
	HostAccount       string     `json:"hostAccount"`
	Protocol          string     `json:"protocol"`
	Status            string     `json:"status"`
	RequestedAt       time.Time  `json:"requestedAt"`
	ReviewedAt        *time.Time `json:"reviewedAt,omitempty"`
	ReviewedBy        *string    `json:"reviewedBy,omitempty"`
	ReviewComment     *string    `json:"reviewComment,omitempty"`
	ApprovedUntil     *time.Time `json:"approvedUntil,omitempty"`
	ConsumedAt        *time.Time `json:"consumedAt,omitempty"`
	ConsumedSessionID *string    `json:"consumedSessionId,omitempty"`
}
