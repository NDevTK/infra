// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package subcommands includes subcommand logic that will be used for the CLI
// front end.
package subcommands

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"

	"infra/cros/cmd/suite_scheduler/common"
	"infra/cros/cmd/suite_scheduler/configparser"
	"infra/cros/cmd/suite_scheduler/ctprequest"
)

const (
	jsonMarshallIndent = "    "
)

var (
	stderr = log.New(os.Stderr, "", log.Lshortfile)
	stdout = log.New(os.Stdout, "", log.Lshortfile)
)

type CLIConfigList = map[time.Time]configparser.ConfigList

// configParserCommand is the struct which represents the configParser Subcommand.
type configParserCommand struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags

	searchAllConfigs     bool
	commandExecutionTime time.Time

	// Top-Level Filters
	// Top-Level filters group large amounts of configs by their trigger types.
	// To further narrow down results bottom-level filters will be used.

	newBuild    bool
	daily       bool
	weekly      bool
	fortnightly bool
	nextNHours  time.Duration

	// Bottom-Level Filters
	// Bottom-level filters are used selected configs based on their contents.

	// Hardware Specific filters
	board   string
	model   string
	variant string
	// Timed Event time range filters
	day  int64
	hour int64
	// Name Specific searching
	configName string
	contains   string

	// Env flags

	// I/O flags
	outputPath         string
	configCFGInputPath string
	labCFGInputPath    string
	// Optional flags
	startTime int64

	// Format Flags
	asCtpRequest bool
	nameOnly     bool
}

// setFlags adds also CLI flags to the subcommand.
func (c *configParserCommand) setFlags() {
	// Top Level Filters

	c.Flags.BoolVar(&c.newBuild, "new-build", false, "Fetch from NEW_BUILD triggered configs")
	c.Flags.BoolVar(&c.daily, "daily", false, "Fetch from { DAILY | NIGHTLY } triggered configs")
	c.Flags.BoolVar(&c.weekly, "weekly", false, "Fetch from WEEKLY triggered configs")
	c.Flags.BoolVar(&c.fortnightly, "fortnightly", false, "Fetch from FORTNIGHTLY triggered configs")
	c.Flags.DurationVar(&c.nextNHours, "hours-ahead",
		common.DefaultHoursAhead, "Number of hours ahead of the current time"+
			" to fetch configs. Format should be in <number>h, any time "+
			"less than an hour will be truncated. E.g. 3h30m will be "+
			"seen as 3h.")

	// Bottom Level Filters

	// Hardware Specific filters
	c.Flags.StringVar(&c.board, "board", common.DefaultString, "Search for configs which target the provided board only")
	c.Flags.StringVar(&c.model, "model", common.DefaultString, "Search for configs which target the provided model only. Board also required when using this argument.")
	c.Flags.StringVar(&c.variant, "variant", common.DefaultString,
		("Providing this field will return all NEW_BUILD configs which" +
			" target this build image. A build target is in the form of " + "board(-<variant>)."))
	// Timed Event time range filters
	c.Flags.Int64Var(&c.day, "day", common.DefaultInt64, "Filter out Timed Events results for a the given day. A Timed Event filter must be provided.")
	c.Flags.Int64Var(&c.hour, "hour", common.DefaultInt64, "Filter out Timed Events results for a the given hour. A Timed Event filter and day must be provided.")
	// Name Specific searching
	c.Flags.StringVar(&c.configName, "config-name", common.DefaultString, "Receive information for the specific given config.")
	c.Flags.StringVar(&c.contains, "contains", common.DefaultString, "Search for configs who's name contains the given text.")

	// Env flags

	// I/O flags
	c.Flags.StringVar(&c.outputPath, "output-path", common.DefaultString, "If provided then search JSON results will be written to the given path. if the file does not exist then it will be created.")
	c.Flags.StringVar(&c.configCFGInputPath, "config-input-path", common.DefaultString, "Provide if a local version of the config .cfg is planned on being used. If omitted, the program will fetch the ToT config .cfg from gerrit.")
	c.Flags.StringVar(&c.labCFGInputPath, "lab-input-path", common.DefaultString, "Provide if a local version of the lab .cfg is planned on being used. If omitted, the program will fetch the ToT lab .cfg from gerrit.")
	// Optional flags
	c.Flags.Int64Var(&c.startTime, "start-time", common.DefaultInt64,
		"[UNIX TIME] If defined, this will set the start time for the -hours-ahead query.")

	// Format flags

	c.Flags.BoolVar(&c.asCtpRequest, "ctp-request", false, "Configs will be returned as the CTP Requests they would generate.")
	c.Flags.BoolVar(&c.nameOnly, "name-only", false, "Only the name of the config will be returned.")

}

