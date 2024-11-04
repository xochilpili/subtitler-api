package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/microcosm-cc/bluemonday"
	"github.com/xochilpili/subtitler-api/internal/models"
)

type Token struct {
	Cookie string `json:"cookie,omitempty"`
	Token  string `json:"token"`
}

func searchDivx(provider *ProviderParams, query string) []models.Subtitle {
	ctx, cancel := context.WithTimeout(provider.ctx, 30*time.Second)
	provider.ctx = ctx
	defer cancel()
	provider.logger.Info().Msgf("searching subtitles for: %s", query)
	version, err := getVersion(provider)
	if err != nil {
		return nil
	}

	token, err := getToken(provider)
	if err != nil {
		return nil
	}

	provider.logger.Debug().Msgf("token: %s, cookie: %s", token.Token, token.Cookie)

	params := &SubdivxSubPayload{
		Tabla:   "resultados",
		Filtros: "",
		Buscar:  query,
		Token:   token.Token,
	}
	buscaVersion := fmt.Sprintf("buscar%s", version)

	queryParams := map[string]string{
		"tabla":      params.Tabla,
		"filtros":    params.Filtros,
		buscaVersion: params.Buscar,
		"token":      params.Token,
	}
	data, err := getSubtitles(provider, queryParams)
	if err != nil {
		provider.logger.Err(err).Msg("error while getting subtitles")
		return nil
	}

	return data
}

func getVersion(provider *ProviderParams) (string, error) {
	res, err := provider.r.R().Get(provider.config.url)
	if err != nil {
		provider.logger.Err(err).Msg("error while getting version")
		return "", errors.New("error while requesting version")
	}
	re := regexp.MustCompile(`<div[^>]*id="vs"[^>]*>([^<]+)</div>`)
	match := re.FindStringSubmatch(string(res.Body()))
	if len(match) > 1 {
		version := match[1]
		return strings.Trim(strings.Replace(strings.TrimPrefix(version, "v"), ".", "", -1), "\n"), nil
	}
	return "", errors.New("error while parsing version")
}

func getToken(provider *ProviderParams) (*Token, error) {
	var token Token
	res, err := provider.r.R().
		SetContext(provider.ctx).
		SetHeaders(map[string]string{"Content-Type": "application/json", "User-Agent": provider.config.userAgent}).
		SetQueryParam("gt", "1").
		SetDebug(provider.config.debug).
		Get(provider.config.url + "inc/gt.php")

	if err != nil {
		provider.logger.Err(err).Msgf("error while getting token")
		return nil, err
	}

	err = json.Unmarshal(res.Body(), &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func getSubtitles(provider *ProviderParams, params map[string]string) ([]models.Subtitle, error) {
	
	provider.r.SetRetryCount(5).SetRetryWaitTime(5*time.Second)
	provider.r.AddRetryCondition(func(r *resty.Response, _ error) bool {
		var tempResult SubdivxResponse[SubData]
		errs := json.Unmarshal(r.Body(), &tempResult)
		if errs != nil{
			return false
		}
		ok, err := strconv.Atoi(tempResult.Secho)
		if err != nil{
			return false
		}
		return ok == 0
	})
	
	var result SubdivxResponse[SubData]
	resp, err := provider.r.R().
		SetContext(provider.ctx).
		SetFormData(params).
		SetHeaders(map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"User-Agent":   provider.config.userAgent,
		}).
		SetDebug(provider.config.debug).
		Post(provider.config.url + provider.config.searchUrl)
	if err != nil {
		provider.logger.Err(err).Msgf("error while getting subtitles")
		return nil, err
	}

	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		provider.logger.Err(err).Msg("error while unmarshal response")
		return nil, err
	}

	wg := &sync.WaitGroup{}
	var subtitles []models.Subtitle
	subtitlesChan := make(chan models.Subtitle, len(result.Data))

	// TODO: Investigate how to resolve error when using go routines
	reg := regexp.MustCompile(`\n|\r\n`)
	for _, item := range result.Data {
		wg.Add(1)
		var group []string
		var quality []string
		var resolution []string
		var duration []string
		var year int
		var itemType string
		var season int
		var episode int
		stripTags := bluemonday.StripTagsPolicy()
		title := reg.ReplaceAllString(stripTags.Sanitize(item.Title), " ")
		desc := reg.ReplaceAllString(stripTags.Sanitize(item.Description), " ")
		itemType, season, episode = parseTitle(title)
		group, quality, resolution, duration = parseExtra(desc)
		y := Parse(title, "year")
		if y != nil {
			yy, _ := strconv.Atoi(y[0])
			year = yy
		}
		subtitle := &models.Subtitle{
			Provider:    "subdivx",
			Type: itemType,
			Id:          item.Id,
			Title:       title,
			Description: desc,
			Language:    "es",
			//Cds:         item.Cds,
			Year: year,
			Season: season,
			Episode: episode,
		}

		subtitle.Group = group
		subtitle.Quality = quality
		subtitle.Resolution = resolution
		subtitle.Duration = duration

		go func(provider *ProviderParams, sub *models.Subtitle, subChan chan<- models.Subtitle, wg *sync.WaitGroup) {
			getComments(provider, subtitle, subChan, wg)
		}(provider, subtitle, subtitlesChan, wg)
	}

	go func() {
		wg.Wait()
		close(subtitlesChan)
	}()

	for item := range subtitlesChan {
		subtitles = append(subtitles, item)
	}
	provider.logger.Info().Msgf("returned %d subtitles", len(subtitles))
	return subtitles, nil
}

