package providers

import (
	"encoding/json"
	"io"
	"strconv"

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

	err = json.Unmarshal(res.Body(), &target)
	if err != nil {
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
		id, _ := strconv.Atoi(item.Id)
		desc := item.Attributes.Release
		group, quality, resolution, duration = parse(desc)
		subtitle := models.Subtitle{
			Provider:    "opensubtitles",
			Id:          id,
			Title:       item.Attributes.FeatureDetails.Title,
			Description: item.Attributes.Release,
			Language:    item.Attributes.Language,
			Group:       group,
			Quality:     quality,
			Resolution:  resolution,
			Duration:    duration,
			Year:        item.Attributes.FeatureDetails.Year,
		}
		subtitles = append(subtitles, subtitle)
	}
	return subtitles
}

func downloadOpenSubtitle(provider *ProviderParams, subtitleId string) (io.ReadCloser, string, string, error) {
	return nil, "", "", nil
}
