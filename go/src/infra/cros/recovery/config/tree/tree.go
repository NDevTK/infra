// Copyright 2022 The Chromium OS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tree

import (
	"infra/cros/recovery/config"
)

// Configuration provides the plans to be used by the recovery engine.
type Configuration struct {
	// List of plans provided by configuration.
	Plans []*Plan `json:"plans,omitempty"`
}

// ConvertConfiguration converts Configuration to a tree representation.
func ConvertConfiguration(configuration *config.Configuration) *Configuration {
	c := &Configuration{}
	for _, planName := range configuration.GetPlanNames() {
		if planName == config.PlanClosing {
			// Skip it as Closing plan is always last.
			continue
		}
		p := configuration.GetPlans()[planName]
		c.Plans = append(c.Plans, ConvertPlan(planName, p))
	}
	if p, ok := configuration.GetPlans()[config.PlanClosing]; ok {
		c.Plans = append(c.Plans, ConvertPlan(config.PlanClosing, p))
	}
	return c
}

// Plan holds information about actions to execute.
type Plan struct {
	// Name of the Plan.
	Name string `json:"name,omitempty"`
	// Critical actions are actions which have to pass for plan to succeed.
	CriticalActions []*Action `json:"critical_actions,omitempty"`
	// Is Plan is allowed to fail. If not then running of configuration will be stopped.
	AllowFail bool `json:"allow_fail,omitempty"`
}

// planConverter holds data to convert actions of single plan.
type planConverter struct {
	srcPlan *config.Plan
	actions map[string]*Action
}

// ConvertPlan converts Plan to a tree representation.
func ConvertPlan(name string, plan *config.Plan) *Plan {
	p := &Plan{
		Name:      name,
		AllowFail: plan.GetAllowFail(),
	}
	converter := &planConverter{
		srcPlan: plan,
		actions: make(map[string]*Action, len(plan.GetActions())),
	}
	for _, actionName := range plan.GetCriticalActions() {
		converter.convertAcion(actionName)
		p.CriticalActions = append(p.CriticalActions, converter.actions[actionName])
	}
	return p
}

// Action describes how to run the action, including its dependencies,
// conditions, and other attributes.
type Action struct {
	// Name of the Action.
	Name string `json:"name,omitempty"`
	// Documentation to describe detail of the action.
	Docs []string `json:"docs,omitempty"`
	// List of actions to determine if this action is applicable for the resource.
	Conditions []interface{} `json:"conditions,omitempty"`
	// List of actions that must pass before executing this action's exec function.
	Dependencies []interface{} `json:"dependencies,omitempty"`
	// Name of the exec function to use.
	ExecName string `json:"exec_name,omitempty"`
	// Extra arguments provided to the exec function.
	ExecArgs []string `json:"exec_args,omitempty"`
	// Allowed time to execute exec function.
	Timeout string `json:"timeout,omitempty"`
	// List of actions used to recover this action if exec function fails.
	Recoveries []interface{} `json:"recoveries,omitempty"`
	// If set to true, then the action is treated as if it passed even if it
	// and all its recovery actions failed.
	AllowFailAfterRecovery bool `json:"allow_fail_after_recovery,omitempty"`
	// Controls how and when the action can be rerun throughout the plan.
	RunControl string `json:"run_control,omitempty"`
}

func (t *planConverter) convertAcion(name string) {
	if _, ok := t.actions[name]; ok {
		// Action already converted.
		return
	}
	action := t.srcPlan.GetActions()[name]
	a := &Action{
		Name:                   name,
		Docs:                   action.GetDocs(),
		ExecName:               action.GetExecName(),
		ExecArgs:               action.GetExecExtraArgs(),
		AllowFailAfterRecovery: action.GetAllowFailAfterRecovery(),
		RunControl:             action.GetRunControl().String(),
	}
	if action.GetExecTimeout() != nil {
		a.Timeout = action.GetExecTimeout().AsDuration().String()
	}
	t.actions[name] = a
	for _, actionName := range action.GetConditions() {
		t.convertAcion(actionName)
		a.Conditions = append(a.Conditions, t.actions[actionName])
	}
	for _, actionName := range action.GetDependencies() {
		t.convertAcion(actionName)
		a.Dependencies = append(a.Dependencies, t.actions[actionName])
	}
	for _, actionName := range action.GetRecoveryActions() {
		t.convertAcion(actionName)
		a.Recoveries = append(a.Recoveries, t.actions[actionName])
	}
}
