package models

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
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	//Cds         int           `json:"cds"`
	//Comments    []SubComments `json:"comments,omitempty"`
	Group      []string `json:"group"`
	Quality    []string `json:"quality"`
	Resolution []string `json:"resolution"`
	Duration   []string `json:"duration"`
}
