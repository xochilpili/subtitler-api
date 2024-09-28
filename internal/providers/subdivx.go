package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/xochilpili/subtitler-api/internal/models"
)

type Token struct {
	Cookie string `json:"cookie,omitempty"`
	Token  string `json:"token"`
}

func search(provider *ProviderParams, query string) []models.Subtitle {
	ctx, cancel := context.WithTimeout(provider.ctx, 30*time.Second)
	provider.ctx = ctx
	defer cancel()
	provider.logger.Info().Msgf("searching subtitles for: %s", query)
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

	data, err := getSubtitles(provider, params)
	if err != nil {
		provider.logger.Err(err).Msg("error while getting subtitles")
		return nil
	}

	return data
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
	headers := res.Header()
	cookies := headers["Set-Cookie"]
	token.Cookie = strings.Split(strings.Split(cookies[0], ";")[0], "=")[1]

	err = json.Unmarshal(res.Body(), &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func getSubtitles(provider *ProviderParams, params *SubdivxSubPayload) ([]models.Subtitle, error) {
	var result SubdivxResponse[SubData]
	resp, err := provider.r.R().
		SetContext(provider.ctx).
		SetFormData(map[string]string{
			"tabla":     params.Tabla,
			"filtros":   params.Filtros,
			"buscar393": params.Buscar,
			"token":     params.Token,
		}).
		SetHeaders(map[string]string{
			"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
			"User-Agent":   provider.config.userAgent,
		}).
		SetDebug(provider.config.debug).
		Post(provider.config.url + provider.config.searchUrl)
	if err != nil {
		provider.logger.Err(err).Msgf("error while getting subtitler")
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

		stripTags := bluemonday.StripTagsPolicy()
		title := reg.ReplaceAllString(stripTags.Sanitize(item.Title), " ")
		desc := reg.ReplaceAllString(stripTags.Sanitize(item.Description), " ")
		g := Parse(desc, "group")
		if g != "" {
			group = append(group, g)
		}
		q := Parse(desc, "quality")
		if q != "" {
			quality = append(quality, q)
		}
		res := Parse(desc, "resolution")
		if res != "" {
			resolution = append(resolution, res)
		}
		d := Parse(desc, "duration")
		if d != "" {
			duration = append(duration, d)
		}
		subtitle := &models.Subtitle{
			Id:          item.Id,
			Title:       title,
			Description: desc,
			//Cds:         item.Cds,
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
	provider.logger.Info().Msgf("returned %d subtitles for %s", len(subtitles), params.Buscar)
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
			g := Parse(desc, "group")
			if g != "" {
				group = append(group, g)
			}
			q := Parse(desc, "quality")
			if q != "" {
				quality = append(quality, q)
			}
			res := Parse(desc, "resolution")
			if res != "" {
				resolution = append(resolution, res)
			}
			d := Parse(desc, "duration")
			if d != "" {
				duration = append(duration, d)
			}

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

func downloadSubtitle(provider *ProviderParams, subtitleId string) (io.ReadCloser, string, string, error) {
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
