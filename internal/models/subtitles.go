package models

type PostFilters struct {
	Year       int
	Group      string
	Quality    string
	Resolution string
}

/*type Subtitle struct {
	Title string `json:"title"`
}*/

/* Temporal */
type SubComments struct {
	Id      int    `json:"id"`
	Comment string `json:"comentario"`
	Nick    string `json:"nick"`
	Date    string `json:"fecha_creacion"`
}
type Subtitle struct {
	Provider    string `json:"provider"`
	Type string `json:"type"`
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Language    string `json:"language"`
	//Cds         int           `json:"cds"`
	//Comments    []SubComments `json:"comments,omitempty"`
	Group      []string `json:"group"`
	Quality    []string `json:"quality"`
	Resolution []string `json:"resolution"`
	Duration   []string `json:"duration"`
	Year       int      `json:"year"`
	Season int `json:"season"`
	Episode int `json:"episode"`
}
