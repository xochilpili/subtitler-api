package providers

import (
	"context"
	"io"
	"strings"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"github.com/xochilpili/subtitler-api/internal/config"
	"github.com/xochilpili/subtitler-api/internal/models"
)

type ProviderConfig struct {
	url       string
	searchUrl string
	userAgent string
	debug     bool
	apiKey    string
}

type ProviderParams struct {
	config *ProviderConfig
	logger *zerolog.Logger
	r      *resty.Client
	ctx    context.Context
}

type Search func(provider *ProviderParams, query string) []models.Subtitle
type Download func(provider *ProviderParams, subtitleId string) (io.ReadCloser, string, string, error)
type Handler struct {
	config   *ProviderConfig
	Search   Search
	Download Download
}

type Manager struct {
	config   *config.Config
	logger   *zerolog.Logger
	r        *resty.Client
	handlers map[string]Handler
}

func New(config *config.Config, logger *zerolog.Logger) *Manager {
	r := resty.New()
	handlers := map[string]Handler{
		"subdivx": {
			config: &ProviderConfig{
				url:       "https://subdivx.com/",
				searchUrl: "inc/ajax.php",
				userAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
				debug:     config.Debug,
				apiKey:    "",
			},
			Search:   searchDivx,
			Download: downloadDivxSubtitle,
		},
		"opensubtitles": {
			config: &ProviderConfig{
				url:       "https://api.opensubtitles.com/",
				searchUrl: "api/v1/subtitles",
				userAgent: "subtitlerApi v1.0.0",
				debug:     config.Debug,
				apiKey:    strings.TrimSpace(config.OpenSubtitlesApiKey),
			},
			Search:   searchOpenSubtitles,
			Download: downloadOpenSubtitle,
		},
	}
	return &Manager{
		config:   config,
		logger:   logger,
		r:        r,
		handlers: handlers,
	}
}

func (m *Manager) Search(ctx context.Context, query string, postFilter *models.PostFilters) []models.Subtitle {
	items := m.search(ctx, query)
	filtered := m.postFiltering(postFilter, items)
	return filtered
}

func (m *Manager) Download(ctx context.Context, subtitleId string) (io.ReadCloser, string, string, error) {
	return m.handlers["subdivx"].Download(&ProviderParams{
		config: m.handlers["subdivx"].config,
		logger: m.logger,
		r:      m.r,
		ctx:    ctx,
	}, subtitleId)
}

func (m *Manager) search(ctx context.Context, query string) []models.Subtitle {
	wg := &sync.WaitGroup{}
	var subtitles []models.Subtitle
	subChan := make(chan []models.Subtitle)

	for p := range m.handlers {
		wg.Add(1)
		go func(ctx context.Context, provider string, query string, subChan chan<- []models.Subtitle, wg *sync.WaitGroup) {
			defer wg.Done()
			m.logger.Info().Msgf("Searching subtitles for provider: %s", provider)
			items := m.handlers[provider].
				Search(&ProviderParams{
					config: m.handlers[provider].config,
					logger: m.logger,
					r:      m.r,
					ctx:    ctx,
				}, query)
			subChan <- items
		}(ctx, p, query, subChan, wg)
	}

	go func() {
		wg.Wait()
		close(subChan)
	}()

	for item := range subChan {
		subtitles = append(subtitles, item...)
	}
	return subtitles
}

func (m *Manager) postFiltering(filters *models.PostFilters, subtitles []models.Subtitle) []models.Subtitle {
	var filtered []models.Subtitle
	for _, item := range subtitles {
		if filters.Year > 0 && item.Year != filters.Year {
			continue
		}
		if filters.Group != "" && !m.contains(filters.Group, item.Group) {

			continue
		}
		if filters.Quality != "" && !m.contains(filters.Quality, item.Quality) {
			continue
		}
		if filters.Resolution != "" && !m.contains(filters.Resolution, item.Resolution) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func (m *Manager) contains(term string, terms []string) bool {
	for _, item := range terms {
		if strings.Contains(term, item) {
			return true
		}
	}
	return false
}
