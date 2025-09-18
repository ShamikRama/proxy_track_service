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
	Date   string `json:"date"` // RFC3339
}

const (
	StatusCreated   = "Created"   // Посылка создана
	StatusInTransit = "Transit"   // В пути
	StatusInCustoms = "Customs"   // На таможне
	StatusDelivered = "Delivered" // Доставлена
	StatusException = "Exception" // Исключение/проблема
	StatusReturned  = "Returned"  // Возвращена
	StatusUnknown   = "Unknown"   // Неизвестный статус
)
