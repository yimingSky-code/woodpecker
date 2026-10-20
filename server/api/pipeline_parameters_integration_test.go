// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build test

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub/memory"
	queue_mocks "go.woodpecker-ci.org/woodpecker/v3/server/queue/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v3/server/scheduler"
	config_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/config/mocks"
	manager_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/mocks"
	registry_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/registry/mocks"
	secret_service_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/secret/mocks"
)

// 真实 HTTP handler、SQLite migration、配置编译与调度；forge/queue 是边界替身，不连接外部发布目标。
func TestTypedParametersManualDeployIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	originalConfig := server.Config
	t.Cleanup(func() { server.Config = originalConfig })
	s := newTestStore(t)
	repo, user := parameterFixture(t, s)
	repo.AllowDeploy = true
	repo.Branch = "main"
	repo.CancelPreviousPipelineEvents = nil
	require.NoError(t, s.UpdateRepo(repo))

	forge := forge_mocks.NewMockForge(t)
	forge.On("Name").Return("mock").Maybe()
	forge.On("URL").Return("https://example.com").Maybe()
	forge.On("BranchHead", mock.Anything, mock.Anything, mock.Anything, "main").Return(&model.Commit{SHA: "fixed-build-sha", ForgeURL: "https://example.com/commit/fixed-build-sha"}, nil).Once()
	forge.On("Netrc", mock.Anything, mock.Anything).Return(&model.Netrc{}, nil).Maybe()
	forge.On("Status", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	configs := config_service_mocks.NewMockService(t)
	manualConfig := []*forge_types.FileMeta{{Name: ".woodpecker.yaml", Data: []byte("when:\n  event: [manual, deployment]\nsteps:\n  probe:\n    image: alpine:3.21\n    commands:\n      - echo ${SERVICES}\n")}}
	// Deploy 必须接到已持久化配置；不借此重新获取分支 HEAD。
	configs.On("Fetch", mock.Anything, forge, mock.Anything, repo, mock.Anything, mock.Anything, false).Return(manualConfig, nil).Once()
	configs.On("Fetch", mock.Anything, forge, mock.Anything, repo, mock.Anything, mock.Anything, true).Run(func(args mock.Arguments) {
		source, ok := args.Get(4).(*model.Pipeline)
		require.True(t, ok)
		assert.Equal(t, "fixed-build-sha", source.Commit)
		old, ok := args.Get(5).([]*forge_types.FileMeta)
		require.True(t, ok)
		require.Len(t, old, 1)
		assert.Equal(t, manualConfig[0].Data, old[0].Data)
	}).Return(manualConfig, nil)
	secrets := secret_service_mocks.NewMockService(t)
	secrets.On("SecretListPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*model.Secret{}, nil).Maybe()
	registries := registry_service_mocks.NewMockService(t)
	registries.On("RegistryListPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*model.Registry{}, nil).Maybe()
	manager := manager_mocks.NewMockManager(t)
	manager.On("ForgeFromRepo", mock.Anything).Return(forge, nil)
	manager.On("ForgeFromUser", mock.Anything).Return(forge, nil).Maybe()
	manager.On("ConfigServiceFromRepo", mock.Anything).Return(configs)
	manager.On("SecretServiceFromRepo", mock.Anything).Return(secrets).Maybe()
	manager.On("RegistryServiceFromRepo", mock.Anything).Return(registries).Maybe()
	manager.On("EnvironmentService").Return(nil).Maybe()
	server.Config.Services.Manager = manager
	queue := queue_mocks.NewMockQueue(t)
	var queued []*model.Task
	queue.On("PushAtOnce", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		tasks, ok := args.Get(1).([]*model.Task)
		require.True(t, ok)
		queued = append(queued, tasks...)
	}).Return(nil)
	server.Config.Services.Scheduler = scheduler.NewScheduler(t.Context(), s, queue, memory.New())

	pushAllowed := true
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("store", s)
		c.Set("repo", repo)
		c.Set("user", user)
		c.Set("perm", &model.Perm{Push: pushAllowed})
		c.Next()
	})
	router.POST("/parameters", session.MustPush, PostParameter)
	router.GET("/parameters", session.MustPush, GetParameterList)
	router.POST("/pipelines", session.MustPush, CreatePipeline)
	router.POST("/pipelines/:number", session.MustPush, func(c *gin.Context) {
		var number int64
		_, err := fmt.Sscan(c.Param("number"), &number)
		require.NoError(t, err)
		source, err := s.GetPipelineNumber(repo, number)
		require.NoError(t, err)
		c.Set("pipeline", source)
		c.Next()
	}, PostPipeline)
	request := func(method, path string, body any) *httptest.ResponseRecorder {
		t.Helper()
		data, err := json.Marshal(body)
		require.NoError(t, err)
		req := httptest.NewRequest(method, path, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}
	for _, definition := range []model.Parameter{
		{Name: "SERVICES", Type: model.ParameterTypeMultiChoice, Event: model.EventManual, Options: []string{"server", "worker"}, Required: true},
		{Name: "SERVICES", Type: model.ParameterTypeMultiChoice, Event: model.EventDeploy, Options: []string{"server", "worker"}, Required: true},
		{Name: "deploy_to", Type: model.ParameterTypeChoice, Event: model.EventDeploy, Options: []string{"testing", "production"}, Required: true},
		{Name: "deploy_task", Type: model.ParameterTypeChoice, Event: model.EventDeploy, Options: []string{"hot", "cold"}, Default: "hot"},
	} {
		response := request(http.MethodPost, "/parameters", definition)
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	}
	listed := request(http.MethodGet, "/parameters", nil)
	require.Equal(t, http.StatusOK, listed.Code)
	var definitions []*model.Parameter
	require.NoError(t, json.Unmarshal(listed.Body.Bytes(), &definitions))
	require.Len(t, definitions, 4)

	invalid := request(http.MethodPost, "/pipelines", map[string]any{"branch": "main", "variables": map[string]string{"SERVICES": "server,unknown"}})
	require.Equal(t, http.StatusUnprocessableEntity, invalid.Code)
	require.Empty(t, queued)
	created := request(http.MethodPost, "/pipelines", map[string]any{"branch": "main", "variables": map[string]string{"SERVICES": "server,worker", "EXTRA": "retained"}})
	require.Equal(t, http.StatusOK, created.Code, created.Body.String())
	var build model.Pipeline
	require.NoError(t, json.Unmarshal(created.Body.Bytes(), &build))
	require.Equal(t, "fixed-build-sha", build.Commit)
	require.Len(t, queued, 1)
	build.RepoID = repo.ID
	build.Status = model.StatusSuccess
	require.NoError(t, s.UpdatePipeline(&build))

	for _, bad := range []string{"deploy_to=unknown&SERVICES=server", "deploy_to=testing&SERVICES=", "deploy_to=testing&SERVICES=server,server", "deploy_to=testing&deploy_task=unknown"} {
		result := request(http.MethodPost, fmt.Sprintf("/pipelines/%d?event=deployment&%s", build.Number, bad), nil)
		require.Equal(t, http.StatusUnprocessableEntity, result.Code, result.Body.String())
	}
	require.Len(t, queued, 1, "非法输入不可排入执行队列")
	previous := build.Number
	for index, target := range []string{"testing", "production"} {
		result := request(http.MethodPost, fmt.Sprintf("/pipelines/%d?event=deployment&deploy_to=%s&deploy_task=cold&SERVICES=worker", previous, target), nil)
		require.Equal(t, http.StatusOK, result.Code, result.Body.String())
		var deployed model.Pipeline
		require.NoError(t, json.Unmarshal(result.Body.Bytes(), &deployed))
		persisted, err := s.GetPipelineNumber(repo, deployed.Number)
		require.NoError(t, err)
		assert.Equal(t, previous, persisted.Parent)
		assert.Equal(t, build.Commit, persisted.Commit)
		assert.Equal(t, target, persisted.DeployTo)
		assert.Equal(t, "cold", persisted.DeployTask)
		assert.Equal(t, "worker", persisted.AdditionalVariables["SERVICES"])
		assert.Equal(t, "retained", persisted.AdditionalVariables["EXTRA"])
		assert.NotContains(t, persisted.AdditionalVariables, "deploy_to")
		assert.NotContains(t, persisted.AdditionalVariables, "deploy_task")
		require.Len(t, queued, index+2)
		assert.True(t, strings.Contains(string(queued[index+1].Data), "worker"), "编译后的实际任务需要使用覆盖后的服务")
		assert.NotContains(t, string(queued[index+1].Data), "server,worker")
		previous = persisted.Number
	}
	untouched, err := s.GetPipelineNumber(repo, build.Number)
	require.NoError(t, err)
	assert.Equal(t, "server,worker", untouched.AdditionalVariables["SERVICES"])
	assert.Empty(t, untouched.DeployTo)

	before := len(queued)
	pushAllowed = false
	denied := request(http.MethodPost, fmt.Sprintf("/pipelines/%d?event=deployment&deploy_to=production", build.Number), nil)
	assert.Equal(t, http.StatusNotFound, denied.Code)
	assert.Equal(t, before, len(queued))
	pushAllowed = true
	repo.AllowDeploy = false
	denied = request(http.MethodPost, fmt.Sprintf("/pipelines/%d?event=deployment&deploy_to=production", build.Number), nil)
	assert.Equal(t, http.StatusForbidden, denied.Code)
	assert.Equal(t, before, len(queued))
}
