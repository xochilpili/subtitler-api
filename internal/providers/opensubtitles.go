package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xochilpili/subtitler-api/internal/models"
)

func searchOpenSubtitles(provider *ProviderParams, query string) []models.Subtitle {
	var target OpenSubtitlesResponse[OpenSubtitlesItem]

	res, err := provider.r.R().
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
			"Api-Key":      provider.config.apiKey,
			"User-Agent":   provider.config.userAgent,
		}).
		SetDebug(provider.config.debug).
		SetContext(provider.ctx).
		SetQueryParams(map[string]string{
			"type":          "movie",
			"query":         query,
			"languages":     "es,en",
			"ai_translated": "true",
		}).
		Get(provider.config.url + provider.config.searchUrl)

	if err != nil {
		provider.logger.Err(err).Msgf("error while fetching openapi subtitles: %v", err)
		return nil
	}
	
	if(res.StatusCode() != 200){
		provider.logger.Err(errors.New("opensubtitles nont ok response")).Msgf("status response %d", res.StatusCode())
		return nil;
	}

	err = json.Unmarshal(res.Body(), &target)
	if err != nil {
		fmt.Printf("%s", res.Body())
		provider.logger.Err(err).Msgf("error while unmarshal opensubtitles json response: %v", err)
		return nil
	}
	return translate2Model(target.Data)
}

func translate2Model(items []OpenSubtitlesItem) []models.Subtitle {
	var subtitles []models.Subtitle
	for _, item := range items {
		var group []string
		var quality []string
		var resolution []string
		var duration []string
		var itemType string
		var season int
		var episode int
		id, _ := strconv.Atoi(item.Id)
		desc := item.Attributes.Release
		itemType, season, episode = parseTitle(item.Attributes.FeatureDetails.Title)
		group, quality, resolution, duration = parseExtra(desc)
		subtitle := models.Subtitle{
			Provider:    "opensubtitles",
			Id:          id,
			Type: itemType,
			Title:       item.Attributes.FeatureDetails.Title,
			Description: item.Attributes.Release,
			Language:    item.Attributes.Language,
			Group:       group,
			Quality:     quality,
			Resolution:  resolution,
			Duration:    duration,
			Year:        item.Attributes.FeatureDetails.Year,
			Season: season,
			Episode: episode,
		}
		subtitles = append(subtitles, subtitle)
	}
	return subtitles
}

func downloadOpenSubtitle(provider *ProviderParams, subtitleId string) (io.ReadCloser, string, string, error) {
	var tokenResponse struct {
		Token string `json:"token"`
	}
	_, err := provider.r.R().SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"Api-Key":      provider.config.apiKey,
		"User-Agent":   provider.config.userAgent,
	}).
		SetBody(struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			Username: provider.config.apiUsername,
			Password: provider.config.apiPassword,
		}).
		SetResult(&tokenResponse).
		SetDebug(provider.config.debug).
		Post(provider.config.url + "api/v1/login")

	if err != nil {
		return nil, "", "", err
	}

	if tokenResponse.Token == "" {
		errs := errors.New("unable to get token")
		return nil, "", "", errs
	}
	var downloadResponse struct {
		Link string `json:"link"`
	}
	_, err = provider.r.R().
		SetHeaders(map[string]string{
			"Content-Type":  "application/json",
			"Api-Key":       provider.config.apiKey,
			"Authorization": "Bearer " + tokenResponse.Token,
			"User-Agent":    provider.config.userAgent,
		}).
		SetDebug(provider.config.debug).
		SetContext(provider.ctx).
		SetResult(&downloadResponse).
		SetBody(struct {
			FileId string `json:"file_id"`
		}{
			FileId: subtitleId,
		}).
		Post(provider.config.url + "api/v1/download")
	if err != nil {
		return nil, "", "", err
	}
	if downloadResponse.Link == "" {
		errs := errors.New("unable to get download link")
		return nil, "", "", errs
	}

	res, err := provider.r.R().
		SetDoNotParseResponse(true).
		SetDebug(provider.config.debug).
		Get(downloadResponse.Link)
	if err != nil {
		return nil, "", "", err
	}

	contentType := res.Header().Get("Content-Type")
	ext := strings.Split(contentType, "/")[0]
	if(ext == "text"){
		ext = "srt"
	}
	filename := fmt.Sprintf("%s.%s", subtitleId, ext)
	provider.logger.Info().Msgf("downloading file: %s", filename)
	return res.RawBody(), filename, contentType, nil

}
