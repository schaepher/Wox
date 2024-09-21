package system

import (
	"context"
	"wox/plugin"
	"wox/util/clipboard"

	"github.com/google/uuid"
)

var toolkitIcon = plugin.NewWoxImageSvg(`<svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g><g id="SVGRepo_iconCarrier"> <path d="M6 6L10.5 10.5M6 6H3L2 3L3 2L6 3V6ZM19.259 2.74101L16.6314 5.36863C16.2354 5.76465 16.0373 5.96265 15.9632 6.19098C15.8979 6.39183 15.8979 6.60817 15.9632 6.80902C16.0373 7.03735 16.2354 7.23535 16.6314 7.63137L16.8686 7.86863C17.2646 8.26465 17.4627 8.46265 17.691 8.53684C17.8918 8.6021 18.1082 8.6021 18.309 8.53684C18.5373 8.46265 18.7354 8.26465 19.1314 7.86863L21.5893 5.41072C21.854 6.05488 22 6.76039 22 7.5C22 10.5376 19.5376 13 16.5 13C16.1338 13 15.7759 12.9642 15.4298 12.8959C14.9436 12.8001 14.7005 12.7521 14.5532 12.7668C14.3965 12.7824 14.3193 12.8059 14.1805 12.8802C14.0499 12.9501 13.919 13.081 13.657 13.343L6.5 20.5C5.67157 21.3284 4.32843 21.3284 3.5 20.5C2.67157 19.6716 2.67157 18.3284 3.5 17.5L10.657 10.343C10.919 10.081 11.0499 9.95005 11.1198 9.81949C11.1941 9.68068 11.2176 9.60347 11.2332 9.44681C11.2479 9.29945 11.1999 9.05638 11.1041 8.57024C11.0358 8.22406 11 7.86621 11 7.5C11 4.46243 13.4624 2 16.5 2C17.5055 2 18.448 2.26982 19.259 2.74101ZM12.0001 14.9999L17.5 20.4999C18.3284 21.3283 19.6716 21.3283 20.5 20.4999C21.3284 19.6715 21.3284 18.3283 20.5 17.4999L15.9753 12.9753C15.655 12.945 15.3427 12.8872 15.0408 12.8043C14.6517 12.6975 14.2249 12.7751 13.9397 13.0603L12.0001 14.9999Z" stroke="#000000" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"></path> </g></svg>`)

func init() {
	plugin.AllSystemPlugin = append(plugin.AllSystemPlugin, &toolkitPlugin{})
}

type toolkitPlugin struct {
	api plugin.API
}

func (c *toolkitPlugin) GetMetadata() plugin.Metadata {
	return plugin.Metadata{
		Id:            "f3715b46-c8c0-983a-16eb-f0e03c823411",
		Name:          "Toolkit",
		Author:        "Wox Launcher",
		Website:       "https://github.com/Wox-launcher/Wox",
		Version:       "1.0.0",
		MinWoxVersion: "2.0.0",
		Runtime:       "Go",
		Description:   "some simple commands",
		Icon:          toolkitIcon.String(),
		Entry:         "",
		TriggerKeywords: []string{
			"tk",
		},
		Commands: []plugin.MetadataCommand{
			{
				Command:     "uuid",
				Description: "create a UUID",
			},
		},
		Features: []plugin.MetadataFeature{},
		SupportedOS: []string{
			"Windows",
			"Macos",
			"Linux",
		},
	}
}

func (c *toolkitPlugin) Init(ctx context.Context, initParams plugin.InitParams) {
	c.api = initParams.API
}

func (c *toolkitPlugin) Query(ctx context.Context, query plugin.Query) (results []plugin.QueryResult) {
	switch query.Command {
	case "uuid":
		return c.uuid()
	default:
	}

	return results
}

func (c *toolkitPlugin) uuid() (results []plugin.QueryResult) {
	id := uuid.NewString()
	results = append(results, plugin.QueryResult{
		Title:    "UUID",
		SubTitle: id,
		Score:    100,
		Icon:     toolkitIcon,
		Actions: []plugin.QueryResultAction{
			{
				Name: "Copy to clipboard",
				Action: func(ctx context.Context, actionContext plugin.ActionContext) {
					clipboard.WriteText(id)
				},
			},
		},
	})

	return results
}
