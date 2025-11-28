package mock

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

type ApiMock struct {
	headersReceived       map[string]map[int]map[string]string
	queriesReceived       map[string]map[int]map[string]string
	requestsReceived      map[string]map[int]map[string]any
	responseMap           map[string]map[int]any
	defaultResponseMap    map[string]map[int]any
	responseStatus        map[string]map[int]int
	defaultResponseStatus map[string]map[int]int
	mockUrl               string
}

func NewApiServer() *ApiMock {
	return &ApiMock{
		headersReceived:       map[string]map[int]map[string]string{},
		queriesReceived:       map[string]map[int]map[string]string{},
		requestsReceived:      map[string]map[int]map[string]any{},
		responseMap:           map[string]map[int]any{},
		defaultResponseMap:    map[string]map[int]any{},
		responseStatus:        map[string]map[int]int{},
		defaultResponseStatus: map[string]map[int]int{},
	}
}

func (a *ApiMock) Start() {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				method := r.Method
				path := r.URL.Path
				index := len(a.requestsReceived[method+path])

				body, _ := io.ReadAll(r.Body)
				var request map[string]any
				_ = json.Unmarshal(body, &request)

				if a.requestsReceived[method+path] == nil {
					a.requestsReceived[method+path] = map[int]map[string]any{}
				}
				if request == nil {
					request = map[string]any{}
				}
				a.requestsReceived[method+path][index] = request

				if a.headersReceived[method+path] == nil {
					a.headersReceived[method+path] = map[int]map[string]string{}
				}
				a.headersReceived[method+path][index] = map[string]string{}
				for key, value := range r.Header {
					a.headersReceived[method+path][index][key] = value[0]
				}

				if a.queriesReceived[method+path] == nil {
					a.queriesReceived[method+path] = map[int]map[string]string{}
				}
				a.queriesReceived[method+path][index] = map[string]string{}
				for key, value := range r.URL.Query() {
					a.queriesReceived[method+path][index][key] = value[0]
				}
				status := a.getResponseStatus(method, path, index)
				w.WriteHeader(status)
				_, _ = w.Write([]byte(a.createBaseResponse(index, method, path)))
			},
		),
	)

	a.mockUrl = server.URL
}

func (a *ApiMock) GetUrl() string {
	return a.mockUrl
}

func (a *ApiMock) SetResponse(index int, method, path string, status int, response map[string]any) {
	if a.headersReceived[method+path] == nil {
		a.headersReceived[method+path] = map[int]map[string]string{}
	}
	if a.queriesReceived[method+path] == nil {
		a.queriesReceived[method+path] = map[int]map[string]string{}
	}
	if a.responseMap[method+path] == nil {
		a.responseMap[method+path] = map[int]any{}
	}
	if a.requestsReceived[method+path] == nil {
		a.requestsReceived[method+path] = map[int]map[string]any{}
	}
	if a.responseStatus[method+path] == nil {
		a.responseStatus[method+path] = map[int]int{}
	}
	if a.defaultResponseMap[method+path] == nil {
		a.defaultResponseMap[method+path] = map[int]any{}
	}
	if a.defaultResponseStatus[method+path] == nil {
		a.defaultResponseStatus[method+path] = map[int]int{}
	}
	if index == -1 {
		a.defaultResponseStatus[method+path][0] = status
		a.defaultResponseMap[method+path][0] = response
	} else {
		a.responseMap[method+path][index] = response
		a.responseStatus[method+path][index] = status
	}
}

func (a *ApiMock) GetRequestBody(method, path string, index int) map[string]any {
	key := a.findMatchingKeyGeneric(a.getMapKeys(a.requestsReceived), method, path, true)
	if key != "" && a.requestsReceived[key] != nil {
		if request, exists := a.requestsReceived[key][index]; exists {
			return request
		}
	}
	return nil
}

func (a *ApiMock) GetRequestHeaders(method, path string, index int) map[string]string {
	key := a.findMatchingKeyGeneric(a.getMapKeys(a.headersReceived), method, path, true)
	if key != "" && a.headersReceived[key] != nil {
		if headers, exists := a.headersReceived[key][index]; exists {
			return headers
		}
	}
	return nil
}

func (a *ApiMock) GetRequestQueries(method, path string, index int) map[string]string {
	key := a.findMatchingKeyGeneric(a.getMapKeys(a.queriesReceived), method, path, true)
	if key != "" && a.queriesReceived[key] != nil {
		if queries, exists := a.queriesReceived[key][index]; exists {
			return queries
		}
	}
	return nil
}

