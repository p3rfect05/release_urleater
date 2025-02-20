package dto

type ConsumerData struct {
	TypeOfMessage string            `json:"type"`
	Data          map[string]string `json:"data"`
}
