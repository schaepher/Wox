package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"wox/plugin"
	"wox/setting"
	"wox/share"
	"wox/util"
	"wox/util/notifier"
	"wox/util/window"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"
)

type uiImpl struct {
	requestMap *util.HashMap[string, chan WebsocketMsg]
}

func (u *uiImpl) ChangeQuery(ctx context.Context, query share.PlainQuery) {
	u.invokeWebsocketMethod(ctx, "ChangeQuery", query)
}

func (u *uiImpl) HideApp(ctx context.Context) {
	u.invokeWebsocketMethod(ctx, "HideApp", nil)
}

func (u *uiImpl) ShowApp(ctx context.Context, showContext share.ShowContext) {
	GetUIManager().SetActiveWindowName(window.GetActiveWindowName())
	GetUIManager().SetActiveWindowPid(window.GetActiveWindowPid())
	u.invokeWebsocketMethod(ctx, "ShowApp", getShowAppParams(ctx, showContext.SelectAll))
}

func (u *uiImpl) ToggleApp(ctx context.Context) {
	GetUIManager().SetActiveWindowName(window.GetActiveWindowName())
	GetUIManager().SetActiveWindowPid(window.GetActiveWindowPid())
	u.invokeWebsocketMethod(ctx, "ToggleApp", getShowAppParams(ctx, true))
}

func (u *uiImpl) GetServerPort(ctx context.Context) int {
	return GetUIManager().serverPort
}

func (u *uiImpl) ChangeTheme(ctx context.Context, theme share.Theme) {
	logger.Info(ctx, fmt.Sprintf("change theme: %s", theme.ThemeName))
	woxSetting := setting.GetSettingManager().GetWoxSetting(ctx)
	woxSetting.ThemeId = theme.ThemeId
	setting.GetSettingManager().SaveWoxSetting(ctx)
	u.invokeWebsocketMethod(ctx, "ChangeTheme", theme)
}

func (u *uiImpl) InstallTheme(ctx context.Context, theme share.Theme) {
	logger.Info(ctx, fmt.Sprintf("install theme: %s", theme.ThemeName))
	GetStoreManager().Install(ctx, theme)
}

func (u *uiImpl) UninstallTheme(ctx context.Context, theme share.Theme) {
	logger.Info(ctx, fmt.Sprintf("uninstall theme: %s", theme.ThemeName))
	GetStoreManager().Uninstall(ctx, theme)
	GetUIManager().ChangeToDefaultTheme(ctx)
}

func (u *uiImpl) OpenSettingWindow(ctx context.Context, windowContext share.SettingWindowContext) {
	u.invokeWebsocketMethod(ctx, "OpenSettingWindow", windowContext)
}

func (u *uiImpl) GetAllThemes(ctx context.Context) []share.Theme {
	return GetUIManager().GetAllThemes(ctx)
}

func (u *uiImpl) RestoreTheme(ctx context.Context) {
	GetUIManager().RestoreTheme(ctx)
}

func (u *uiImpl) Notify(ctx context.Context, msg share.NotifyMsg) {
	if u.isNotifyInToolbar(ctx, msg.PluginId) {
		u.invokeWebsocketMethod(ctx, "ShowToolbarMsg", msg)
	} else {
		notifier.Notify(msg.Text)
	}
}

func (u *uiImpl) isNotifyInToolbar(ctx context.Context, pluginId string) bool {
	isVisible, err := u.invokeWebsocketMethod(ctx, "IsVisible", nil)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("isNotifyInToolbar isVisible error: %s", err.Error()))
		return false
	}
	if !isVisible.(bool) {
		return false
	}

	respData, err := u.invokeWebsocketMethod(ctx, "GetCurrentQuery", nil)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("isNotifyInToolbar error: %s", err.Error()))
		return false
	}
	//first marshal to json , then unmarshal to share.PlainQuery
	jsonData, marshalErr := json.Marshal(respData)
	if marshalErr != nil {
		logger.Error(ctx, fmt.Sprintf("isNotifyInToolbar marshal error: %s", marshalErr.Error()))
		return false
	}
	var currentQuery share.PlainQuery
	unmarshalErr := json.Unmarshal(jsonData, &currentQuery)
	if unmarshalErr != nil {
		logger.Error(ctx, fmt.Sprintf("isNotifyInToolbar unmarshal error: %s", unmarshalErr.Error()))
		return false
	}

	// if current query is plugin specific query, we show notify in toolbar
	if currentQuery.QueryType == plugin.QueryTypeSelection {
		return true
	}

	queryPlugin, pluginInstance, queryErr := plugin.GetPluginManager().NewQuery(ctx, currentQuery)
	if queryErr != nil {
		logger.Error(ctx, fmt.Sprintf("isNotifyInToolbar new query error: %s", queryErr.Error()))
		return false
	}
	if pluginInstance == nil {
		logger.Error(ctx, "isNotifyInToolbar plugin instance not found")
		return false
	}

	if !queryPlugin.IsGlobalQuery() && pluginInstance.Metadata.Id == pluginId {
		return true
	}

	return false
}