func GetConfigParserCommand(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "configs <options>",
		LongDesc: ("The configs command is used to access the config parsing and searching logic available to SuiteScheduler." +
			" The accepted usages are:" +
			"\n\t- suite-scheduler configs -new-build [ -board | -model | -variant | -configName | -contains ]" +
			"\n\t- suite-scheduler configs { -daily | -weekly | -fortnightly } [ -day | -hour | -configName | -contains | -board | -model | -variant ]" +
			"\n\t- suite-scheduler configs -hours-ahead [ -start-time | -board | -model | -variant | -configName | -contains ]" +
			"\n\t- suite-scheduler configs [ -configName | -contains | -board | -model | -variant ]\n" +
			"The flags -output-path, -config-input-path, -lab-input-path, -ctp-request, and -name-only can be used with any command."),
		CommandRun: func() subcommands.CommandRun {
			cmd := &configParserCommand{}
			cmd.authFlags = authcli.Flags{}
			cmd.authFlags.Register(cmd.GetFlags(), authOpts)
			cmd.setFlags()
			return cmd
		},
	}
}

// isSingularTopLevelFilter takes an array of bools and ensures that only one
// is set as true.
func isSingularTopLevelFilter(bools []bool) bool {
	count := 0

	for _, b := range bools {
		if count > 1 {
			return false
		}

		if b {
			count++
		}
	}

	return true
}

// validate reads the user given flags and ensures that no improper combinations
// were given.
func (c *configParserCommand) validate() error {

	// If the user did not select a top-level filter then we will assume that
	// they are trying to search from the set of all configs.
	c.searchAllConfigs = !(c.newBuild || c.daily || c.weekly || c.fortnightly || (c.nextNHours != common.DefaultHoursAhead))

	// GENERAL RULES

	// Only one top-level filter flag can be given for any CLI invocation.
	if !isSingularTopLevelFilter([]bool{c.newBuild, c.daily, c.weekly, c.fortnightly, c.nextNHours != common.DefaultHoursAhead}) {
		return fmt.Errorf("only one type of top-level filter can be provided")
	}

	if c.day != common.DefaultInt64 && (c.day > 13 || c.day < 0) {
		return fmt.Errorf("-day can only be within [0,6] for weekly and [0,13] for fortnightly")
	}

	if c.hour != common.DefaultInt64 && (c.hour > 23 || c.hour < 0) {
		return fmt.Errorf("-hour can only be within [0,23]")
	}

	// Model must be provided with an accompanying board. There is no strict
	// rule which states that two boards may not share a common model name so
	// this check eliminates and confusion should this lab configuration exist.
	if c.model != common.DefaultString && c.board == common.DefaultString {
		return fmt.Errorf("model cannot be provided without a board value")
	}

	// NOTE: This is a bit of strict search criteria but since many boards share
	// the same variant type (E.g. -kernelnext) I(juahurta) felt that it was
	// more intuitive to require a board. If this is removed then the
	// application will get the feature of being able to search for all configs
	// that use V variant.
	if c.variant != common.DefaultString && c.board == common.DefaultString {
		return fmt.Errorf("variant cannot be provided without an accompanying board")
	}

	if c.startTime != common.DefaultInt64 {
		if c.startTime <= 0 {
			return fmt.Errorf("-start-time cannot be set to, or before, the epoch")
		}

		if c.nextNHours == common.DefaultHoursAhead || c.daily || c.weekly || c.fortnightly || c.searchAllConfigs {
			return fmt.Errorf("-start-time can only be used with -hours-ahead")
		}
	}

	// CTP request changes the format and thus this rule eliminates a large
	// amount of proto transformation that would otherwise be thrown away.
	if c.asCtpRequest && c.nameOnly {
		return fmt.Errorf("-ctp-request and -name-only cannot be provided together")
	}

	if c.nextNHours != common.DefaultHoursAhead {
		if c.nextNHours < 0 {
			return fmt.Errorf("hours ahead cannot be set to a negative value")
		}

		if c.day != common.DefaultInt64 || c.hour != common.DefaultInt64 {
			return fmt.Errorf("-day or -hour cannot be used with -hours-ahead")
		}
	}

	// Rules specific to respective top-level filters.
	if c.newBuild {
		if c.day != common.DefaultInt64 || c.hour != common.DefaultInt64 {
			return fmt.Errorf("-day nor -hour can be provided when searching for NEW_BUILD configs")
		}
	} else if c.daily {
		if c.day != common.DefaultInt64 && c.daily {
			return fmt.Errorf("the -day flag cannot be used when searching for daily configs")
		}
	} else if c.searchAllConfigs {
		if c.day != common.DefaultInt64 || c.hour != common.DefaultInt64 {
			return fmt.Errorf("if searching through all configs, day and hour cannot be passed in as daily, weekly, and fortnightly have different time parameters")
		}
	}

	return nil
}

