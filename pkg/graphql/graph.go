package graphql

type Response struct {
	Errors []Errors `json:"errors"`
	Data   Data     `json:"data"`
}

type Data struct {
	DocumentQuery DocumentQuery `json:"documentQuery"`
}

type DocumentQuery struct {
	GenerateUploadHyperlink Hyperlink `json:"generateUploadHyperlink"`
}

type Hyperlink struct {
	URL  string `json:"url"`
	Verb string `json:"verb"`
}

type Errors struct {
	Message string `json:"message"`
}

type UpdateBuildRunRequest struct {
	Query     string `json:"query"`
	Variables `json:"variables"`
}

type Variables struct {
	ID    string `json:"ID"`
	STATE string `json:"STATE"`
}
