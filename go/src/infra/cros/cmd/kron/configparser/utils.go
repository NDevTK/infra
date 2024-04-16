// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package configparser implements logic to handle SuiteScheduler configuration files.
package configparser

import (
	"fmt"

	suschpb "go.chromium.org/chromiumos/infra/proto/go/testplans"
)

// isDayCompliant checks the day int type to ensure that it is within the
// accepted bounds. A flag for fortnightly is required for calculation of day
// range values.
func isDayCompliant(day int, isFortnightly bool) error {
	highBound := 6

	if isFortnightly {
		highBound = 13
	}

	if day < 0 || day > highBound {
		return fmt.Errorf("hay %d is not within the supported range [0,%d]", day, highBound)
	}

	return nil
}

// isHourCompliant checks the hour int type to ensure that it is within the
// accepted bounds.
func isHourCompliant(hour int) error {
	if hour < 0 || hour > 23 {
		return fmt.Errorf("hour %d is not within the supported range [0,23]", hour)
	}

	return nil
}

// getBoardsList returns a TargetOptions map of the boards being tracked by the
// given config filters.
//
// The rules are as follows:
//   - BoardList (explicitly add boards)
//   - ExcludeBoards (add ALL non excluded boards)
//
// NOTE: Some configs may provide neither of the options below and the boards
// to-be-tracked are added via the VariantsList. This logic is non-intuitive and
// may be removed later.
func getBoardsList(targets TargetOptions, labBoards map[Board]*BoardEntry, boardsList, excludeBoardsList []string) (TargetOptions, error) {
	// If boards provided, add to list
	if len(boardsList) != 0 {
		for _, board := range boardsList {
			// If the board is in the lab config add it to the list.
			if _, ok := labBoards[Board(board)]; !ok {
				return nil, fmt.Errorf("Board %s not in the lab config", board)
			}
			targets[Board(board)] = &TargetOption{Board: board}
		}
	} else if len(excludeBoardsList) != 0 {
		// Build the map we'll use to capture explicit excludes.
		excludeBoards := map[Board]bool{}
		for _, board := range excludeBoardsList {
			excludeBoards[Board(board)] = true
		}

		for _, board := range labBoards {
			// The board is excluded, do not add to target list.
			if _, ok := excludeBoards[Board(board.GetName())]; ok {
				continue
			}
			targets[Board(board.GetName())] = &TargetOption{Board: board.GetName()}
		}
	}

	return targets, nil
}

// getModelsList returns a TargetOptions map of the models being tracked by the
// given config filters.
//
// The rules are as follows:
//   - AnyModel (if false and no models are provided add all models of target boards)
//   - ModelsList (explicitly add models)
//   - ExcludeModels (add ALL non excluded models)
func getModelsList(targets TargetOptions, labModels map[Model]*BoardEntry, labBoards map[Board]*BoardEntry, modelsList, excludeModelsList []string, AnyModel bool) (TargetOptions, error) {
	// If models provided, add to list.
	if len(modelsList) != 0 {
		for _, model := range modelsList {
			// If the board is in the lab config add it to the list.
			entry, ok := labModels[Model(model)]
			if !ok {
				return nil, fmt.Errorf("model %s not in the lab config", model)
			}
			if _, ok := targets[Board(entry.board.Name)]; !ok {
				targets[Board(entry.GetName())] = &TargetOption{Board: entry.GetName()}
			}
			targets[Board(entry.board.Name)].Models = append(targets[Board(entry.board.Name)].Models, model)
		}
	} else if len(excludeModelsList) != 0 {
		// If no explicit models were provided then add all models which are not
		// explicitly excluded.

		// Build the map we'll use to capture explicit excludes.
		excludeModels := map[Model]bool{}
		for _, model := range excludeModelsList {
			excludeModels[Model(model)] = true
		}

		// Iterate through board targets and add models which aren't excluded.
		for boardName, target := range targets {
			// Ensure the board exists in the lab configuration.
			if _, ok := labBoards[boardName]; !ok {
				return nil, fmt.Errorf("target list is tracking a board not seen in the lab configurations")
			}

			// Iterate though the models of the current board and check for any
			// explicit exclude rules.
			for _, model := range labBoards[boardName].board.Models {
				if _, ok := excludeModels[Model(model)]; ok {
					continue
				}
				target.Models = append(target.Models, model)
			}
		}
	} else if len(modelsList) == 0 && !AnyModel {
		// Iterate through board targets and add models all models since none
		// are excluded.
		for boardName, target := range targets {
			// Ensure the board exists in the lab configuration.
			if _, ok := labBoards[boardName]; !ok {
				return nil, fmt.Errorf("target list is tracking a board not seen in the lab configurations")
			}

			target.Models = append(target.Models, labBoards[boardName].board.Models...)
		}
	}
	return targets, nil
}

