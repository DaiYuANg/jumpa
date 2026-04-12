package api

import "time"

type Result[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type PageResult[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type Me struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	User         Me     `json:"user"`
}

type Overview struct {
	ProductName       string   `json:"productName"`
	DatabaseDriver    string   `json:"databaseDriver"`
	BastionEnabled    bool     `json:"bastionEnabled"`
	SSHListenAddr     string   `json:"sshListenAddr"`
	RecordingDir      string   `json:"recordingDir"`
	IdentityModes     []string `json:"identityModes"`
	PasswordAuthReady bool     `json:"passwordAuthReady"`
	CapabilityNotes   []string `json:"capabilityNotes"`
}

type Host struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Address         string    `json:"address"`
	Port            int       `json:"port"`
	Protocol        string    `json:"protocol"`
	Environment     string    `json:"environment"`
	Platform        string    `json:"platform"`
	Authentication  string    `json:"authentication"`
	JumpEnabled     bool      `json:"jumpEnabled"`
	RecordingPolicy string    `json:"recordingPolicy"`
	CreatedAt       time.Time `json:"createdAt"`
}

type AccessRequest struct {
	ID             string     `json:"id"`
	PrincipalName  string     `json:"principalName"`
	PrincipalEmail string     `json:"principalEmail"`
	HostName       string     `json:"hostName"`
	HostAccount    string     `json:"hostAccount"`
	Protocol       string     `json:"protocol"`
	Status         string     `json:"status"`
	RequestedAt    time.Time  `json:"requestedAt"`
	ReviewedAt     *time.Time `json:"reviewedAt"`
}

type Session struct {
	ID            string     `json:"id"`
	HostName      string     `json:"hostName"`
	HostAccount   string     `json:"hostAccount"`
	PrincipalName string     `json:"principalName"`
	Protocol      string     `json:"protocol"`
	Status        string     `json:"status"`
	StartedAt     time.Time  `json:"startedAt"`
	EndedAt       *time.Time `json:"endedAt"`
}

type Gateway struct {
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
