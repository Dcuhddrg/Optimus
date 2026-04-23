package common

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type PingResponse struct {
	Message string `json:"message"`
}
