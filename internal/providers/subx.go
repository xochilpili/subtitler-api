package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/xochilpili/subtitler-api/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func searchSubX(provider *ProviderParams, query string) []models.Subtitle {
	tracer := otel.Tracer("subx")
	ctx, span := tracer.Start(provider.ctx, "SubdX.Search")
	defer span.End()

	span.SetAttributes(attribute.String("query", query))

	provider.logger.Info().Msgf("searching subtitles for: %s", query)

	var result SubXResponse[SubXResponseItem]
	res, err := provider.r.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"title": query,
		}).
		SetHeaders(map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded",
			"User-Agent":    provider.config.userAgent,
			"Authorization": "Bearer " + provider.config.apiKey,
		}).
		SetDebug(provider.config.debug).
		Get(provider.config.url + "/" + provider.config.searchUrl)

	if err != nil {
		provider.logger.Err(err).Msgf("error while getting subtitles")
		return nil
	}

	if res.StatusCode() != 200 {
		provider.logger.Err(errors.New("opensubtitles non ok response")).Msgf("status response %d", res.StatusCode())
		return nil
	}

	err = json.Unmarshal(res.Body(), &result)
	if err != nil {
		provider.logger.Err(err).Msgf("error while unmarshal subx json response: %v", err)
		return nil
	}

	subtitles := translate2ModelSubx(result.Items)
	provider.logger.Info().Msgf("returned %d subtitles", len(subtitles))
	return subtitles
}

func translate2ModelSubx(items []SubXResponseItem) []models.Subtitle {
	var subtitles []models.Subtitle
	reg := regexp.MustCompile(`\n|\r\n`)
	for i, item := range items {
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

		subtitle := models.Subtitle{
			Provider:    "subx",
			Type:        itemType,
			Id:          i,
			ExternalId:  item.Id,
			Title:       title,
			Description: desc,
			Language:    "es",
			Year:        year,
			Season:      season,
			Episode:     episode,
		}

		subtitle.Group = group
		subtitle.Quality = quality
		subtitle.Resolution = resolution
		subtitle.Duration = duration
		subtitles = append(subtitles, subtitle)
	}
	return subtitles
}

func downloadSubX(provider *ProviderParams, subtitleId string) (io.ReadCloser, string, string, error) {
	tracer := otel.Tracer(provider.config.url + provider.config.searchUrl)
	ctx, rootSpan := tracer.Start(provider.ctx, "Download Subtitle Flow")
	defer rootSpan.End()

	ctxDownload, spanDownloaded := tracer.Start(ctx, "Download Subtitle File")
	spanDownloaded.SetAttributes(attribute.String("subtitle_id", subtitleId))

	res, err := provider.r.R().
		SetHeaders(map[string]string{
			"Content-Type":        "application/x-www-form-urlencoded",
			"content-disposition": "attachment",
			"Authorization":       "Bearer " + provider.config.apiKey,
		}).
		SetDoNotParseResponse(true).
		SetDebug(provider.config.debug).
		SetContext(ctxDownload).
		Get(provider.config.url + "/subtitles/" + subtitleId + "/download")

	if err != nil {
		spanDownloaded.RecordError(err)
		spanDownloaded.SetStatus(404, "file download error")
		return nil, "", "", err
	}

	contentType := res.Header().Get("Content-Type")
	ext := strings.Split(contentType, "/")[1]
	if ext == "text" {
		ext = "srt"
	}
	filename := fmt.Sprintf("%s.%s", subtitleId, ext)
	provider.logger.Info().Msgf("downloading file: %s", filename)
	rootSpan.AddEvent(fmt.Sprintf("downloaded file: %s, format: %s", filename, ext))
	spanDownloaded.End()

	return res.RawBody(), filename, contentType, nil
}