// fetchLabConfigs fetches and ingests the lab configs. It will
// determine where to read the configs from based on the user provided flags.
func fetchLabConfigs(path string) (*configparser.LabConfigs, error) {
	var err error
	var labBytes []byte

	// If a file path was passed in for the Lab then parse that file. If not
	// then fetch the LabConfig from the ToT .cfg and ingest it in memory.
	if path != common.DefaultString {
		labBytes, err = common.ReadLocalFile(path)
		if err != nil {
			return nil, err
		}
	} else {
		labBytes, err = common.FetchFileFromURL(common.LabCfgURL)
		if err != nil {
			return nil, err
		}
	}

	labProto, err := configparser.BytesToLabProto(labBytes)
	if err != nil {
		return nil, err
	}

	labConfigs := configparser.IngestLabConfigs(labProto)

	return labConfigs, nil
}

// fetchSchedulerConfigs fetches and ingests the SuiteScheduler configs. It will
// determine where to read the configs from based on the user provided flags.
func fetchSchedulerConfigs(path string, labConfigs *configparser.LabConfigs) (*configparser.SuiteSchedulerConfigs, error) {
	var err error
	var schedulerBytes []byte

	// If a file path was passed in for the ScheduleConfigs then parse that file. If not
	// then fetch the SuiteSchedulerConfigs from the ToT .cfg and ingest it in memory.
	if path != common.DefaultString {
		schedulerBytes, err = common.ReadLocalFile(path)
		if err != nil {
			return nil, err
		}

	} else {
		schedulerBytes, err = common.FetchFileFromURL(common.SuiteSchedulerCfgURL)
		if err != nil {
			return nil, err
		}
	}

	// Convert from []byte to a usable object type.
	scheduleProto, err := configparser.BytesToSchedulerProto(schedulerBytes)
	if err != nil {
		return nil, err
	}

	// Ingest the configs into a data structure which easier and more efficient
	// to search.
	schedulerConfigs, err := configparser.IngestSuSchConfigs(scheduleProto.Configs, labConfigs)
	if err != nil {
		return nil, err
	}

	return schedulerConfigs, nil
}

