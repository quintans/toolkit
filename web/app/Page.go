package app

type Page struct {
	// max number of records returned by the datasource
	Count *int64 `json:"count"`
	// if this is the last page
	Last    bool          `json:"last"`
	Results []interface{} `json:"results"`
}
