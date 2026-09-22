package rms

import "encoding/json"

// Timestamp is either an ISO-ish string or a Unix seconds int, depending on endpoint.
type Timestamp struct {
	Value string
}

func (t Timestamp) String() string { return t.Value }

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		t.Value = s
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	t.Value = n.String()
	return nil
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	if t.Value == "" {
		return []byte("null"), nil
	}
	return json.Marshal(t.Value)
}

// Company mirrors the /companies entity envelope.
type Company struct {
	ID          json.Number `json:"id"`
	Name        string      `json:"name"`
	ParentID    json.Number `json:"parent_id,omitempty"`
	Email       string      `json:"email,omitempty"`
	Level       json.Number `json:"level,omitempty"`
	DeviceCount json.Number `json:"device_count,omitempty"`
	CreatedAt   Timestamp   `json:"created_at,omitempty"`
	UpdatedAt   Timestamp   `json:"updated_at,omitempty"`
}

type CompanyCreate struct {
	Name     string `json:"name"`
	ParentID int64  `json:"parent_id"`
}

type CompanyUpdate struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// User mirrors the /users entity envelope.
type User struct {
	ID        json.Number `json:"id"`
	Email     string      `json:"email,omitempty"`
	FirstName string      `json:"first_name,omitempty"`
	LastName  string      `json:"last_name,omitempty"`
	Role      string      `json:"role,omitempty"`
	CompanyID json.Number `json:"company_id,omitempty"`
	CreatedAt Timestamp   `json:"created_at,omitempty"`
	UpdatedAt Timestamp   `json:"updated_at,omitempty"`
}

type InviteUser struct {
	Role      string `json:"role"`
	Email     string `json:"email"`
	CompanyID int64  `json:"company_id"`
}

type UserUpdate struct {
	Role string `json:"role,omitempty"`
}

