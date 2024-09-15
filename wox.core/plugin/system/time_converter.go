package system

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"wox/plugin"
	"wox/setting/definition"
	"wox/setting/validator"
	"wox/share"
	"wox/util/clipboard"

	"github.com/araddon/dateparse"
)

var timeConvertIcon = plugin.NewWoxImageSvg(`<svg viewBox="0 0 1024 1024" class="icon" version="1.1" xmlns="http://www.w3.org/2000/svg" fill="#000000"><g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g><g id="SVGRepo_iconCarrier"><path d="M597.678 480.76L390.797 333.998c-22.209-15.766-53-10.532-68.762 11.687l-2.04 2.871c-15.753 22.214-10.526 53 11.691 68.762l206.876 146.771c22.218 15.757 53 10.527 68.766-11.687l2.035-2.876c15.768-22.218 10.529-53.005-11.685-68.766z" fill="#F39A2B"></path><path d="M585.066 423.392l-2.871-2.034c-22.218-15.763-53.004-10.527-68.766 11.687L279.007 763.472c-15.762 22.214-10.527 53.005 11.69 68.763l2.871 2.04c22.218 15.762 53.004 10.53 68.762-11.688l234.423-330.428c15.767-22.22 10.531-53.001-11.687-68.767z" fill="#E5594F"></path><path d="M891.662 525.126c-0.363 50.106-8.104 91.767-27.502 142.522-13.232 34.625-44.231 82.177-70.529 111.108-62.993 69.31-152.478 113.292-240.772 121.615-100.773 9.501-189.621-17.478-271.287-78.551 7.65 5.723-7.536-6.408-7.061-6.009-4.562-3.821-8.967-7.82-13.369-11.824-8.803-8.003-17.105-16.535-25.225-25.224-18.148-19.432-26.188-30.526-41.439-54.866-27.11-43.264-40.704-80.283-51.007-132.536-4.015-20.354-5.395-39.803-5.586-66.233-0.531-73.33-114.29-73.381-113.758 0 1.607 222.487 154.098 420.146 370.093 475.715 216.482 55.697 449.039-49.258 553.91-245.54 37.754-70.664 56.715-150.224 57.293-230.179 0.526-73.379-113.231-73.328-113.761 0.002z" fill="#4A5699"></path><path d="M137.884 501.467c0.362-50.104 8.103-91.762 27.502-142.52 13.233-34.621 44.233-82.173 70.53-111.108 62.993-69.309 152.472-113.29 240.768-121.615 100.773-9.5 189.626 17.479 271.292 78.554-7.652-5.721 7.532 6.408 7.057 6.01 4.563 3.819 8.968 7.821 13.371 11.823 8.803 8 17.108 16.535 25.228 25.225 18.147 19.43 26.187 30.526 41.438 54.866 27.111 43.264 40.709 80.28 51.009 132.533 4.014 20.352 5.396 39.804 5.586 66.232 0.529 73.33 114.287 73.384 113.76 0-1.608-222.489-154.107-420.144-370.099-475.715-216.482-55.7-449.036 49.26-553.905 245.541-37.753 70.664-56.715 150.219-57.292 230.174-0.534 73.384 113.225 73.33 113.755 0z" fill="#C45FA0"></path></g></svg>`)

const timeConvertSettingKey = "timeConvert"

func init() {
	plugin.AllSystemPlugin = append(plugin.AllSystemPlugin, &timeConverterPlugin{})
}

var (
	regexTimestamp = regexp.MustCompile(`^\d{10}$|^\d{13}$`)
	regexDigits    = regexp.MustCompile(`^\d+$`)
)

const (
	timestampSecondLen     = 10
	timestampMillSecondLen = 13
)

type timezone struct {
	Location string
	Timezone string
}

type timeConverterPlugin struct {
	api       plugin.API
	timezones []timezone
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
			"time",
		},
		Commands: []plugin.MetadataCommand{
			{
				Command:     "now",
				Description: "",
			},
			{
				Command:     "today",
				Description: "",
			},
			{
				Command:     "yesterday",
				Description: "",
			},
			{
				Command:     "tomorrow",
				Description: "",
			},
			{
				Command:     "timezone",
				Description: "",
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
		SettingDefinitions: []definition.PluginSettingDefinitionItem{
			{
				Type: definition.PluginSettingDefinitionTypeTable,
				Value: &definition.PluginSettingValueTable{
					Key:           timeConvertSettingKey,
					Title:         "Timezones",
					SortColumnKey: "Keyword",
					SortOrder:     definition.PluginSettingValueTableSortOrderAsc,
					Columns: []definition.PluginSettingValueTableColumn{
						{
							Key:   "Location",
							Label: "Location",
							Type:  definition.PluginSettingValueTableColumnTypeText,
							Validators: []validator.PluginSettingValidator{
								{
									Type:  validator.PluginSettingValidatorTypeNotEmpty,
									Value: &validator.PluginSettingValidatorNotEmpty{},
								},
							},
							Width: 80,
						},
						{
							Key:   "Timezone",
							Label: "Timezone",
							Type:  definition.PluginSettingValueTableColumnTypeText,
							Validators: []validator.PluginSettingValidator{
								{
									Type:  validator.PluginSettingValidatorTypeNotEmpty,
									Value: &validator.PluginSettingValidatorNotEmpty{},
								},
							},
							Width: 80,
						},
					},
				},
			},
		},
	}
}

