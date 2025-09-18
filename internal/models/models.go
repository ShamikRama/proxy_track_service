package models

type TrackRequest struct {
	TrackCode string `json:"track_code"`
}

type BatchTrackRequest struct {
	TrackCodes []string `json:"track_codes"`
}

type TrackResponse struct {
	Status bool       `json:"status"`
	Data   *TrackData `json:"data,omitempty"`
	Error  string     `json:"error,omitempty"`
}

type TrackData struct {
	Countries []string `json:"countries"`
	Events    []Event  `json:"events"`
}

type Event struct {
	Status string `json:"status"`
	Date   string `json:"date"`
}

const (
	StatusCreated   = "Created"
	StatusInTransit = "Transit"
	StatusInCustoms = "Customs"
	StatusDelivered = "Delivered"
	StatusException = "Exception"
	StatusReturned  = "Returned"
	StatusUnknown   = "Unknown"
)