// fetchConfigs reads the lab and scheduler configs into memory. If a local path
// is given then it will read from there otherwise it will read from the ToT
// configs.
func fetchConfigs(labPath, scheduleConfigsPath string) (*configparser.LabConfigs, *configparser.SuiteSchedulerConfigs, error) {
	labConfigs, err := fetchLabConfigs(labPath)
	if err != nil {
		return nil, nil, err
	}

	schedulerConfigs, err := fetchSchedulerConfigs(scheduleConfigsPath, labConfigs)
	if err != nil {
		return nil, nil, err
	}

	return labConfigs, schedulerConfigs, nil
}

// increaseTimeByAnHour increased the given hour by one and handles the day and
// week boundaries. We do not perform Day type validation because the user can
// omit the day return value should it not be needed.
func increaseTimeByAnHour(hour configparser.Hour, day configparser.Day, isFortnightly bool) (configparser.Hour, configparser.Day) {
	hour += 1

	// Handle the new day boundary.
	if hour > 23 {
		hour = 0
		day += 1
	}

	// Handle the new week boundary
	if isFortnightly {
		if day > 13 {

			day = 0
		}
	} else if day > 6 {
		day = 0
	}

	return hour, day
}

// FetchNextNHoursDailyConfigs returns all DAILY configs which will be triggered in
// the next N hours after the given start time.
func fetchNextNHoursDailyConfigs(startTime time.Time, hoursAhead int64, configMap *configparser.SuiteSchedulerConfigs) (CLIConfigList, error) {

	_, startHour := configparser.TimeToSuSchTime(startTime, false)

	// Validate that all input values fit within the expected bounds.
	if err := configparser.ValidateHoursAheadArgs(startHour, configparser.Day(common.DefaultInt64), hoursAhead, false); err != nil {
		return nil, err
	}

	configs := CLIConfigList{}

	// If hours ahead is zero then send the configs which we be triggered at the
	// startHour.
	if hoursAhead == 0 {
		configList, err := configMap.FetchDailyByHour(startHour)
		if err != nil {
			return nil, err
		}
		configs[startTime] = configList
		return configs, nil
	}

	for i := 0; i < int(hoursAhead); i++ {
		configList, err := configMap.FetchDailyByHour(startHour)
		if err != nil {
			return nil, err
		}
		configs[startTime] = configList

		// Push time the time one hour for the next iteration.
		startTime = startTime.Add(time.Hour)
		startHour, _ = increaseTimeByAnHour(startHour, configparser.Day(common.DefaultInt64), false)
	}

	return configs, nil
}

// fetchNextNHoursConfigsNotDaily returns all configs which will be triggered in
// the next N hours after the given start time. WEEKLY and FORTNIGHTLY share
// nearly all the same logic so this function is used as a base for both types
// of configs.
func fetchNextNHoursConfigsNotDaily(startTime time.Time, hoursAhead int64, isFortnightly bool, fetch func(configparser.Day, configparser.Hour) (configparser.ConfigList, error)) (CLIConfigList, error) {
	startDay, startHour := configparser.TimeToSuSchTime(startTime, false)
	// Validate that all input values fit within the expected bounds.
	if err := configparser.ValidateHoursAheadArgs(startHour, startDay, hoursAhead, false); err != nil {
		return nil, err
	}

	var err error
	configs := CLIConfigList{}
	for i := 0; i < int(hoursAhead); i++ {
		configs[startTime], err = fetch(startDay, startHour)
		if err != nil {
			return nil, err
		}

		// Push time the time one hour for the next iteration.
		startHour, startDay = increaseTimeByAnHour(startHour, startDay, isFortnightly)
	}

	return configs, nil
}

