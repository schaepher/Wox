package system

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"wox/plugin"
	"wox/util"

	"github.com/samber/lo"
)

var urlIcon = plugin.PluginUrlIcon

func init() {
	plugin.AllSystemPlugin = append(plugin.AllSystemPlugin, &UrlPlugin{})
}

type UrlHistory struct {
	Url   string
	Icon  plugin.WoxImage
	Title string
}

type UrlPlugin struct {
	api        plugin.API
	reg        *regexp.Regexp
	recentUrls []UrlHistory
}

func (r *UrlPlugin) GetMetadata() plugin.Metadata {
	return plugin.Metadata{
		Id:            "1af58721-6c97-4901-b291-620daf08d9c9",
		Name:          "Url",
		Author:        "Wox Launcher",
		Website:       "https://github.com/Wox-launcher/Wox",
		Version:       "1.0.0",
		MinWoxVersion: "2.0.0",
		Runtime:       "Go",
		Description:   "Open the typed/selected URL from Wox",
		Icon:          urlIcon.String(),
		Entry:         "",
		TriggerKeywords: []string{
			"*",
		},
		Commands: []plugin.MetadataCommand{},
		Features: []plugin.MetadataFeature{
			{
				Name: plugin.MetadataFeatureQuerySelection,
			},
		},
		SupportedOS: []string{
			"Windows",
			"Macos",
			"Linux",
		},
	}
}

func (r *UrlPlugin) Init(ctx context.Context, initParams plugin.InitParams) {
	r.api = initParams.API
	r.reg = r.getReg()
	r.recentUrls = r.loadRecentUrls(ctx)
}

func (r *UrlPlugin) loadRecentUrls(ctx context.Context) []UrlHistory {
	urlsJson := r.api.GetSetting(ctx, "recentUrls")
	if urlsJson == "" {
		return []UrlHistory{}
	}

	var urls []UrlHistory
	err := json.Unmarshal([]byte(urlsJson), &urls)
	if err != nil {
		r.api.Log(ctx, plugin.LogLevelError, fmt.Sprintf("load recent urls error: %s", err.Error()))
		return []UrlHistory{}
	}

	return urls
}

func (r *UrlPlugin) getReg() *regexp.Regexp {
	// based on https://gist.github.com/dperini/729294
	return regexp.MustCompile(`^(http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/)?[a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$`)
}

func (r *UrlPlugin) Query(ctx context.Context, query plugin.Query) (results []plugin.QueryResult) {
	search := query.Search
	if query.Type == plugin.QueryTypeSelection {
		search = query.Selection.String()
	}

	if len(search) >= 2 {
		existingUrlHistory := lo.Filter(r.recentUrls, func(item UrlHistory, index int) bool {
			return strings.Contains(strings.ToLower(item.Url), strings.ToLower(search))
		})

		for _, history := range existingUrlHistory {
			results = append(results, plugin.QueryResult{
				Title:    history.Url,
				SubTitle: history.Title,
				Score:    100,
				Icon:     history.Icon.Overlay(urlIcon, 0.4, 0.6, 0.6),
				Actions: []plugin.QueryResultAction{
					{
						Name: "i18n:plugin_url_open",
						Icon: plugin.OpenIcon,
						Action: func(ctx context.Context, actionContext plugin.ActionContext) {
							openErr := util.ShellOpen(history.Url)
							if openErr != nil {
								r.api.Log(ctx, "Error opening URL", openErr.Error())
							}
						},
					},
					{
						Name: "i18n:plugin_url_remove",
						Icon: plugin.TrashIcon,
						Action: func(ctx context.Context, actionContext plugin.ActionContext) {
							r.removeRecentUrl(ctx, history.Url)
						},
					},
				},
			})
		}
	}

	if len(r.reg.FindStringIndex(search)) > 0 {
		results = append(results, plugin.QueryResult{
			Title:    search,
			SubTitle: "i18n:plugin_url_open_in_browser",
			Score:    100,
			Icon:     urlIcon,
			Actions: []plugin.QueryResultAction{
				{
					Name: "i18n:plugin_url_open",
					Icon: urlIcon,
					Action: func(ctx context.Context, actionContext plugin.ActionContext) {
						url := search
						if !strings.HasPrefix(url, "http") {
							url = "https://" + url
						}
						openErr := util.ShellOpen(url)
						if openErr != nil {
							r.api.Log(ctx, "Error opening URL", openErr.Error())
						} else {
							util.Go(ctx, "saveRecentUrl", func() {
								r.saveRecentUrl(ctx, url)
							})
						}
					},
				},
			},
		})
	}
	return
}

func (r *UrlPlugin) saveRecentUrl(ctx context.Context, url string) {
	icon, err := getWebsiteIconWithCache(ctx, url)
	if err != nil {
		r.api.Log(ctx, plugin.LogLevelError, fmt.Sprintf("get url icon error: %s", err.Error()))
		icon = urlIcon
	}

	title := ""
	body, err := util.HttpGet(ctx, url)
	if err == nil {
		titleStart := strings.Index(string(body), "<title>")
		titleEnd := strings.Index(string(body), "</title>")
		if titleStart != -1 && titleEnd != -1 {
			title = string(body[titleStart+7 : titleEnd])
		}
	} else {
		r.api.Log(ctx, plugin.LogLevelError, fmt.Sprintf("get url title error: %s", err.Error()))
	}

	newHistory := UrlHistory{
		Url:   url,
		Icon:  icon,
		Title: title,
	}

	// remove duplicate urls
	r.recentUrls = lo.Filter(r.recentUrls, func(item UrlHistory, index int) bool {
		return item.Url != url
	})
	r.recentUrls = append([]UrlHistory{newHistory}, r.recentUrls...)

	urlsJson, err := json.Marshal(r.recentUrls)
	if err != nil {
		r.api.Log(ctx, plugin.LogLevelError, fmt.Sprintf("save url setting error: %s", err.Error()))
		return
	}

	r.api.SaveSetting(ctx, "recentUrls", string(urlsJson), false)
}

func (r *UrlPlugin) removeRecentUrl(ctx context.Context, url string) {
	r.recentUrls = lo.Filter(r.recentUrls, func(item UrlHistory, index int) bool {
		return item.Url != url
	})

	urlsJson, err := json.Marshal(r.recentUrls)
	if err != nil {
		r.api.Log(ctx, plugin.LogLevelError, fmt.Sprintf("save url setting error: %s", err.Error()))
		return
	}

	r.api.SaveSetting(ctx, "recentUrls", string(urlsJson), false)
}