// getVariantsList returns a TargetOptions map of the variants being tracked by the
// given config filters.
//
// The rules are as follows:
//   - SkipVariants (differing from v1 SuSch this function will not be called if this is true)
//   - VariantsList (Explicitly add variants of specific boards)
//   - ExcludeVariants (Exclude specific variants but add all other variants per board)
//
// NOTE: Some configs may provide no board options and instead add targeted
// boards via the variants options. This logic is non-intuitive and
// may be removed later.
func getVariantsList(targets TargetOptions, labBoards map[Board]*BoardEntry, variantsList, excludeVariantsList []*suschpb.BoardVariant) (TargetOptions, error) {
	if len(variantsList) != 0 {
		for _, variant := range variantsList {
			if _, ok := targets[Board(variant.Board)]; !ok {
				targets[Board(variant.Board)] = &TargetOption{
					Board: variant.Board,
				}
			}

			targets[Board(variant.Board)].Variants = append(targets[Board(variant.Board)].Variants, variant.Variant)
		}
	} else {
		// Build the map we'll use to capture explicit excludes.
		excludeVariantsMap := map[Board]map[Variant]bool{}
		for _, variant := range excludeVariantsList {
			subMap, ok := excludeVariantsMap[Board(variant.Board)]
			if !ok {
				excludeVariantsMap[Board(variant.Board)] = map[Variant]bool{}
				subMap = excludeVariantsMap[Board(variant.Board)]
			}

			subMap[Variant(variant.Variant)] = true
		}

		// This covers the case where no board options have been included but
		// rather, the expectation is that the variants options fill in the
		// targeted boards.
		if len(targets) == 0 {
			for boardName := range excludeVariantsMap {
				targets[boardName] = &TargetOption{
					Board:    string(boardName),
					Models:   []string{},
					Variants: []string{},
				}
			}
		}

		// Add all non-excluded variants.
		for boardName, target := range targets {
			// Ensure the board exists in the lab configuration.
			if _, ok := labBoards[boardName]; !ok {
				return nil, fmt.Errorf("target list is tracking a board not seen in the lab configurations")
			}

			excludedVariants, notExcluded := excludeVariantsMap[boardName]

			// Board is not outright excluded, add all variants.
			if !notExcluded {
				target.Variants = append(target.Variants, labBoards[boardName].board.Variants...)
				continue
			}

			// Add all variants which aren't excluded
			for _, variant := range labBoards[boardName].board.Variants {
				// If variant isn't excluded add to the target options.
				if _, ok := excludedVariants[Variant(variant)]; !ok {
					targets[boardName].Variants = append(targets[boardName].Variants, variant)
				}
			}
		}
	}

	return targets, nil
}

// GetTargetOptions adds all board(-variant)/models combinations specified by
// the config.
func GetTargetOptions(config *suschpb.SchedulerConfig, lab *LabConfigs) (TargetOptions, error) {
	targets := TargetOptions{}
	var err error

	// If ModelsList and (ExcludeBoards and ExcludeVariants) are provided, skip
	// adding the excludes as the modelsLists field takes precedence. This is
	// because the getModelsList adds the required board if not found. Not
	// including this edge case checker leads to us adding all boards/models
	// which weren't excluded explicitly.
	modelsListOnly := len(config.TargetOptions.ModelsList) > 0 && len(config.TargetOptions.ExcludeBoards) > 0 && len(config.TargetOptions.ExcludeVariants) > 0

	// If skip variants is set to off (default value for bool) but no other
	// variant includes are provided then we should act as if it was set to
	// true. This functionally sets the default value to true.
	defaultSkipVariants := !config.TargetOptions.SkipVariants && len(config.TargetOptions.VariantsList) == 0 && len(config.TargetOptions.ExcludeVariants) == 0

	// In the cases where only a variantsList is provided we don't want to
	// target non-variant buildTargets.
	variantsOnly := len(config.TargetOptions.BoardsList) == 0 && len(config.TargetOptions.VariantsList) > 0

	if !modelsListOnly {
		// Add boards to targets list
		targets, err = getBoardsList(targets, lab.Boards, config.TargetOptions.BoardsList, config.TargetOptions.ExcludeBoards)
		if err != nil {
			return nil, err
		}

		// If we aren't skipping variants then begin adding the variants to the
		// list.
		if !config.TargetOptions.SkipVariants || defaultSkipVariants {
			targets, err = getVariantsList(targets, lab.Boards, config.TargetOptions.VariantsList, config.TargetOptions.ExcludeVariants)
			if err != nil {
				return nil, err
			}
		}
	}

	// If any model is given then do not add models to the target options list.
	if !config.TargetOptions.AnyModel {
		targets, err = getModelsList(targets, lab.Models, lab.Boards, config.TargetOptions.ModelsList, config.TargetOptions.ExcludeModels, config.TargetOptions.AnyModel)
		if err != nil {
			return nil, err
		}
	}

	// Apply the calculated variantsOnly value to each target as they all share
	// the same value in one config.
	for _, target := range targets {
		target.VariantsOnly = variantsOnly
	}

	return targets, nil
}

// GetBuildTargets generates the BuildTarget for the given target options. A
// build target is in the syntax of <board>(-<variant>). If there are no
// variants then this will return a single value. If there are variants in the
// target option then this will return a list of all combinations using the
// above syntax. Models are not considered when generating the BuildTarget
func GetBuildTargets(target *TargetOption, variantsOnly bool) []BuildTarget {
	buildTargets := []BuildTarget{}

	// If variantsOnly is false then add the board sans variant to the build
	// target list.
	if !variantsOnly {
		buildTargets = append(buildTargets, BuildTarget(target.Board))
	}

	// If the target has variants then build a build target for each variant
	// name. Typical form is board-<variant> name, but some exceptions apply
	// (64,_li, etc).
	if len(target.Variants) > 0 {
		for _, variant := range target.Variants {
			buildTargets = append(buildTargets, BuildTarget(fmt.Sprintf("%s%s", target.Board, variant)))
		}
	}
	return buildTargets
}

// GetBuildTargetsForAllTargets creates a list of build targets that the configuration is
// tracking. A build target refers to the build image that should be used by the
// CTP run. A build target is in the form of board(-<variant>).
func GetBuildTargetsForAllTargets(targetOptions TargetOptions) []BuildTarget {
	buildTargets := []BuildTarget{}
	for _, target := range targetOptions {
		buildTargets = append(buildTargets, GetBuildTargets(target, target.VariantsOnly)...)
	}

	return buildTargets
}