// sieveViaTopLevelFilter fetch the list of configurations specified by the
// given top-level filters. This list will be later filtered again by any
// relevant bottom-level filters.  These top-level filters apply to trigger
// mechanics of the configs.
func (c *configParserCommand) sieveViaTopLevelFilter(configs *configparser.SuiteSchedulerConfigs) (CLIConfigList, error) {
	filteredConfigs := CLIConfigList{}

	if c.newBuild {
		filteredConfigs[c.commandExecutionTime] = configs.FetchAllNewBuildConfigs()
	} else if c.daily {
		filteredConfigs[c.commandExecutionTime] = configs.FetchAllDailyConfigs()
	} else if c.weekly {
		filteredConfigs[c.commandExecutionTime] = configs.FetchAllWeeklyConfigs()
	} else if c.fortnightly {
		filteredConfigs[c.commandExecutionTime] = configs.FetchAllFortnightlyConfigs()
	} else if c.searchAllConfigs {
		filteredConfigs[c.commandExecutionTime] = configs.FetchAllConfigs()
	} else if c.nextNHours != common.DefaultHoursAhead {

		// Convert time.Time to a SuSch usable form.
		weeklyDay, weeklyHour := configparser.TimeToSuSchTime(c.commandExecutionTime, false)
		fnDay, fnHour := configparser.TimeToSuSchTime(c.commandExecutionTime, false)
		stdout.Printf("Looking ahead %d hours from a start time of %s %s. SuSch time: weekly (day:hour) %d:%d Fortnightly (day:hour) %d:%d\n", int(c.nextNHours.Hours()), c.commandExecutionTime.Weekday().String(), c.commandExecutionTime, weeklyDay, weeklyHour, fnDay, fnHour)

		// Daily
		// NOTE: This will include duplicate tasks if the a hours ahead value is
		// greater than 23 hours. They will be separated in different lists
		// keyed by their anticipated trigger datetime.
		hoursList, err := fetchNextNHoursDailyConfigs(c.commandExecutionTime, int64(c.nextNHours.Hours()), configs)
		if err != nil {
			return nil, err
		}
		for key, list := range hoursList {
			filteredConfigs[key] = list
		}

		// Weekly
		// NOTE: This will include duplicate tasks if the a hours ahead value is
		// greater than 7 days (168 hours). They will be separated in different lists
		// keyed by their anticipated trigger datetime.
		weeklyList, err := fetchNextNHoursConfigsNotDaily(c.commandExecutionTime, int64(c.nextNHours.Hours()), false, configs.FetchWeeklyByDayHour)
		if err != nil {
			return nil, err
		}

		// Add the configs to the return map.
		for key, list := range weeklyList {
			if _, ok := filteredConfigs[key]; !ok {
				filteredConfigs[key] = configparser.ConfigList{}
			}
			filteredConfigs[key] = list
		}

		// Fortnightly
		// NOTE: This will include duplicate tasks if the a hours ahead value is
		// greater than 14 days (336 hours). They will be separated in different lists
		// keyed by their anticipated trigger datetime.
		fortnightlyList, err := fetchNextNHoursConfigsNotDaily(c.commandExecutionTime, int64(c.nextNHours.Hours()), true, configs.FetchFortnightlyByDayHour)
		if err != nil {
			return nil, err
		}
		// Add the configs to the return map.
		for key, list := range fortnightlyList {
			if _, ok := filteredConfigs[key]; !ok {
				filteredConfigs[key] = configparser.ConfigList{}
			}
			filteredConfigs[key] = list
		}

	} else {
		// We shouldn't get here because validate() should cover this case but,
		// we'll add it as a sanity check.
		return nil, fmt.Errorf("the CLI encountered an unexpected set of instructions")
	}

	return filteredConfigs, nil
}

