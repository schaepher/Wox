package system

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"wox/plugin"
	"wox/util"

	"github.com/blevesearch/bleve"
	"github.com/mitchellh/go-homedir"
)

var browserBookmarkIcon = plugin.NewWoxImageBase64(`data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAACXBIWXMAAAsTAAALEwEAmpwYAAAC9UlEQVR4nO1bTW8SURSdahfWdf0l6u9whfoPdNtNo2Ye07jRtNbE2LU1UZgHiZLUiJpIqXVhWw2ygkZMdGNTrMaApYChHHOnSJpWyswwH8z03uQkBN5M7j1z77sfvFEUFhYWlj6iaTihSlxWdbxUJcpCAk5h5skv4PMVW6Br99+ro9uLqI5LCjCiOCHXkzij6lhy0mi3CDiArBbD+EDGaws4LXR8cMt4lwkgvJ9IYsw2AULippvGe0AAhYVm7+lnMarq+Bl0AoSOH2SLnad/3m3jPSFAAjfiOGedAB0XwkIA2WKZgKhEJCwEkC1MgFWJsgeAQ0DwHgDeBAVnAXAaFFwH4BAJt1JtbHyvo1mrmUJ9u4aVYjM8hdCDTMu08f9QqexYIqBamDQwlARoCeD1xz/IlxqmkPvUwKOl1qH7TB8goF26ilLuHvRXOUwlWgboM31Hv+1fOx2GUjiaaONrfgaVwjUsv03hbmqr51r6jdZUi5P4kp81rg08AcILkpkAuOMBc+ldxJdbphB70zIyR2g8YC69i8a2tSywUa4fbwK+hYkAYSMEbj8NUQiIAIEJkOwBEVdK4UzefClsFVQ6P+6UzrPP2lhbb/53HelAungeAvM2miG7zRN1kUeto8bMn3a4bL4dtgpqn98V9tpn8oRqdadnau2XXXgPkLwJRhwPAREgsAdIn2eCvyu17jRotdg0NjinNkvSoV+XORQzwZXOQJRSm9MZw5c0qFmYCa6uN3FnYe8pkSdQkeNUwUQ6+FIIiQCBCZDsARHfByLxI/BwsdWNY9rRaXhi9lrSITAjseYRoM2M7mvl7zYC6dCPhGARULZOwH0/CBDHPQREgMAESPaAyNAelRVewM5RWeHRYWlPkMRZywRoWYzSUXPflR8QqsRWJImTlgkgETqm/DbAAaiKXZlIYoxeOwns09exps3jlDKIaDGMqzoW/TbGBjIDvzTVFWBETeCiKpEWEptDYFwvbAodz42879RrcywsLEqY5S+u/BSNeloCCQAAAABJRU5ErkJggg==`)

func init() {
	plugin.AllSystemPlugin = append(plugin.AllSystemPlugin, &BrowserBookmarkPlugin{})
}

type Bookmark struct {
	Name      string
	NameLower string
	Url       string
}

type BrowserBookmarkPlugin struct {
	api       plugin.API
	bookmarks []Bookmark
	index     bleve.Index
}

func (c *BrowserBookmarkPlugin) GetMetadata() plugin.Metadata {
	return plugin.Metadata{
		Id:            "95d041d3-be7e-4b20-8517-88dda2db280b",
		Name:          "BrowserBookmark",
		Author:        "Wox Launcher",
		Website:       "https://github.com/Wox-launcher/Wox",
		Version:       "1.0.0",
		MinWoxVersion: "2.0.0",
		Runtime:       "Go",
		Description:   "Search browser bookmarks",
		Icon:          browserBookmarkIcon.String(),
		Entry:         "",
		TriggerKeywords: []string{
			"*",
		},
		Commands: []plugin.MetadataCommand{},
		SupportedOS: []string{
			"Windows",
			"Macos",
			"Linux",
		},
	}
}

func (c *BrowserBookmarkPlugin) Init(ctx context.Context, initParams plugin.InitParams) {
	c.api = initParams.API

	if util.IsMacOS() {
		profiles := []string{"Default", "Profile 1", "Profile 2", "Profile 3"}
		for _, profile := range profiles {
			chromeBookmarks := c.loadChromeBookmarkInMacos(ctx, profile)
			c.bookmarks = append(c.bookmarks, chromeBookmarks...)
		}
	}

	for idx, bookmark := range c.bookmarks {
		bookmark.NameLower = strings.ToLower(bookmark.Name)
		c.bookmarks[idx] = bookmark
	}

	indexMapping := bleve.NewIndexMapping()
	index, _ := bleve.NewMemOnly(indexMapping)
	c.index = index

	for idx, bookmark := range c.bookmarks {
		index.Index(fmt.Sprint(idx), bookmark)
	}
}

func (c *BrowserBookmarkPlugin) Query(ctx context.Context, query plugin.Query) (results []plugin.QueryResult) {
	results = make([]plugin.QueryResult, 0)

	q := bleve.NewMatchQuery(query.Search)
	searchRequest := bleve.NewSearchRequest(q)
	searchRequest.Size = 50
	searchResult, err := c.index.Search(searchRequest)
	if err != nil || searchResult == nil {
		return
	}

	fields := strings.Fields(query.Search)
	for index, field := range fields {
		fields[index] = strings.ToLower(field)
	}

	for _, hit := range searchResult.Hits {
		id, _ := strconv.Atoi(hit.ID)
		bookmark := c.bookmarks[id]

		containsAll := true
		for _, term := range fields {
			if !strings.Contains(bookmark.NameLower, term) {
				containsAll = false
				break
			}
		}
		if !containsAll {
			continue
		}

		results = append(results, plugin.QueryResult{
			Title:    bookmark.Name,
			SubTitle: bookmark.Url,
			Score:    int64(hit.Score * 100),
			Icon:     browserBookmarkIcon,
			Actions: []plugin.QueryResultAction{
				{
					Name: "i18n:plugin_browser_bookmark_open_in_browser",
					Action: func(ctx context.Context, actionContext plugin.ActionContext) {
						util.ShellOpen(bookmark.Url)
					},
				},
			},
		})
	}

	return
}

func (c *BrowserBookmarkPlugin) loadChromeBookmarkInMacos(ctx context.Context, profile string) (results []Bookmark) {
	bookmarkLocation, _ := homedir.Expand(fmt.Sprintf("~/Library/Application Support/Google/Chrome/%s/Bookmarks", profile))
	if _, err := os.Stat(bookmarkLocation); os.IsNotExist(err) {
		return
	}
	file, readErr := os.ReadFile(bookmarkLocation)
	if readErr != nil {
		c.api.Log(ctx, plugin.LogLevelError, fmt.Sprintf("error reading chrome bookmark file: %s", readErr.Error()))
		return
	}

	groups := util.FindRegexGroups(`(?ms)name": "(?P<name>.*?)",.*?type": "url",.*?"url": "(?P<url>.*?)".*?}, {`, string(file))
	for _, group := range groups {
		results = append(results, Bookmark{
			Name: group["name"],
			Url:  group["url"],
		})
	}

	return results
}