// Tag mirrors the /tags entity envelope.
type Tag struct {
	ID          json.Number `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Color       string      `json:"color,omitempty"`
	CompanyID   json.Number `json:"company_id,omitempty"`
	CreatedAt   Timestamp   `json:"created_at,omitempty"`
	UpdatedAt   Timestamp   `json:"updated_at,omitempty"`
}

type TagCreate struct {
	Name        string `json:"name"`
	CompanyID   int64  `json:"company_id"`
	Description string `json:"description"`
	Color       string `json:"color,omitempty"`
}

type TagUpdate struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

// Device mirrors the /devices entity envelope — RMS exposes a very wide schema; we surface
// the widely useful attributes and expose the rest as raw JSON for callers that need it.
type Device struct {
	ID           json.Number     `json:"id"`
	Name         string          `json:"name,omitempty"`
	Serial       string          `json:"serial,omitempty"`
	MAC          string          `json:"mac,omitempty"`
	Model        string          `json:"model,omitempty"`
	Manufacturer string          `json:"manufacturer,omitempty"`
	FWVersion    string          `json:"fw_version,omitempty"`
	Status       string          `json:"status,omitempty"`
	CompanyID    json.Number     `json:"company_id,omitempty"`
	Latitude     json.Number     `json:"latitude,omitempty"`
	Longitude    json.Number     `json:"longitude,omitempty"`
	Description  string          `json:"description,omitempty"`
	LastSeen     Timestamp       `json:"last_seen,omitempty"`
	CreatedAt    Timestamp       `json:"created_at,omitempty"`
	UpdatedAt    Timestamp       `json:"updated_at,omitempty"`
	Raw          json.RawMessage `json:"-"`
}

func (d *Device) UnmarshalJSON(b []byte) error {
	type alias Device
	var tmp alias
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	*d = Device(tmp)
	d.Raw = append(d.Raw[:0], b...)
	return nil
}

type DeviceUpdate struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	CompanyID   int64  `json:"company_id,omitempty"`
}

// Alert mirrors the /alerts entity envelope.
type Alert struct {
	ID          json.Number `json:"id"`
	DeviceID    json.Number `json:"device_id,omitempty"`
	Name        string      `json:"name,omitempty"`
	Type        string      `json:"type,omitempty"`
	Severity    string      `json:"severity,omitempty"`
	Description string      `json:"description,omitempty"`
	Message     string      `json:"message,omitempty"`
	Status      string      `json:"status,omitempty"`
	CreatedAt   Timestamp   `json:"created_at,omitempty"`
}

// AlertConfiguration mirrors /alerts-configurations entity envelope.
type AlertConfiguration struct {
	ID         json.Number     `json:"id"`
	Name       string          `json:"name,omitempty"`
	Type       string          `json:"type,omitempty"`
	CompanyID  json.Number     `json:"company_id,omitempty"`
	Enabled    bool            `json:"enabled,omitempty"`
	Conditions json.RawMessage `json:"conditions,omitempty"`
	Actions    json.RawMessage `json:"actions,omitempty"`
	CreatedAt  Timestamp       `json:"created_at,omitempty"`
	UpdatedAt  Timestamp       `json:"updated_at,omitempty"`
}

// EmailConfiguration mirrors /email-configurations entity envelope.
type EmailConfiguration struct {
	ID        json.Number `json:"id"`
	Name      string      `json:"name,omitempty"`
	Host      string      `json:"host,omitempty"`
	Port      json.Number `json:"port,omitempty"`
	Email     string      `json:"email,omitempty"`
	Username  string      `json:"username,omitempty"`
	CompanyID json.Number `json:"company_id,omitempty"`
	CreatedAt Timestamp   `json:"created_at,omitempty"`
	UpdatedAt Timestamp   `json:"updated_at,omitempty"`
}

type EmailConfigurationCreate struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type EmailConfigurationUpdate struct {
	Name     string `json:"name,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// Role mirrors /roles entity envelope.
type Role struct {
	ID          json.Number   `json:"id"`
	Title       string        `json:"title,omitempty"`
	Description string        `json:"description,omitempty"`
	CompanyID   []json.Number `json:"company_id,omitempty"`
	Permissions []json.Number `json:"permission_id,omitempty"`
	CreatedAt   Timestamp     `json:"created_at,omitempty"`
	UpdatedAt   Timestamp     `json:"updated_at,omitempty"`
}

type RoleCreate struct {
	Title        string  `json:"title"`
	Description  string  `json:"description,omitempty"`
	CompanyIDs   []int64 `json:"company_id"`
	PermissionID []int64 `json:"permission_id"`
}

type RoleUpdate struct {
	ID           int64   `json:"id"`
	Title        string  `json:"title,omitempty"`
	Description  string  `json:"description,omitempty"`
	PermissionID []int64 `json:"permission_id,omitempty"`
}

// Permission mirrors permission_body from OpenAPI.
type Permission struct {
	ID          json.Number `json:"id"`
	Title       string      `json:"title,omitempty"`
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Category    string      `json:"category,omitempty"`
}

// File mirrors /files entity envelope (metadata only; upload uses multipart separately).
type File struct {
	ID          json.Number `json:"id"`
	Name        string      `json:"name,omitempty"`
	Type        string      `json:"type,omitempty"`
	Description string      `json:"description,omitempty"`
	CompanyID   json.Number `json:"company_id,omitempty"`
	Size        json.Number `json:"size,omitempty"`
	CreatedAt   Timestamp   `json:"created_at,omitempty"`
	UpdatedAt   Timestamp   `json:"updated_at,omitempty"`
}

// Automation mirrors /automations entity envelope.
type Automation struct {
	ID          json.Number     `json:"id"`
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	CompanyID   json.Number     `json:"company_id,omitempty"`
	Enabled     bool            `json:"enabled,omitempty"`
	Trigger     json.RawMessage `json:"trigger,omitempty"`
	Conditions  json.RawMessage `json:"conditions,omitempty"`
	Actions     json.RawMessage `json:"actions,omitempty"`
	CreatedAt   Timestamp       `json:"created_at,omitempty"`
	UpdatedAt   Timestamp       `json:"updated_at,omitempty"`
}

// VPNHub mirrors /vpn/hubs entity envelope.
type VPNHub struct {
	ID          json.Number   `json:"id"`
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty"`
	CompanyID   json.Number   `json:"company_id,omitempty"`
	HubZone     string        `json:"hub_zone,omitempty"`
	VPNType     string        `json:"vpn_type,omitempty"`
	Enabled     bool          `json:"enabled,omitempty"`
	TagIDs      []json.Number `json:"tag_id,omitempty"`
	CreatedAt   Timestamp     `json:"created_at,omitempty"`
	UpdatedAt   Timestamp     `json:"updated_at,omitempty"`
}

type VPNHubCreate struct {
	Name        string  `json:"name"`
	CompanyID   int64   `json:"company_id,omitempty"`
	Description string  `json:"description,omitempty"`
	HubZone     string  `json:"hub_zone"`
	VPNType     string  `json:"vpn_type,omitempty"`
	TagID       []int64 `json:"tag_id,omitempty"`
}

type VPNHubUpdate struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	VPNType     string `json:"vpn_type,omitempty"`
}

