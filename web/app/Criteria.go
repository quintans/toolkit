package app

type Criteria struct {
	CountRecords bool    `json:"countRecords"`
	Page         int64   `json:"page"`
	PageSize     int64   `json:"pageSize"`
	OrderBy      *string `json:"orderBy"`
	Ascending    *bool   `json:"ascending"`
}