func (u *uiImpl) PickFiles(ctx context.Context, params share.PickFilesParams) []string {
	respData, err := u.invokeWebsocketMethod(ctx, "PickFiles", params)
	if err != nil {
		return nil
	}
	if _, ok := respData.([]any); !ok {
		logger.Error(ctx, fmt.Sprintf("pick files response data type error: %T", respData))
		return nil
	}

	var result []string
	lo.ForEach(respData.([]any), func(file any, _ int) {
		result = append(result, file.(string))
	})
	return result
}

func (u *uiImpl) GetActiveWindowName() string {
	return GetUIManager().GetActiveWindowName()
}

func (u *uiImpl) GetActiveWindowPid() int {
	return GetUIManager().GetActiveWindowPid()
}

func (u *uiImpl) invokeWebsocketMethod(ctx context.Context, method string, data any) (responseData any, responseErr error) {
	requestID := uuid.NewString()
	resultChan := make(chan WebsocketMsg)
	u.requestMap.Store(requestID, resultChan)
	defer u.requestMap.Delete(requestID)

	traceId := util.GetContextTraceId(ctx)

	err := requestUI(ctx, WebsocketMsg{
		RequestId: requestID,
		TraceId:   traceId,
		Method:    method,
		Data:      data,
	})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("send message to UI error: %s", err.Error()))
		return "", err
	}

	var timeout = time.Second * 2
	if method == "PickFiles" {
		// pick files may take a long time
		timeout = time.Second * 180
	}
	select {
	case <-time.NewTimer(timeout).C:
		logger.Error(ctx, fmt.Sprintf("invoke ui method %s response timeout", method))
		return "", fmt.Errorf("request timeout, request id: %s", requestID)
	case response := <-resultChan:
		if !response.Success {
			return response.Data, errors.New("ui method response error")
		} else {
			return response.Data, nil
		}
	}
}

func getShowAppParams(ctx context.Context, selectAll bool) map[string]any {
	return map[string]any{
		"SelectAll":      selectAll,
		"Position":       NewMouseScreenPosition(),
		"QueryHistories": setting.GetSettingManager().GetLatestQueryHistory(ctx, 10),
		"LastQueryMode":  setting.GetSettingManager().GetWoxSetting(ctx).LastQueryMode,
	}
}

func onUIWebsocketRequest(ctx context.Context, request WebsocketMsg) {
	if request.Method != "Log" {
		logger.Debug(ctx, fmt.Sprintf("got <%s> request from ui", request.Method))
	}

	// we handle time/amount sensitive requests in websocket, other requests in http (see router.go)
	switch request.Method {
	case "Log":
		handleWebsocketLog(ctx, request)
	case "Query":
		handleWebsocketQuery(ctx, request)
	case "Action":
		handleWebsocketAction(ctx, request)
	case "Refresh":
		handleWebsocketRefresh(ctx, request)
	}
}

func onUIWebsocketResponse(ctx context.Context, response WebsocketMsg) {
	logger.Debug(ctx, fmt.Sprintf("got <%s> response from ui", response.Method))

	requestID := response.RequestId
	if requestID == "" {
		logger.Error(ctx, "response id not found")
		return
	}

	resultChan, exist := GetUIManager().GetUI(ctx).(*uiImpl).requestMap.Load(requestID)
	if !exist {
		logger.Error(ctx, fmt.Sprintf("response id not found: %s", requestID))
		return
	}

	resultChan <- response
}

