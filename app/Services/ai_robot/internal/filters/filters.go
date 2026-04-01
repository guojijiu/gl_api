package filters

type ProjectFilters struct {
	Number         string
	Name           string
	WorkflowNameCN string
	IsFilterTime   int
	IsUsedFree     int
}

type ContractFilters struct {
	ContractNumber string
	Name           string
}

type TaskFilters struct {
	ProjectNumber  string
	ProjectName    string
	UUID           string
	Name           string
	ToolName       string
	StatusValue    string
	WorkflowNameCN string
	CreatedAtStart string
	CreatedAtEnd   string
}

type ProjectArticleFilterItem struct {
	Column   string `json:"column"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type ProjectArticleFilters struct {
	SearchFilter []ProjectArticleFilterItem
}
