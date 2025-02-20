package kafkaProducerConsumer

type ConsumerData struct {
	TypeOfMessage string `json:"type"`
	Data          any    `json:"data"`
}
