package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/xochilpili/subtitler-api/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func searchOpenSubtitles(provider *ProviderParams, query string) []models.Subtitle {
	tracer := otel.Tracer("opensubtitles") // Changed to provider url as app
	ctx, span := tracer.Start(provider.ctx, "OpenSubtitles.API.Search")
	defer span.End()

	span.SetAttributes(attribute.String("query", query))

	var target OpenSubtitlesResponse[OpenSubtitlesItem]

	res, err := provider.r.R().
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
			"Api-Key":      provider.config.apiKey,
			"User-Agent":   provider.config.userAgent,
		}).
		SetDebug(provider.config.debug).
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"type":          "movie",
			"query":         query,
			"languages":     "es,en",
			"ai_translated": "true",
		}).
		Get(provider.config.url + provider.config.searchUrl)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(499, "error while fetching opensubtitles subtitles")
		provider.logger.Err(err).Msgf("error while fetching opensubtitles: %v", err)
		return nil
	}

	if res.StatusCode() != 200 {
		provider.logger.Err(errors.New("opensubtitles non ok response")).Msgf("status response %d", res.StatusCode())
		return nil
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
		id := item.Attributes.Files[0].FileId
		desc := item.Attributes.Release
		itemType, season, episode = parseTitle(item.Attributes.FeatureDetails.Title)
		group, quality, resolution, duration = parseExtra(desc)
		subtitle := models.Subtitle{
			Provider:    "opensubtitles",
			Id:          id,
			Type:        itemType,
			Title:       item.Attributes.FeatureDetails.Title,
			Description: item.Attributes.Release,
			Language:    item.Attributes.Language,
			Group:       group,
			Quality:     quality,
			Resolution:  resolution,
			Duration:    duration,
			Year:        item.Attributes.FeatureDetails.Year,
			Season:      season,
			Episode:     episode,
		}
		subtitles = append(subtitles, subtitle)
	}
	return subtitles
}

func downloadOpenSubtitle(provider *ProviderParams, subtitleId string) (io.ReadCloser, string, string, error) {
	tracer := otel.Tracer(provider.config.url + provider.config.searchUrl)
	ctx, rootSpan := tracer.Start(provider.ctx, "Download Subtitle Flow")
	defer rootSpan.End()

	ctxLogin, loginSpan := tracer.Start(ctx, "Login")
	loginSpan.SetAttributes(attribute.String("endpoint", provider.config.url+"api/v1/login"))

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
		SetContext(ctxLogin).
		SetResult(&tokenResponse).
		SetDebug(provider.config.debug).
		Post(provider.config.url + "api/v1/login")

	if err != nil {
		loginSpan.RecordError(err)
		loginSpan.SetStatus(401, "login error")
		return nil, "", "", err
	}

	if tokenResponse.Token == "" {
		errs := errors.New("unable to get token")
		loginSpan.RecordError(errs)
		loginSpan.SetStatus(401, "unable to get token")
		return nil, "", "", errs
	}
	rootSpan.AddEvent("login request completed")
	loginSpan.End()

	var downloadResponse struct {
		Link string `json:"link"`
	}

	ctxDownloadApi, spanDownload := tracer.Start(ctx, "Request Download Link")
	spanDownload.SetAttributes(attribute.String("file_id", subtitleId))

	_, err = provider.r.R().
		SetHeaders(map[string]string{
			"Content-Type":  "application/json",
			"Api-Key":       provider.config.apiKey,
			"Authorization": "Bearer " + tokenResponse.Token,
			"User-Agent":    provider.config.userAgent,
		}).
		SetDebug(provider.config.debug).
		SetContext(ctxDownloadApi).
		SetResult(&downloadResponse).
		SetBody(struct {
			FileId string `json:"file_id"`
		}{
			FileId: subtitleId,
		}).
		Post(provider.config.url + "api/v1/download")

	if err != nil || downloadResponse.Link == "" {
		spanDownload.RecordError(err)
		spanDownload.SetStatus(404, "failed to get download link")
		return nil, "", "", err
	}
	rootSpan.AddEvent("download link received")
	spanDownload.End()

	ctxDownload, spanDownloaded := tracer.Start(ctx, "Download Subtitle File")
	spanDownloaded.SetAttributes(attribute.String("download_url", downloadResponse.Link))

	res, err := provider.r.R().
		SetDoNotParseResponse(true).
		SetDebug(provider.config.debug).
		SetContext(ctxDownload).
		Get(downloadResponse.Link)

	if err != nil {
		spanDownloaded.RecordError(err)
		spanDownloaded.SetStatus(404, "file download error")
		return nil, "", "", err
	}

	contentType := res.Header().Get("Content-Type")
	ext := strings.Split(contentType, "/")[0]
	if ext == "text" {
		ext = "srt"
	}
	filename := fmt.Sprintf("%s.%s", subtitleId, ext)
	provider.logger.Info().Msgf("downloading file: %s", filename)
	rootSpan.AddEvent(fmt.Sprintf("downloaded file: %s, format: %s", filename, ext))
	spanDownloaded.End()

	return res.RawBody(), filename, contentType, nil
}
