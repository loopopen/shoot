package dto

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type QueryUsersReq struct {
	Name     string `rest:"alias=name"`
	PageSize int    `rest:"alias=size"`
	PageIdx  int    `rest:"alias=page_idx"`
	Ignore   string `rest:"-"`
}

type QueryUsersResp struct{}