func (c *timeConverterPlugin) Init(ctx context.Context, initParams plugin.InitParams) {
	c.api = initParams.API
	c.timezones = c.loadTimezones(ctx)

	c.api.OnSettingChanged(ctx, func(key string, value string) {
		if key == timeConvertSettingKey {
			c.timezones = c.loadTimezones(ctx)
		}
	})
}

func (c *timeConverterPlugin) loadTimezones(ctx context.Context) (timezones []timezone) {
	timezoneJson := c.api.GetSetting(ctx, timeConvertSettingKey)
	if timezoneJson == "" {
		timezones = c.defaultSettings()
		return
	}

	unmarshalErr := json.Unmarshal([]byte(timezoneJson), &timezones)
	if unmarshalErr != nil {
		c.api.Log(ctx, plugin.LogLevelError, fmt.Sprintf("failed to unmarshal timezones: %s", unmarshalErr.Error()))
		return
	}

	return
}

func (c *timeConverterPlugin) defaultSettings() (timezones []timezone) {
	return []timezone{
		{Location: "China", Timezone: "Asia/Shanghai"},
		{Location: "Indonesia", Timezone: "Asia/Jakarta"},
		{Location: "Thailand", Timezone: "Asia/Bangkok"},
		{Location: "Vietnam", Timezone: "Asia/Ho_Chi_Minh"},
		{Location: "Brazil", Timezone: "America/Sao_Paulo"},
	}
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

	now := time.Now()
	switch query.Command {
	// default is now
	case "now":
		t = now
	case "today":
		t = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	case "yesterday":
		t = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.Local)
	case "tomorrow":
		t = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local)
	case "timezone":
		splitted := strings.SplitN(text, " ", 2)
		if len(splitted) < 2 || splitted[1] == "" {
			return c.returnTimezoneList()
		} else {
			tz, err := time.LoadLocation(splitted[0])
			if err != nil {
				return []plugin.QueryResult{}
			}
			// match string datetime
			t, err = dateparse.ParseIn(text, tz)
			if err != nil {
				return []plugin.QueryResult{}
			}
		}
	default:
		if text == "" {
			t = now
		} else if regexDigits.MatchString(text) {
			if !regexTimestamp.MatchString(text) {
				return []plugin.QueryResult{}
			}

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
	}

	timeDisplay := map[string]string{
		"Timestamp":     strconv.FormatInt(t.Unix(), 10),
		"TimestampMill": strconv.FormatInt(t.UnixMilli(), 10),
		"Datetime":      t.Format("2006-01-02 15:04:05"),
		"Month":         t.Format("Jan"),
		"Weekday":       t.Format("Mon"),
	}
	displayOrder := []string{"Timestamp", "TimestampMill", "Datetime", "RFC3339", "Month"}

	for _, timezone := range c.timezones {
		tz, err := time.LoadLocation(timezone.Timezone)
		if err != nil {
			continue
		}
		timeDisplay[timezone.Location] = t.In(tz).Format(time.RFC3339)
		displayOrder = append(displayOrder, timezone.Location)
	}

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

func (c *timeConverterPlugin) returnTimezoneList() (results []plugin.QueryResult) {
	for _, tz := range c.timezones {
		results = append(results, plugin.QueryResult{
			Title:    tz.Location,
			SubTitle: tz.Timezone,
			Score:    100,
			Icon:     timeConvertIcon,
			Actions: []plugin.QueryResultAction{
				{
					Name:                   "Use timezone " + tz.Timezone,
					PreventHideAfterAction: true,
					Action: func(ctx context.Context, actionContext plugin.ActionContext) {
						c.api.ChangeQuery(ctx, share.PlainQuery{
							QueryType: plugin.QueryTypeInput,
							QueryText: fmt.Sprintf("time timezone %s ", tz.Timezone),
						})
					},
				},
			},
		})
	}

	return results
}