func (a *ApiMock) createBaseResponse(index int, method, path string) string {
	response := a.getResponseBody(method, path, index)

	responseString, _ := json.Marshal(response)

	return string(responseString)
}

func (a *ApiMock) ClearResponses(method, path string) {
	for key := range a.headersReceived {
		if strings.HasPrefix(key, method+path) {
			delete(a.headersReceived, key)
		}
	}
	for key := range a.requestsReceived {
		if strings.HasPrefix(key, method+path) {
			delete(a.requestsReceived, key)
		}
	}
	for key := range a.responseMap {
		if strings.HasPrefix(key, method+path) {
			delete(a.responseMap, key)
		}
	}
	for key := range a.responseStatus {
		if strings.HasPrefix(key, method+path) {
			delete(a.responseStatus, key)
		}
	}
	for key := range a.queriesReceived {
		if strings.HasPrefix(key, method+path) {
			delete(a.queriesReceived, key)
		}
	}
	for key := range a.defaultResponseMap {
		if strings.HasPrefix(key, method+path) {
			delete(a.defaultResponseMap, key)
		}
	}
	for key := range a.defaultResponseStatus {
		if strings.HasPrefix(key, method+path) {
			delete(a.defaultResponseStatus, key)
		}
	}
}

func (a *ApiMock) getResponseBody(method string, path string, index int) any {
	key := a.findMatchingKeyGeneric(a.getMapKeys(a.responseMap), method, path)
	if key != "" && a.responseMap[key] != nil {
		if response, exists := a.responseMap[key][index]; exists {
			if response != nil {
				return response
			}
		}
	}

	defaultKey := a.findMatchingKeyGeneric(a.getMapKeys(a.defaultResponseMap), method, path)
	if defaultKey != "" && a.defaultResponseMap[defaultKey] != nil {
		if response, exists := a.defaultResponseMap[defaultKey][0]; exists {
			if response != nil {
				return response
			}
		}
	}

	response := map[string]any{}
	exactKey := method + path
	if a.responseMap[exactKey] == nil {
		a.responseMap[exactKey] = map[int]any{}
	}
	a.responseMap[exactKey][index] = response
	return response
}

func (a *ApiMock) getResponseStatus(method string, path string, index int) int {
	key := a.findMatchingKeyGeneric(a.getMapKeys(a.responseStatus), method, path)
	if key != "" && a.responseStatus[key] != nil {
		if status, exists := a.responseStatus[key][index]; exists {
			if status != 0 {
				return status
			}
		}
	}

	defaultKey := a.findMatchingKeyGeneric(a.getMapKeys(a.defaultResponseStatus), method, path)
	if defaultKey != "" && a.defaultResponseStatus[defaultKey] != nil {
		if status, exists := a.defaultResponseStatus[defaultKey][0]; exists {
			if status != 0 {
				return status
			}
		}
	}

	// Return 200 as a safe default to prevent panic from WriteHeader(0)
	return 200
}

func (a *ApiMock) matchPath(pattern string, path string) bool {
	if pattern == path {
		return true
	}

	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i := range patternParts {
		if patternParts[i] != "*" && pathParts[i] != "*" && patternParts[i] != pathParts[i] {
			return false
		}
	}

	return true
}

func (a *ApiMock) findMatchingKeyGeneric(keys []string, method string, path string, strict ...bool) string {
	isStrict := false
	if len(strict) > 0 {
		isStrict = strict[0]
	}

	exactKey := method + path
	for _, key := range keys {
		if key == exactKey && (!isStrict || !strings.Contains(key, "*")) {
			return key
		}
	}

	for _, key := range keys {
		if isStrict && strings.Contains(key, "*") {
			continue
		}
		if strings.HasPrefix(key, method) {
			keyPath := strings.TrimPrefix(key, method)
			if a.matchPath(keyPath, path) {
				return key
			}
		}
	}

	return ""
}

func (a *ApiMock) getMapKeys(m interface{}) []string {
	var keys []string
	switch v := m.(type) {
	case map[string]map[int]map[string]any:
		for key := range v {
			keys = append(keys, key)
		}
	case map[string]map[int]map[string]string:
		for key := range v {
			keys = append(keys, key)
		}
	case map[string]map[int]any:
		for key := range v {
			keys = append(keys, key)
		}
	case map[string]map[int]int:
		for key := range v {
			keys = append(keys, key)
		}
	}
	return keys
}
