package models

type User struct {
	Uid      string `json:"uid"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Node struct {
	NodeId   string `json:"node_id"`
	Address  string `json:"address"`
	Model    string `json:"model"`
	Ram      int `json:"ram"`
	SellerId string `json:"seller_id"`
}

type Job struct {
	JobId          string `json:"job_id"`
	BuyerId        string `json:"buyer_id"`
	ImageName      string   `json:"image_name"`
	Command        []string `json:"command"`
	Args           []string `json:"args"`
	CPULimits      string   `json:"cpu_limits"`
	MemoryLimits   string   `json:"memory_limits"`
	Namespace      string   `json:"namespace"`
	JobName        string   `json:"job_name"`
}