// sieveViaBottomLevelFilters further filters out any configs which do not match
// the user given inputs. These bottom-level filters apply to attributes within
// the config.
func (c *configParserCommand) sieveViaBottomLevelFilters(configs CLIConfigList, lab configparser.LabConfigs, suiteIndex *configparser.SuiteSchedulerConfigs) (CLIConfigList, error) {
	bottomLevelFilteredConfigs := CLIConfigList{}

	// Check if the filter is not default. If it is not then check it's status against
	// the filter. If it fails then remove it. Otherwise add it to a new further-filtered array.
	for datetimeKey, list := range configs {
		tempList := configparser.ConfigList{}
		for _, config := range list {
			// This bool will continually be updated throughout the loop to
			// determine if we should add the config to the further-filtered list.
			shouldAddConfig := true

			// If this is being searched, then it supersedes all other filters
			if c.configName != common.DefaultString {
				if c.configName == config.Name {
					bottomLevelFilteredConfigs[datetimeKey] = configparser.ConfigList{config}
					return bottomLevelFilteredConfigs, nil
				} else {
					continue
				}
			}

			if c.contains != common.DefaultString {
				shouldAddConfig = shouldAddConfig && strings.Contains(config.Name, c.contains)
			}

			if c.day != common.DefaultInt64 {
				if !common.IsTimedEvent(config) {
					return nil, fmt.Errorf("NEW_BUILD config %s was scanned when only timed event configs were requested", config.Name)
				} else {
					shouldAddConfig = shouldAddConfig && (c.day == int64(config.LaunchCriteria.Day))
				}

			}

			if c.hour != common.DefaultInt64 {
				if !common.IsTimedEvent(config) {
					return nil, fmt.Errorf("NEW_BUILD config %s was scanned when only timed event configs were requested", config.Name)
				} else {
					shouldAddConfig = shouldAddConfig && (c.hour == int64(config.LaunchCriteria.Hour))
				}
			}

			// Fetch the cached TargetOptions for the current config. If they do not
			// exist, an error has occurred during ingestion and the run should be terminated.
			targetOptions, err := suiteIndex.FetchConfigTargetOptions(config.Name)
			if err != nil {
				return nil, err
			}

			var target *configparser.TargetOption
			ok := false

			if c.board != common.DefaultString {
				target, ok = targetOptions[configparser.Board(c.board)]
				shouldAddConfig = shouldAddConfig && ok
			}

			if c.model != common.DefaultString && ok {
				shouldAddConfig = shouldAddConfig && common.HasString(c.model, target.Models)
			}

			if c.variant != common.DefaultString && ok {
				shouldAddConfig = shouldAddConfig && common.HasString(c.variant, target.Variants)
			}

			if shouldAddConfig {
				tempList = append(tempList, config)
			}
		}

		if len(tempList) != 0 {
			bottomLevelFilteredConfigs[datetimeKey] = tempList
		}

	}

	return bottomLevelFilteredConfigs, nil
}

// nameOnlyFormat strips out the names of the configs and returns the json
// formatted []byte.
func nameOnlyFormat(configs CLIConfigList, includeTimestamp bool) ([]byte, error) {
	// Given the formatting we can return one of each of the following types.
	timestampMap := map[time.Time][]string{}
	nameOnlyList := []string{}

	// Strip all values but the name of the config.
	for datetimeKey, configList := range configs {
		tempList := []string{}
		for _, config := range configList {
			tempList = append(tempList, config.Name)
		}

		// To save on space, only add to the object that we will be using
		// for the json return.
		if includeTimestamp {
			timestampMap[datetimeKey] = tempList
		} else {
			nameOnlyList = append(nameOnlyList, tempList...)
		}
	}

	var outputMap any
	if includeTimestamp {
		outputMap = timestampMap
	} else {
		outputMap = nameOnlyList
	}

	return json.MarshalIndent(outputMap, "", jsonMarshallIndent)
}

// ctpRequestFormat converts all found configs to their respective CTPRequests
// and returns it as a json formatted []byte.
func ctpRequestFormat(configs CLIConfigList, configTargetOptions map[string]configparser.TargetOptions, includeTimestamp bool) ([]byte, error) {
	timestampMap := map[time.Time]ctprequest.CTPRequests{}
	ctpRequestOnlyList := ctprequest.CTPRequests{}

	// Per datetime:config build all CTP requests that should be generated
	// from it's invocation.
	for datetimeKey, configList := range configs {
		for _, config := range configList {
			if _, ok := configTargetOptions[config.Name]; !ok {
				return nil, fmt.Errorf("config %s is not tracked in the target options cache", config.Name)
			}
			requests := ctprequest.BuildAllCTPRequests(config, configTargetOptions[config.Name])

			// To save on space, only add to the object that we will be using
			// for the json return.
			if includeTimestamp {
				timestampMap[datetimeKey] = append(timestampMap[datetimeKey], requests...)
			} else {
				ctpRequestOnlyList = append(ctpRequestOnlyList, requests...)
			}
		}
	}

	// Select which type we will be converting to a json.
	var outputMap any
	if includeTimestamp {
		outputMap = timestampMap
	} else {
		outputMap = ctpRequestOnlyList
	}

	return json.MarshalIndent(outputMap, "", jsonMarshallIndent)
}