// VPNHubUser mirrors /vpn/hubs/users entity envelope.
type VPNHubUser struct {
	ID        json.Number `json:"id"`
	Name      string      `json:"name,omitempty"`
	Username  string      `json:"username,omitempty"`
	HubID     json.Number `json:"hub_id,omitempty"`
	Enabled   bool        `json:"enabled,omitempty"`
	CreatedAt Timestamp   `json:"created_at,omitempty"`
	UpdatedAt Timestamp   `json:"updated_at,omitempty"`
}

type VPNHubUserCreate struct {
	Name  string `json:"name"`
	HubID int64  `json:"hub_id"`
}

// DataCollectConfig mirrors /data-collect/configs entity envelope.
type DataCollectConfig struct {
	ID          json.Number     `json:"id"`
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	CompanyID   json.Number     `json:"company_id,omitempty"`
	Interval    json.Number     `json:"interval,omitempty"`
	Enabled     bool            `json:"enabled,omitempty"`
	Fields      json.RawMessage `json:"fields,omitempty"`
	CreatedAt   Timestamp       `json:"created_at,omitempty"`
	UpdatedAt   Timestamp       `json:"updated_at,omitempty"`
}

// ConfiguratorTemplate mirrors /devices/configurator/templates entity.
type ConfiguratorTemplate struct {
	ID          json.Number     `json:"id"`
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	CompanyID   json.Number     `json:"company_id,omitempty"`
	Model       string          `json:"model,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
	CreatedAt   Timestamp       `json:"created_at,omitempty"`
	UpdatedAt   Timestamp       `json:"updated_at,omitempty"`
}

// CreditsSummary is the /credits/summary envelope.
type CreditsSummary struct {
	Data json.RawMessage `json:"data"`
}

// DeviceMonitoring is one entry from /devices/monitoring.
type DeviceMonitoring struct {
	ID     json.Number     `json:"id"`
	Name   string          `json:"name,omitempty"`
	Status string          `json:"status,omitempty"`
	Raw    json.RawMessage `json:"-"`
}

func (d *DeviceMonitoring) UnmarshalJSON(b []byte) error {
	type alias DeviceMonitoring
	var tmp alias
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	*d = DeviceMonitoring(tmp)
	d.Raw = append(d.Raw[:0], b...)
	return nil
}

// Hotspot is one entry from /hotspots.
type Hotspot struct {
	ID       json.Number `json:"id"`
	Name     string      `json:"name,omitempty"`
	SSID     string      `json:"ssid,omitempty"`
	DeviceID json.Number `json:"device_id,omitempty"`
	Enabled  bool        `json:"enabled,omitempty"`
}
