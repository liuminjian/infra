package base

type Response struct {
	Status  int         `json:"status"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

type PageData struct {
	Data  interface{} `json:"data"`
	Count uint64      `json:"count"`
}
