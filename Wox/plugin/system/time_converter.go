package system

import (
	"context"
	"regexp"
	"strconv"
	"time"
	"wox/plugin"
	"wox/util/clipboard"

	"github.com/araddon/dateparse"
)

var timeConvertIcon = plugin.NewWoxImageSvg(``)

func init() {
	plugin.AllSystemPlugin = append(plugin.AllSystemPlugin, &timeConverterPlugin{})
}

var regexTimestamp = regexp.MustCompile(`^\d{10}$|^\d{13}$`)

const (
	timestampSecondLen     = 10
	timestampMillSecondLen = 13
)

type timeConverterPlugin struct {
	api plugin.API
}

func (c *timeConverterPlugin) GetMetadata() plugin.Metadata {
	return plugin.Metadata{
		Id:            "002fcc7b-1557-fad3-4314-9e0e5d5ea7a0",
		Name:          "Time Converter",
		Author:        "Wox Launcher",
		Website:       "https://github.com/Wox-launcher/Wox",
		Version:       "1.0.0",
		MinWoxVersion: "2.0.0",
		Runtime:       "Go",
		Description:   "Convert datetime or timestamp to several formats",
		Icon:          timeConvertIcon.String(),
		Entry:         "",
		TriggerKeywords: []string{
			"*",
			"now",
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

func (c *timeConverterPlugin) Init(ctx context.Context, initParams plugin.InitParams) {
	c.api = initParams.API
}

func (c *timeConverterPlugin) Query(ctx context.Context, query plugin.Query) (results []plugin.QueryResult) {
	text := query.Search
	if query.Type == plugin.QueryTypeSelection {
		text = query.Selection.Text
	}

	var (
		t   time.Time
		err error
	)

	// default is now
	if text == "" || text == "now" {
		t = time.Now()
	} else if regexTimestamp.MatchString(text) {
		// match timestamp
		ts, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return []plugin.QueryResult{}
		}

		if len(text) == timestampSecondLen {
			t = time.Unix(ts, 0)
		} else if len(text) == timestampMillSecondLen {
			t = time.UnixMilli(ts)
		} else {
			return []plugin.QueryResult{}
		}
	} else {
		// match string datetime
		t, err = dateparse.ParseIn(text, time.Local)
		if err != nil {
			return []plugin.QueryResult{}
		}
	}

	timeDisplay := map[string]string{
		"Timestamp":     strconv.FormatInt(t.Unix(), 10),
		"TimestampMill": strconv.FormatInt(t.UnixMilli(), 10),
		"Datetime":      t.Format("2006-01-02 15:04:05"),
		"RFC3339":       t.Local().Format(time.RFC3339),
		"RFC3339-ID":    t.In(time.FixedZone("Indonesia", 7*60*60)).Format(time.RFC3339),
		"RFC3339-TH":    t.In(time.FixedZone("Thailand", 7*60*60)).Format(time.RFC3339),
		"RFC3339-VN":    t.In(time.FixedZone("Vietnam", 7*60*60)).Format(time.RFC3339),
		"RFC3339-BR":    t.In(time.FixedZone("Brazil", -3*60*60)).Format(time.RFC3339),
		"Month":         t.Format("Jan"),
	}

	displayOrder := []string{"Timestamp", "TimestampMill", "Datetime", "RFC3339", "RFC3339-ID", "RFC3339-TH", "RFC3339-VN", "RFC3339-BR", "Month"}

	score := int64(100)
	for _, title := range displayOrder {
		display := timeDisplay[title]

		results = append(results, plugin.QueryResult{
			Title:    title,
			SubTitle: display,
			Score:    score,
			Icon:     timeConvertIcon,
			Actions: []plugin.QueryResultAction{
				{
					Name: "Copy to clipboard",
					Action: func(ctx context.Context, actionContext plugin.ActionContext) {
						clipboard.WriteText(display)
					},
				},
			},
		})

		score--
	}

	return results
}
