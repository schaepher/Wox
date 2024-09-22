package system

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"wox/plugin"
	"wox/share"
	"wox/util"
	"wox/util/clipboard"
	"wox/util/keyboard"
)

var snippetsIcon = plugin.NewWoxImageSvg(`<svg version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 128 128" enable-background="new 0 0 128 128" xml:space="preserve" fill="#000000"><g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g><g id="SVGRepo_iconCarrier"> <g> <g> <path fill-rule="evenodd" clip-rule="evenodd" fill="#546E7A" d="M88,0H16C7.164,0,0,7.164,0,16v72c0,8.836,7.164,16,16,16h33.168 l8-4l-8.004-4H16c-4.412,0-8-3.586-8-8V16c0-4.414,3.588-8,8-8h72c4.414,0,8,3.586,8,8v64.578l0.32-0.156 c0.914-5.031,3.754-9.367,7.68-12.336V16C104,7.164,96.836,0,88,0z"></path> </g> </g> <path fill-rule="evenodd" clip-rule="evenodd" fill="#F44336" d="M116,96c-6.629,0-12-5.375-12-12c0-6.629,5.371-12,12-12 c6.625,0,12,5.371,12,12C128,90.625,122.625,96,116,96z M116,80c-2.211,0-4,1.789-4,4c0,2.207,1.789,4,4,4s4-1.793,4-4 C120,81.789,118.211,80,116,80z"></path> <path fill-rule="evenodd" clip-rule="evenodd" fill="#F44336" d="M116,128c-6.629,0-12-5.375-12-12c0-6.629,5.371-12,12-12 c6.625,0,12,5.371,12,12C128,122.625,122.625,128,116,128z M116,112c-2.211,0-4,1.789-4,4s1.789,4,4,4s4-1.789,4-4 S118.211,112,116,112z"></path> <path fill="#B0BEC5" d="M24,32h56v-8H24V32z"></path> <path fill="#B0BEC5" d="M24,48h56v-8H24V48z"></path> <path fill="#B0BEC5" d="M24,64h56v-8H24V64z"></path> <path fill="#B0BEC5" d="M24,80h16v-8H24V80z"></path> <path fill-rule="evenodd" clip-rule="evenodd" fill="#B0BEC5" d="M100.371,96.285L92.945,100l7.426,3.715 c-1.656,2.082-2.918,4.48-3.637,7.117l-12.73-6.363l-30.215,15.109c-0.574,0.289-1.184,0.422-1.785,0.422 c-1.469,0-2.879-0.813-3.582-2.211c-0.988-1.977-0.188-4.375,1.789-5.367L75.059,100L50.211,87.578 c-1.977-0.992-2.777-3.391-1.789-5.367s3.391-2.781,5.367-1.789L84,95.531l12.73-6.367C97.445,91.805,98.711,94.199,100.371,96.285z "></path> </g></svg>`)

const snippetsSettings = "snippets"

func init() {
	plugin.AllSystemPlugin = append(plugin.AllSystemPlugin, &snippetsPlugin{})
}

type snippet struct {
	ID   string
	Name string
	Data string
}

type snippetsPlugin struct {
	api      plugin.API
	snippets map[string]snippet
}

