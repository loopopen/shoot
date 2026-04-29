package dto

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type QueryUsersReq struct {
	Name     string `shoot:"alias=name"`
	PageSize int    `shoot:"alias=size"`
	PageIdx  int    `shoot:"alias=page_idx"`
}

type QueryUsersResp struct{}
