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
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type ParameterType string

const (
	ParameterTypeString       ParameterType = "string"
	ParameterTypeNumber       ParameterType = "number"
	ParameterTypeBoolean      ParameterType = "boolean"
	ParameterTypeChoice       ParameterType = "choice"
	ParameterTypeMultiChoice  ParameterType = "multi_choice"
	ParameterSourceRepoConfig               = "repo_config"
)

var parameterNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Parameter 定义仓库表单输入；同名参数在 manual 与 deployment 之间独立配置。
type Parameter struct {
	ID          int64         `json:"id" xorm:"pk autoincr 'id'"`
	RepoID      int64         `json:"repo_id" xorm:"UNIQUE(s) INDEX 'repo_id'"`
	Event       WebhookEvent  `json:"event" xorm:"UNIQUE(s) INDEX 'event' default 'manual'"`
	Name        string        `json:"name" xorm:"UNIQUE(s) INDEX 'name'"`
	Type        ParameterType `json:"type" xorm:"'param_type'"`
	Description string        `json:"description" xorm:"TEXT 'description'"`
	// Default 仅在未提交该字段时生效；显式空值不会重新选中默认服务。
	Default string `json:"default" xorm:"TEXT 'default_value'"`
	// Options 为静态可选值；multi_choice 用逗号分隔传给既有变量通道。
	Options  []string `json:"options" xorm:"json 'options'"`
	Required bool     `json:"required" xorm:"required"`
	Order    int      `json:"order" xorm:"display_order"`
	Source   string   `json:"source" xorm:"source"`
} // @name Parameter

func (Parameter) TableName() string { return "parameters" }

func (p *Parameter) Validate() error {
	if p.Event == "" {
		p.Event = EventManual
	}
	if p.Event != EventManual && p.Event != EventDeploy {
		return fmt.Errorf("invalid parameter event %q", p.Event)
	}
	if !parameterNameRegex.MatchString(p.Name) || p.Name == "event" || p.Name == "fork" {
		return fmt.Errorf("parameter name %q must be a non-reserved environment variable identifier", p.Name)
	}
	if p.Name == "deploy_to" || p.Name == "deploy_task" {
		if p.Event != EventDeploy || (p.Type != ParameterTypeString && p.Type != ParameterTypeChoice) {
			return fmt.Errorf("native field %q requires a deployment string or choice parameter", p.Name)
		}
	}
	switch p.Type {
	case ParameterTypeString, ParameterTypeNumber, ParameterTypeBoolean, ParameterTypeChoice, ParameterTypeMultiChoice:
	default:
		return fmt.Errorf("invalid parameter type %q", p.Type)
	}
	if p.Type == ParameterTypeBoolean && p.Required {
		return fmt.Errorf("boolean parameter %q cannot be required", p.Name)
	}
	if p.Type == ParameterTypeChoice || p.Type == ParameterTypeMultiChoice {
		if len(p.Options) == 0 {
			return fmt.Errorf("parameter %q requires options", p.Name)
		}
		seen := map[string]bool{}
		for _, option := range p.Options {
			if option == "" || seen[option] || (p.Type == ParameterTypeMultiChoice && strings.Contains(option, ",")) {
				return fmt.Errorf("parameter %q has empty, duplicate or ambiguous options", p.Name)
			}
			seen[option] = true
		}
	}
	if p.Default != "" {
		if err := p.ValidateValue(p.Default); err != nil {
			return fmt.Errorf("invalid default: %w", err)
		}
	}
	return nil
}

func (p *Parameter) ValidateValue(value string) error {
	switch p.Type {
	case ParameterTypeNumber:
		n, err := strconv.ParseFloat(value, 64)
		if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
			return fmt.Errorf("parameter %q requires a finite number", p.Name)
		}
	case ParameterTypeBoolean:
		if value != "true" && value != "false" {
			return fmt.Errorf("parameter %q requires true or false", p.Name)
		}
	case ParameterTypeChoice:
		if !slices.Contains(p.Options, value) {
			return fmt.Errorf("parameter %q value is not an allowed option", p.Name)
		}
	case ParameterTypeMultiChoice:
		seen := map[string]bool{}
		for _, selected := range strings.Split(value, ",") {
			if seen[selected] || !slices.Contains(p.Options, selected) {
				return fmt.Errorf("parameter %q contains an invalid or duplicate choice", p.Name)
			}
			seen[selected] = true
		}
	}
	return nil
}

func ParametersForEvent(params []*Parameter, event WebhookEvent) []*Parameter {
	result := make([]*Parameter, 0)
	for _, p := range params {
		if p.Event == event || (p.Event == "" && event == EventManual) {
			result = append(result, p)
		}
	}
	return result
}

// ValidateParameterValues 兼容自由变量；只校验配置的字段，不增加发布业务门禁。
func ValidateParameterValues(params []*Parameter, values map[string]string) error {
	for _, p := range params {
		value, present := values[p.Name]
		if !present {
			value = p.Default
			if value == "" && p.Type == ParameterTypeBoolean {
				value = "false"
			}
			if value != "" {
				values[p.Name] = value
			}
		}
		if value == "" {
			if p.Type == ParameterTypeBoolean {
				return fmt.Errorf("parameter %q requires true or false", p.Name)
			}
			if p.Required {
				return fmt.Errorf("required parameter %q is missing", p.Name)
			}
			continue
		}
		if err := p.ValidateValue(value); err != nil {
			return err
		}
	}
	return nil
}

type ParameterPatch struct {
	Event       *WebhookEvent  `json:"event"`
	Name        *string        `json:"name"`
	Type        *ParameterType `json:"type"`
	Description *string        `json:"description"`
	Default     *string        `json:"default"`
	Options     []string       `json:"options"`
	Required    *bool          `json:"required"`
	Order       *int           `json:"order"`
} // @name ParameterPatch