func (c *snippetsPlugin) GetMetadata() plugin.Metadata {
	return plugin.Metadata{
		Id:            "ee767db8-e2ca-48d2-94fc-f520879de380",
		Name:          "Snippets",
		Author:        "Wox Launcher",
		Website:       "https://github.com/Wox-launcher/Wox",
		Version:       "1.0.0",
		MinWoxVersion: "2.0.0",
		Runtime:       "Go",
		Description:   "manage snippets",
		Icon:          snippetsIcon.String(),
		Entry:         "",
		TriggerKeywords: []string{
			"snippets",
		},
		Commands: []plugin.MetadataCommand{
			{
				Command:     "add",
				Description: "add snippets",
			},
			{
				Command:     "rename",
				Description: "rename the title of a snippet",
			},
		},
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

func (c *snippetsPlugin) Init(ctx context.Context, initParams plugin.InitParams) {
	c.api = initParams.API
	c.snippets = make(map[string]snippet)
	c.loadSnippets(ctx)
}

func (c *snippetsPlugin) Query(ctx context.Context, query plugin.Query) (results []plugin.QueryResult) {
	if query.Type == plugin.QueryTypeSelection {
		results = c.handleSelection(ctx, query)
	} else if query.Command != "" {
		results = c.handleCommand(ctx, query)
	} else {
		results = c.search(ctx, query.Search)
	}

	return results
}

func (c *snippetsPlugin) getTriggerKeyword() string {
	return c.GetMetadata().TriggerKeywords[0]
}

func (c *snippetsPlugin) search(ctx context.Context, search string) (results []plugin.QueryResult) {
	primaryActionCode := c.api.GetSetting(ctx, primaryActionSettingKey)

	search = strings.ToLower(search)

	for _, snippet := range c.snippets {
		var (
			dataContains bool
			nameContains bool
		)

		nameContains = strings.Contains(strings.ToLower(snippet.Name), search)
		if !nameContains {
			dataContains = strings.Contains(strings.ToLower(snippet.Data), search)
		}

		if dataContains || nameContains {
			results = append(results, plugin.QueryResult{
				Title:    snippet.Name,
				SubTitle: fmt.Sprintf("ID: %s", snippet.ID),
				Score:    100,
				Icon:     snippetsIcon,
				Preview: plugin.WoxPreview{
					PreviewType: plugin.WoxPreviewTypeText,
					PreviewData: snippet.Data,
				},
				Actions: []plugin.QueryResultAction{
					{
						Name:      "Copy to clipboard",
						IsDefault: primaryActionValueCopy == primaryActionCode,
						Action: func(ctx context.Context, actionContext plugin.ActionContext) {
							clipboard.WriteText(snippet.Data)
						},
					},
					{
						Name:      "Paste to active app",
						IsDefault: primaryActionValuePaste == primaryActionCode,
						Action: func(ctx context.Context, actionContext plugin.ActionContext) {
							clipboard.WriteText(snippet.Data)
							util.Go(context.Background(), "snippet copy", func() {
								time.Sleep(time.Millisecond * 100)
								err := keyboard.SimulatePaste()
								if err != nil {
									c.api.Log(ctx, plugin.LogLevelError, fmt.Sprintf("simulate paste clipboard failed, err=%s", err.Error()))
								} else {
									c.api.Log(ctx, plugin.LogLevelInfo, "simulate paste clipboard success")
								}
							})
						},
					},
					{
						Name:                   "Rename",
						PreventHideAfterAction: true,
						Action: func(ctx context.Context, actionContext plugin.ActionContext) {
							c.api.ChangeQuery(ctx, share.PlainQuery{
								QueryType: plugin.QueryTypeInput,
								QueryText: fmt.Sprintf("%s rename %s ", c.getTriggerKeyword(), snippet.ID),
							})
						},
					},
					{
						Name:                   "Delete",
						PreventHideAfterAction: true,
						Action: func(ctx context.Context, actionContext plugin.ActionContext) {
							delete(c.snippets, snippet.ID)
							util.Go(ctx, "delete snippet", func() {
								c.saveSnippets(ctx)
							})

							c.api.ChangeQuery(ctx, share.PlainQuery{
								QueryType: plugin.QueryTypeInput,
								QueryText: fmt.Sprintf("%s ", c.getTriggerKeyword()),
							})
						},
					},
				},
			})
		}
	}
	return results
}

func (c *snippetsPlugin) handleCommand(ctx context.Context, query plugin.Query) (results []plugin.QueryResult) {
	switch query.Command {
	case "add":
		results = c.handleAdd(ctx, query.Search)
	case "rename":
		results = c.handleRename(ctx, query)
	default:
	}

	return results
}

func (c *snippetsPlugin) handleAdd(ctx context.Context, data string) (results []plugin.QueryResult) {
	results = append(results, plugin.QueryResult{
		Title: "Add to snippets",
		Score: 100,
		Icon:  snippetsIcon,
		Preview: plugin.WoxPreview{
			PreviewType: plugin.WoxPreviewTypeText,
			PreviewData: data,
		},
		Actions: []plugin.QueryResultAction{
			{
				Name:                   "add to snippets",
				PreventHideAfterAction: true,
				Action: func(ctx context.Context, actionContext plugin.ActionContext) {
					c.addSnippet(ctx, data)
				},
			},
		},
	})

	return results
}

func (c *snippetsPlugin) addSnippet(ctx context.Context, data string) (results []plugin.QueryResult) {
	util.Go(ctx, "add snippet", func() {
		sum := md5.Sum([]byte(data))
		sumHex := hex.EncodeToString(sum[:])

		snippet := snippet{
			ID:   sumHex,
			Name: sumHex,
			Data: data,
		}

		c.snippets[snippet.ID] = snippet

		clipboard.WriteText(snippet.ID)
		c.api.ChangeQuery(ctx, share.PlainQuery{
			QueryType: plugin.QueryTypeInput,
			QueryText: fmt.Sprintf("%s rename %s ", c.GetMetadata().TriggerKeywords[0], snippet.ID),
		})

		c.saveSnippets(ctx)
	})
	return
}

func (c *snippetsPlugin) handleRename(ctx context.Context, query plugin.Query) (results []plugin.QueryResult) {
	q := strings.SplitN(query.Search, " ", 2)
	if len(q) < 2 {
		return
	}

	snippetID := q[0]
	newName := q[1]

	snippet, ok := c.snippets[snippetID]
	if !ok {
		return
	}

	results = append(results, plugin.QueryResult{
		Title:    "Rename snippet",
		SubTitle: fmt.Sprintf(`Rename "%s" to "%s"`, snippet.Name, newName),
		Score:    100,
		Icon:     snippetsIcon,
		Actions: []plugin.QueryResultAction{
			{
				Name:                   "Rename snipppet",
				PreventHideAfterAction: true,
				Action: func(_ context.Context, actionContext plugin.ActionContext) {
					snippet.Name = newName
					c.snippets[snippetID] = snippet

					util.Go(ctx, "rename snippet", func() {
						c.saveSnippets(ctx)
					})

					c.api.ChangeQuery(ctx, share.PlainQuery{
						QueryType: plugin.QueryTypeInput,
						QueryText: fmt.Sprintf("%s %s", query.TriggerKeyword, snippet.Name),
					})
				},
			},
		},
	})

	return
}

func (c *snippetsPlugin) handleSelection(ctx context.Context, query plugin.Query) (results []plugin.QueryResult) {
	if query.Selection.IsEmpty() || query.Selection.Type != util.SelectionTypeText {
		return results
	}

	return c.handleAdd(ctx, query.Selection.Text)
}

func (c *snippetsPlugin) loadSnippets(ctx context.Context) {
	snippets := c.api.GetSetting(ctx, snippetsSettings)
	if snippets == "" {
		c.snippets = make(map[string]snippet)
		return
	}

	var sni map[string]snippet
	unmarshalErr := json.Unmarshal([]byte(snippets), &sni)
	if unmarshalErr != nil {
		c.api.Log(ctx, plugin.LogLevelError, fmt.Sprintf("unmarshal snippets text failed, err=%s", unmarshalErr.Error()))
		return
	}

	c.snippets = sni
}

func (c *snippetsPlugin) saveSnippets(ctx context.Context) {
	if len(c.snippets) == 0 {
		return
	}

	s, err := json.Marshal(c.snippets)
	if err != nil {
		c.api.Log(ctx, plugin.LogLevelError, fmt.Sprintf("marshal snippets text failed, err=%s", err.Error()))
		return
	}

	c.api.SaveSetting(ctx, snippetsSettings, string(s), false)
}