func handleWebsocketLog(ctx context.Context, request WebsocketMsg) {
	traceId, traceIdErr := getWebsocketMsgParameter(ctx, request, "traceId")
	if traceIdErr != nil {
		logger.Error(ctx, traceIdErr.Error())
		responseUIError(ctx, request, traceIdErr.Error())
		return
	}
	level, levelErr := getWebsocketMsgParameter(ctx, request, "level")
	if levelErr != nil {
		logger.Error(ctx, levelErr.Error())
		responseUIError(ctx, request, levelErr.Error())
		return
	}
	message, messageErr := getWebsocketMsgParameter(ctx, request, "message")
	if messageErr != nil {
		logger.Error(ctx, messageErr.Error())
		responseUIError(ctx, request, messageErr.Error())
		return
	}

	logCtx := util.NewComponentContext(util.NewTraceContextWith(traceId), " UI")

	switch level {
	case "debug":
		logger.Debug(logCtx, message)
	case "info":
		logger.Info(logCtx, message)
	case "warn":
		logger.Warn(logCtx, message)
	case "error":
		logger.Error(logCtx, message)
	default:
		logger.Error(ctx, fmt.Sprintf("unsupported log level: %s", level))
		responseUIError(ctx, request, fmt.Sprintf("unsupported log level: %s", level))
	}
	responseUISuccess(ctx, request)
}

func handleWebsocketQuery(ctx context.Context, request WebsocketMsg) {
	queryId, queryIdErr := getWebsocketMsgParameter(ctx, request, "queryId")
	if queryIdErr != nil {
		logger.Error(ctx, queryIdErr.Error())
		responseUIError(ctx, request, queryIdErr.Error())
		return
	}
	queryType, queryTypeErr := getWebsocketMsgParameter(ctx, request, "queryType")
	if queryTypeErr != nil {
		logger.Error(ctx, queryTypeErr.Error())
		responseUIError(ctx, request, queryTypeErr.Error())
		return
	}
	queryText, queryTextErr := getWebsocketMsgParameter(ctx, request, "queryText")
	if queryTextErr != nil {
		logger.Error(ctx, queryTextErr.Error())
		responseUIError(ctx, request, queryTextErr.Error())
		return
	}
	querySelectionJson, querySelectionErr := getWebsocketMsgParameter(ctx, request, "querySelection")
	if querySelectionErr != nil {
		logger.Error(ctx, querySelectionErr.Error())
		responseUIError(ctx, request, querySelectionErr.Error())
		return
	}
	var querySelection util.Selection
	json.Unmarshal([]byte(querySelectionJson), &querySelection)

	var changedQuery share.PlainQuery
	if queryType == plugin.QueryTypeInput {
		changedQuery = share.PlainQuery{
			QueryType: plugin.QueryTypeInput,
			QueryText: queryText,
		}
	} else if queryType == plugin.QueryTypeSelection {
		changedQuery = share.PlainQuery{
			QueryType:      plugin.QueryTypeSelection,
			QueryText:      queryText,
			QuerySelection: querySelection,
		}
	} else {
		logger.Error(ctx, fmt.Sprintf("unsupported query type: %s", queryType))
		responseUIError(ctx, request, fmt.Sprintf("unsupported query type: %s", queryType))
		return
	}

	logger.Info(ctx, fmt.Sprintf("start to handle query changed: %s, queryId: %s", changedQuery.String(), queryId))

	if changedQuery.QueryType == plugin.QueryTypeInput && changedQuery.QueryText == "" {
		responseUISuccessWithData(ctx, request, []string{})
		return
	}
	if changedQuery.QueryType == plugin.QueryTypeSelection && changedQuery.QuerySelection.String() == "" {
		responseUISuccessWithData(ctx, request, []string{})
		return
	}

	query, queryPlugin, queryErr := plugin.GetPluginManager().NewQuery(ctx, changedQuery)
	if queryErr != nil {
		logger.Error(ctx, queryErr.Error())
		responseUIError(ctx, request, queryErr.Error())
		return
	}

	var totalResultCount int
	var startTimestamp = util.GetSystemTimestamp()
	var resultDebouncer = util.NewDebouncer(24, func(results []plugin.QueryResultUI, reason string) {
		logger.Info(ctx, fmt.Sprintf("query %s: %s, result flushed (reason: %s), total results: %d", query.Type, query.String(), reason, totalResultCount))
		responseUISuccessWithData(ctx, request, results)
	})
	resultDebouncer.Start(ctx)
	logger.Info(ctx, fmt.Sprintf("query %s: %s, result flushed (new start)", query.Type, query.String()))
	resultChan, doneChan := plugin.GetPluginManager().Query(ctx, query)
	for {
		select {
		case results := <-resultChan:
			if len(results) == 0 {
				continue
			}
			lo.ForEach(results, func(_ plugin.QueryResultUI, index int) {
				results[index].QueryId = queryId
			})
			totalResultCount += len(results)
			resultDebouncer.Add(ctx, results)
		case <-doneChan:
			logger.Info(ctx, fmt.Sprintf("query done, total results: %d, cost %d ms", totalResultCount, util.GetSystemTimestamp()-startTimestamp))

			// if there is no result, show fallback search
			if totalResultCount == 0 {
				fallbackResults := plugin.GetPluginManager().QueryFallback(ctx, query, queryPlugin)
				if len(fallbackResults) > 0 {
					lo.ForEach(fallbackResults, func(_ plugin.QueryResultUI, index int) {
						fallbackResults[index].QueryId = queryId
					})
					resultDebouncer.Add(ctx, fallbackResults)
					logger.Info(ctx, fmt.Sprintf("no result, show %d fallback results", len(fallbackResults)))
				} else {
					logger.Info(ctx, "no result, no fallback results")
				}
			}

			resultDebouncer.Done(ctx)
			return
		case <-time.After(time.Minute):
			logger.Info(ctx, fmt.Sprintf("query timeout, query: %s, request id: %s", query.String(), request.RequestId))
			resultDebouncer.Done(ctx)
			responseUIError(ctx, request, fmt.Sprintf("query timeout, query: %s, request id: %s", query.String(), request.RequestId))
			return
		}
	}

}