func suiteSchedulerConfigFormat(configs CLIConfigList, includeTimestamp bool) ([]byte, error) {

	var outputMap any

	if includeTimestamp {
		outputMap = configs
	} else {
		configsOnlyList := configparser.ConfigList{}
		for _, configList := range configs {
			configsOnlyList = append(configsOnlyList, configList...)
		}
		outputMap = configsOnlyList
	}

	return json.MarshalIndent(outputMap, "", jsonMarshallIndent)
}

// formatOutput will strip or transform the configs according to the user given
// flags.
func (c *configParserCommand) formatOutput(configs CLIConfigList, configTargetOptions map[string]configparser.TargetOptions) ([]byte, error) {

	// Only include the timestamp in the output if we search for configs in the
	// next N hours.
	includeTimestamp := c.nextNHours != common.DefaultHoursAhead

	if c.nameOnly {
		return nameOnlyFormat(configs, includeTimestamp)
	} else if c.asCtpRequest {
		return ctpRequestFormat(configs, configTargetOptions, includeTimestamp)
	} else {
		return suiteSchedulerConfigFormat(configs, includeTimestamp)
	}
}

// outputResults will send the search results to the user defined output
// location.
func (c *configParserCommand) outputResults(path string, data []byte) error {
	if c.outputPath != common.DefaultString {
		err := common.WriteToFile(c.outputPath, data)
		if err != nil {
			return err
		}
		stdout.Printf("Results printed out to %s.\n", path)
		return nil
	}

	// We don't use the globally set stdout here so that the text is not
	// outputted with the line text prefix in the logger.
	fmt.Println(string(data))
	return nil
}

// Run is the "main" function of the subcommand. This
func (c *configParserCommand) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	// Validate that flags passed in are according to spec.
	if err := c.validate(); err != nil {
		stderr.Println(err)
		return 1
	}

	// Set the execution time of the application. This will only have an affect
	// on the `-hours-ahead` flag as that command is dependant on the "start
	// time" of the command run.
	if c.startTime != common.DefaultInt64 {
		c.commandExecutionTime = time.Unix(c.startTime, 0)
	} else {
		// NOTE: This assumes that SuSuch is ran in UTC. If this needs to be
		// transferred to PST we will perform that action here. Using UTC
		// everywhere is a much cleaner system overall though.
		c.commandExecutionTime = time.Now().UTC().Truncate(time.Hour)
	}

	// Fetch and ingest the configurations.
	labConfigs, schedulerConfigs, err := fetchConfigs(c.labCFGInputPath, c.configCFGInputPath)
	if err != nil {
		stderr.Println(err)
		return 1
	}

	// Retrieve the list of configs according to the top-level filter.
	filteredConfigs, err := c.sieveViaTopLevelFilter(schedulerConfigs)
	if err != nil {
		stderr.Println(err)
		return 1
	}

	// Filter out configs which don't match the bottom level filters.
	filteredConfigs, err = c.sieveViaBottomLevelFilters(filteredConfigs, *labConfigs, schedulerConfigs)
	if err != nil {
		stderr.Println(err)
		return 1
	}

	output, err := c.formatOutput(filteredConfigs, schedulerConfigs.FetchAllTargetOptions())
	if err != nil {
		stderr.Println(err)
		return 1
	}

	// return results. If a destination file is given then write the json text
	// to that file. Otherwise print the results to stdout.
	err = c.outputResults(c.outputPath, output)
	if err != nil {
		stderr.Println(err)
		return 1
	}

	return 0
}