func getComments(provider *ProviderParams, subtitle *models.Subtitle, subChan chan<- models.Subtitle, wg *sync.WaitGroup) {
	defer wg.Done()
	var result SubdivxResponse[SubComments]
	res, err := provider.r.R().
		SetHeaders(map[string]string{
			"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
			"User-Agent":   provider.config.userAgent,
		}).
		SetFormData(map[string]string{
			"getComentarios": strconv.Itoa(subtitle.Id),
		}).
		SetDebug(provider.config.debug).
		Post(provider.config.url + provider.config.searchUrl)
	if err != nil {
		provider.logger.Err(err).Msgf("error while getting comments for subtitle: %s", subtitle.Title)
	}

	err = json.Unmarshal(res.Body(), &result)
	if err != nil {
		provider.logger.Err(err).Msgf("error while getting comments for title: %s", subtitle.Title)
	}

	//var comments []models.SubComments
	var group []string
	var quality []string
	var resolution []string
	var duration []string
	stripTags := bluemonday.StripTagsPolicy()
	reg := regexp.MustCompile(`\n|\r\n`)
	for _, comment := range result.Data {
		if comment.Comment != "" {
			desc := reg.ReplaceAllString(stripTags.Sanitize(comment.Comment), " ")
			//nick := reg.ReplaceAllString(stripTags.Sanitize(comment.Nick), " ")
			group, quality, resolution, duration = parseExtra(desc)
			/* comments = append(comments, models.SubComments{
				Id:      comment.Id,
				Comment: desc,
				Nick:    nick,
				Date:    comment.Date,
			}) */
		}
	}

	subtitle.Group = append(subtitle.Group, group...)
	subtitle.Quality = append(subtitle.Quality, quality...)
	subtitle.Resolution = append(subtitle.Resolution, resolution...)
	subtitle.Duration = append(subtitle.Duration, duration...)

	//subtitle.Comments = comments
	subChan <- *subtitle
}

func downloadDivxSubtitle(provider *ProviderParams, subtitleId string) (io.ReadCloser, string, string, error) {
	res, err := provider.r.R().
		SetContext(provider.ctx).
		SetHeaders(map[string]string{
			"User-Agent": provider.config.userAgent,
			"Referer":    provider.config.url + "descargar.php",
			"Connection": "keep-alive",
		}).
		SetDebug(provider.config.debug).
		SetDoNotParseResponse(true).
		SetQueryParam("id", subtitleId).
		Get(provider.config.url + "descargar.php")
	if err != nil {
		return nil, "", "", err
	}

	contentType := res.Header().Get("Content-Type")
	ext := strings.Split(contentType, "/")[1]
	filename := fmt.Sprintf("%s.%s", subtitleId, ext)
	provider.logger.Info().Msgf("downloading file: %s", filename)
	return res.RawBody(), filename, contentType, nil
}

func parseTitle(text string) (itemType string, season int, episode int){
	s := Parse(text, "season")
	if s != nil {
		re := regexp.MustCompile("[^0-9]+")
		seasonStr := re.ReplaceAllString(s[0], "")
		season, _ = strconv.Atoi(seasonStr)
		itemType = "serie"
	}
	e := Parse(text, "episode")
	if e != nil{
		re := regexp.MustCompile("[^0-9]+")
		str := re.ReplaceAllString(e[0], "")
		episode, _ = strconv.Atoi(str)
		itemType = "serie"
	}
	if itemType == ""{
		itemType = "movie"
	}
	return
}
func parseExtra(text string) ([]string, []string, []string, []string) {
	var group []string
	var quality []string
	var resolution []string
	var duration []string
	
	g := Parse(text, "group")
	if g != nil {
		group = append(group, g...)
	}
	q := Parse(text, "quality")
	if q != nil {
		quality = append(quality, q...)
	}
	res := Parse(text, "resolution")
	if res != nil {
		resolution = append(resolution, res...)
	}
	d := Parse(text, "duration")
	if d != nil {
		duration = append(duration, d...)
	}
	return group, quality, resolution, duration
}
