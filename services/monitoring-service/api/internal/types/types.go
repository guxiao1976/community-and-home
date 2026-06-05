package types

type HealthResponse struct {
	Timestamp     string            `json:"timestamp"`
	OverallStatus string            `json:"overall_status"`
	Services      []ServiceHealth   `json:"services"`
	Containers    []ContainerHealth `json:"containers"`
	AiModels      []AiModelHealth   `json:"ai_models"`
}

type ServiceHealth struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	ApiPort     int    `json:"api_port"`
	ApiStatus   string `json:"api_status"`
	ApiError    string `json:"api_error,omitempty"`
	RpcPort     int    `json:"rpc_port"`
	RpcStatus   string `json:"rpc_status"`
	RpcError    string `json:"rpc_error,omitempty"`
}

type ContainerHealth struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Image       string `json:"image"`
	Status      string `json:"status"`
	RunningFor  string `json:"running_for"`
	Error       string `json:"error,omitempty"`
}

type AiModelHealth struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Provider    string `json:"provider"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
}
