package providers

type SubComments struct {
	Id      int    `json:"id"`
	Comment string `json:"comentario"`
	Nick    string `json:"nick"`
	Date    string `json:"fecha_creacion"`
}

type SubData struct {
	Id          int    `json:"id"`
	Title       string `json:"titulo"`
	Description string `json:"descripcion"`
	Cds         int    `json:"cds"`
	Downloads   int    `json:"descargas"`
	Comments    int    `json:"comentarios"`
}

type SubdivxResponse[T any] struct {
	Secho                string `json:"sEcho"`
	ITotalRecords        int    `json:"iTotalRecords"`
	ITotalDisplayRecords int    `json:"iTotalDisplayRecords"`
	Data                 []T    `json:"aaData"`
}

type SubdivxSubPayload struct {
	Tabla   string `json:"tabla"`
	Filtros string `json:"filtros,omitempty"`
	Buscar  string `json:"buscar"`
	Token   string `json:"token"`
}

type SubdivxCommentPayload struct {
	GetComments string `json:"getComentatios"`
}
