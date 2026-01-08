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

/* OpenSubtitles Api Response */
type OpenSubtitlesResponse[T any] struct {
	TotalPages int `json:"total_pages"`
	TotalCount int `json:"total_count"`
	PerPage    int `json:"per_page"`
	Page       int `json:"page"`
	Data       []T `json:"data"`
}

type OpenSubtitlesItemFile struct {
	FileId   int    `json:"file_id"`
	CdNumber int    `json:"cd_number"`
	FileName string `json:"file_name"`
}

type OpenSubtitlesItemFeature struct {
	FeatureId   int    `json:"feature_id"`
	FeatureType string `json:"feature_type"`
	Year        int    `json:"year"`
	Title       string `json:"title"`
	MovieName   string `json:"movie_name"`
	ImdbId      int    `json:"imdb_id"`
	TmdbId      int    `json:"tmdb_id"`
}

type OpenSubtitlesItemAttr struct {
	SubtitleId        string   `json:"subtitle_id"`
	Language          string   `json:"language"`
	DownloadCount     int      `json:"download_count"`
	NewDownloadCount  int      `json:"new_download_count"`
	HearingImpaired   bool     `json:"hearing_impaired"`
	Hd                bool     `json:"hd"`
	Fps               float64  `json:"fps"`
	Votes             int      `json:"votes"`
	Ratings           float64  `json:"ratings"`
	FromTrusted       bool     `json:"from_trusted"`
	ForeignPartsOnly  bool     `json:"foreign_parts_only"`
	UploadDate        string   `json:"upload_date"`
	FileHashes        []string `json:"file_hashes,omitempty"`
	AiTranslated      bool     `json:"ai_translated"`
	NbCd              int      `json:"nb_cd"`
	Slug              string   `json:"slug"`
	MachineTranslated bool     `json:"machine_translated"`
	Release           string   `json:"release"`
	Comments          string   `json:"comments,omitempty"`
	LegacySubtitleId  int      `json:"legacy_subtitle_id"`
	LegacyUploaderId  int      `json:"legacy_uploader_id"`
	Uploader          struct {
		UploaderId int    `json:"uploader_id,omitempty"`
		Name       string `json:"name,omitempty"`
		Rank       string `json:"rank,omitempty"`
	} `json:"uploader"`
	FeatureDetails OpenSubtitlesItemFeature `json:"feature_details"`
	Url            string                   `json:"url"`
	RelatedLinks   []struct {
		Label  string `json:"label"`
		Url    string `json:"url"`
		ImgUrl string `json:"image_url"`
	} `json:"related_links"`
	Files []OpenSubtitlesItemFile `json:"files"`
}

type OpenSubtitlesItem struct {
	Id         string                `json:"id"`
	Type       string                `json:"type"`
	Attributes OpenSubtitlesItemAttr `json:"attributes"`
}

type SubXResponse[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

type SubXResponseItem struct {
	Id           string `json:"id"`
	VideoType    string `json:"video_type"`
	Title        string `json:"title"`
	Season       int    `json:"season,omitempty"`
	Episode      int    `json:"episode,omitempty"`
	ImdbId       string `json:"imdb_id"`
	Description  string `json:"description"`
	UploaderName string `json:"uploader_name"`
	PostedAt     string `json:"posted_at"`
	Downloads    int    `json:"downloads"`
}
