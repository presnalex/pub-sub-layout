package storage

type DbResponse struct {
	Response []byte `db:"p_addresult"`
}

type AnimalMsg struct {
	AnimalId int32  `json:"animalId"`
	Animal   string `json:"animal"`
	Price    int32  `json:"price"`
}

type ResponseMsg struct {
	MessageId    int32  `json:"messageId"`
	Status       string `json:"status"`
	ErrorDetails string `json:"errorDetails"`
}
