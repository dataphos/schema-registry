package apicuriosr

type insertInfo struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	CreatedBy   string        `json:"createdBy"`
	CreatedOn   string        `json:"createdOn"`
	ModifiedBy  string        `json:"modifiedBy"`
	ModifiedOn  string        `json:"modifiedOn"`
	Id          string        `json:"id"`
	Version     string        `json:"version"`
	Type        string        `json:"type"`
	GlobalId    int64         `json:"globalId"`
	State       string        `json:"state"`
	GroupId     string        `json:"groupId"`
	ContentId   int64         `json:"contentId"`
	Labels      []string      `json:"labels"`
	Properties  []interface{} `json:"properties"`
	References  []interface{} `json:"references"`
}
