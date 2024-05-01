// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tree

import (
	"fmt"

	"infra/cros/recovery/config"
)

// Configuration provides the plans to be used by the recovery engine.
type Configuration struct {
	// List of plans provided by configuration.
	Plans []*Plan `json:"plans,omitempty"`
}

// ConvertConfiguration converts Configuration to a tree representation.
func ConvertConfiguration(configuration *config.Configuration, shortVersion bool) *Configuration {
	c := &Configuration{}
	for _, planName := range configuration.GetPlanNames() {
		if planName == config.PlanClosing {
			// Skip it as Closing plan is always last.
			continue
		}
		p := configuration.GetPlans()[planName]
		c.Plans = append(c.Plans, ConvertPlan(planName, p, shortVersion))
	}
	if p, ok := configuration.GetPlans()[config.PlanClosing]; ok {
		c.Plans = append(c.Plans, ConvertPlan(config.PlanClosing, p, shortVersion))
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
}

// ConvertPlan converts Plan to a tree representation.
func ConvertPlan(name string, plan *config.Plan, shortVersion bool) *Plan {
	p := &Plan{
		Name:      name,
		AllowFail: plan.GetAllowFail(),
	}
	if shortVersion && p.AllowFail {
		p.AllowFail = false
		p.Name = fmt.Sprintf("%s (Allow to fail)", p.Name)
	}
	converter := &planConverter{srcPlan: plan}
	for _, actionName := range plan.GetCriticalActions() {
		p.CriticalActions = append(p.CriticalActions, converter.convertAcion(actionName, false, shortVersion))
	}
	return p
}

// Action describes how to run the action, including its dependencies,
// conditions, and other attributes.
type Action struct {
	// Name of the Action.
	Name string `json:"name,omitempty"`
	// Name of the exec function to use.
	ExecName string `json:"exec_name,omitempty"`
	// Documentation to describe detail of the action.
	Docs []string `json:"docs,omitempty"`
	// List of actions to determine if this action is applicable for the resource.
	Conditions []interface{} `json:"conditions,omitempty"`
	// List of actions that must pass before executing this action's exec function.
	Dependencies []interface{} `json:"dependencies,omitempty"`
	// List of actions used to recover this action if exec function fails.
	Recoveries []interface{} `json:"recoveries,omitempty"`
	// Extra arguments provided to the exec function.
	ExecArgs []string `json:"exec_args,omitempty"`
	// Allowed time to execute exec function.
	Timeout string `json:"timeout,omitempty"`
	// If set to true, then the action is treated as if it passed even if it
	// and all its recovery actions failed.
	AllowFailAfterRecovery bool `json:"allow_fail_after_recovery,omitempty"`
	// Controls how and when the action can be rerun throughout the plan.
	RunControl string `json:"run_control,omitempty"`
}

func (t *planConverter) convertAcion(name string, excludeRecoveries, shortVersion bool) *Action {
	action, ok := t.srcPlan.GetActions()[name]
	if !ok {
		return nil
	}
	var a *Action
	if shortVersion {
		a = &Action{
			Name: name,
		}
		if action.GetAllowFailAfterRecovery() {
			a.Name = fmt.Sprintf("%s (Allow to fail)", a.Name)
		}
		if name != action.GetExecName() {
			a.ExecName = action.GetExecName()
		}
	} else {
		a = &Action{
			Name:                   name,
			ExecName:               action.GetExecName(),
			AllowFailAfterRecovery: action.GetAllowFailAfterRecovery(),
		}
		a.Docs = action.GetDocs()
		a.ExecArgs = action.GetExecExtraArgs()
		a.RunControl = action.GetRunControl().String()
		if action.GetExecTimeout() != nil {
			a.Timeout = action.GetExecTimeout().AsDuration().String()
		}
	}
	for _, actionName := range action.GetConditions() {
		a.Conditions = append(a.Conditions, t.convertAcion(actionName, true, shortVersion))
	}
	for _, actionName := range action.GetDependencies() {
		a.Dependencies = append(a.Dependencies, t.convertAcion(actionName, excludeRecoveries || false, shortVersion))
	}
	if !excludeRecoveries {
		for _, actionName := range action.GetRecoveryActions() {
			a.Recoveries = append(a.Recoveries, t.convertAcion(actionName, true, shortVersion))
		}
	}
	return a
}
