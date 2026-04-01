package asciicast

// Header is the first line of an asciicast v2 file.
type Header struct {
	Version       int               `json:"version"`
	Width         int               `json:"width"`
	Height        int               `json:"height"`
	Timestamp     int64             `json:"timestamp,omitempty"`
	Duration      float64           `json:"duration,omitempty"`
	IdleTimeLimit float64           `json:"idle_time_limit,omitempty"`
	Command       string            `json:"command,omitempty"`
	Title         string            `json:"title,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
}
