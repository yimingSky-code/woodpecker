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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParameterValidate(t *testing.T) {
	tests := []struct {
		name      string
		parameter Parameter
		wantErr   bool
	}{
		{
			name:      "valid string parameter",
			parameter: Parameter{Name: "SOME_VAR", Type: ParameterTypeString},
		},
		{
			name:      "valid choice parameter",
			parameter: Parameter{Name: "TARGET", Type: ParameterTypeChoice, Options: []string{"a", "b"}, Default: "a"},
		},
		{
			name:      "invalid env var name",
			parameter: Parameter{Name: "not valid", Type: ParameterTypeString},
			wantErr:   true,
		},
		{
			name:      "name starting with digit",
			parameter: Parameter{Name: "1VAR", Type: ParameterTypeString},
			wantErr:   true,
		},
		{
			name:      "unknown type",
			parameter: Parameter{Name: "SOME_VAR", Type: "unknown"},
			wantErr:   true,
		},
		{
			name:      "choice without options",
			parameter: Parameter{Name: "TARGET", Type: ParameterTypeChoice},
			wantErr:   true,
		},
		{
			name:      "choice default not in options",
			parameter: Parameter{Name: "TARGET", Type: ParameterTypeChoice, Options: []string{"a", "b"}, Default: "c"},
			wantErr:   true,
		},
		{
			name:      "boolean with invalid default",
			parameter: Parameter{Name: "FLAG", Type: ParameterTypeBoolean, Default: "yes"},
			wantErr:   true,
		},
		{
			name:      "required boolean is rejected",
			parameter: Parameter{Name: "FLAG", Type: ParameterTypeBoolean, Required: true},
			wantErr:   true,
		},
		{
			name:      "number with invalid default",
			parameter: Parameter{Name: "COUNT", Type: ParameterTypeNumber, Default: "abc"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.parameter.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateParameterValues(t *testing.T) {
	params := []*Parameter{
		{Name: "TARGET", Type: ParameterTypeChoice, Options: []string{"staging", "production"}, Default: "staging"},
		{Name: "REASON", Type: ParameterTypeString, Required: true},
		{Name: "DRY_RUN", Type: ParameterTypeBoolean, Default: "false"},
	}

	t.Run("fills defaults and keeps extra variables", func(t *testing.T) {
		values := map[string]string{"REASON": "hotfix", "AD_HOC": "kept"}
		assert.NoError(t, ValidateParameterValues(params, values))
		assert.Equal(t, map[string]string{
			"TARGET":  "staging",
			"REASON":  "hotfix",
			"DRY_RUN": "false",
			"AD_HOC":  "kept",
		}, values)
	})

	t.Run("missing required parameter fails", func(t *testing.T) {
		values := map[string]string{}
		assert.Error(t, ValidateParameterValues(params, values))
	})

	t.Run("choice value outside options fails", func(t *testing.T) {
		values := map[string]string{"TARGET": "nope", "REASON": "x"}
		assert.Error(t, ValidateParameterValues(params, values))
	})

	t.Run("no parameters leaves values untouched", func(t *testing.T) {
		values := map[string]string{"ANY": "thing"}
		assert.NoError(t, ValidateParameterValues(nil, values))
		assert.Equal(t, map[string]string{"ANY": "thing"}, values)
	})
}

func TestMultiChoiceAndEventParameters(t *testing.T) {
	p := &Parameter{Name: "BUILD_SERVICES", Type: ParameterTypeMultiChoice, Options: []string{"server", "scheduler", "worker"}, Default: "server,worker", Required: true}
	assert.NoError(t, p.Validate())
	values := map[string]string{}
	assert.NoError(t, ValidateParameterValues([]*Parameter{p}, values))
	assert.Equal(t, "server,worker", values[p.Name])
	for _, value := range []string{"", "server,unknown", "server,server", "server,", "server, worker"} {
		assert.Error(t, ValidateParameterValues([]*Parameter{p}, map[string]string{p.Name: value}), value)
	}
	p.Required = false
	values[p.Name] = ""
	assert.NoError(t, ValidateParameterValues([]*Parameter{p}, values))
	assert.Equal(t, "", values[p.Name], "显式清空不能重新选择默认服务")
	p.Options = []string{"server,worker"}
	assert.Error(t, p.Validate())
	manual := &Parameter{Name: "SERVICES", Type: ParameterTypeString}
	deployment := &Parameter{Name: "deploy_to", Event: EventDeploy, Type: ParameterTypeChoice, Options: []string{"testing", "production"}, Required: true}
	assert.NoError(t, deployment.Validate())
	assert.Equal(t, []*Parameter{manual}, ParametersForEvent([]*Parameter{manual, deployment}, EventManual))
	assert.Equal(t, []*Parameter{deployment}, ParametersForEvent([]*Parameter{manual, deployment}, EventDeploy))
	deployment.Event = EventManual
	assert.Error(t, deployment.Validate())
	for _, value := range []string{"NaN", "+Inf", "-Inf"} {
		assert.Error(t, (&Parameter{Name: "N", Type: ParameterTypeNumber}).ValidateValue(value))
	}
}