func handleWebsocketAction(ctx context.Context, request WebsocketMsg) {
	resultId, idErr := getWebsocketMsgParameter(ctx, request, "resultId")
	if idErr != nil {
		logger.Error(ctx, idErr.Error())
		responseUIError(ctx, request, idErr.Error())
		return
	}
	actionId, actionIdErr := getWebsocketMsgParameter(ctx, request, "actionId")
	if actionIdErr != nil {
		logger.Error(ctx, actionIdErr.Error())
		responseUIError(ctx, request, actionIdErr.Error())
		return
	}

	executeErr := plugin.GetPluginManager().ExecuteAction(ctx, resultId, actionId)
	if executeErr != nil {
		responseUIError(ctx, request, executeErr.Error())
		return
	}

	responseUISuccess(ctx, request)
}

func handleWebsocketRefresh(ctx context.Context, request WebsocketMsg) {
	resultStr, resultErr := getWebsocketMsgParameter(ctx, request, "refreshableResult")
	if resultErr != nil {
		logger.Error(ctx, resultErr.Error())
		responseUIError(ctx, request, resultErr.Error())
		return
	}

	queryId, queryIdErr := getWebsocketMsgParameter(ctx, request, "queryId")
	if queryIdErr != nil {
		logger.Error(ctx, queryIdErr.Error())
		responseUIError(ctx, request, queryIdErr.Error())
		return
	}

	var result plugin.RefreshableResultWithResultId
	unmarshalErr := json.Unmarshal([]byte(resultStr), &result)
	if unmarshalErr != nil {
		logger.Error(ctx, unmarshalErr.Error())
		responseUIError(ctx, request, unmarshalErr.Error())
		return
	}

	startTime := util.GetSystemTimestamp()
	logger.Debug(ctx, fmt.Sprintf("start executing refresh for result: %s (resultId:%s, queryId:%s)", result.Title, result.ResultId, queryId))

	// replace remote preview with local preview
	if result.Preview.PreviewType == plugin.WoxPreviewTypeRemote {
		preview, err := plugin.GetPluginManager().GetResultPreview(util.NewTraceContext(), result.ResultId)
		if err != nil {
			logger.Error(ctx, err.Error())
			responseUIError(ctx, request, err.Error())
			return
		}
		result.Preview = preview
	}

	newResult, refreshErr := plugin.GetPluginManager().ExecuteRefresh(ctx, result)
	logger.Debug(ctx, fmt.Sprintf("finished refresh %s, cost: %dms", result.ResultId, util.GetSystemTimestamp()-startTime))
	if refreshErr != nil {
		logger.Error(ctx, refreshErr.Error())
		responseUIError(ctx, request, refreshErr.Error())
		return
	}

	responseUISuccessWithData(ctx, request, newResult)
}

func getWebsocketMsgParameter(ctx context.Context, msg WebsocketMsg, key string) (string, error) {
	jsonData, marshalErr := json.Marshal(msg.Data)
	if marshalErr != nil {
		return "", marshalErr
	}

	paramterData := gjson.GetBytes(jsonData, key)
	if !paramterData.Exists() {
		return "", fmt.Errorf("%s parameter not found", key)
	}

	return paramterData.String(), nil
}