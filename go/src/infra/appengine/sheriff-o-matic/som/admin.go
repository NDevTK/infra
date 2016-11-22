// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package som

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	"github.com/luci/gae/service/datastore"
	"github.com/luci/luci-go/server/settings"
)

type settingsUIPage struct {
	settings.BaseUIPage
}

func (settingsUIPage) Title(c context.Context) (string, error) {
	return "Admin SOM settings", nil
}

func (settingsUIPage) Fields(c context.Context) ([]settings.UIField, error) {
	fields := []settings.UIField{
		{
			ID:    "Trees",
			Title: "Trees in SOM",
			Type:  settings.UIFieldText,
			Help:  "Trees listed in SOM. Comma separated values. treeA,treeB",
		},
		{
			ID:    "BugQueueLabels",
			Title: "Bug Queue Labels",
			Type:  settings.UIFieldText,
			Help:  "Bug queue label for each tree. treeA:queueA,treeB:queueB",
		},
	}

	q := datastore.NewQuery("Tree")
	trees := []*Tree{}
	datastore.GetAll(c, q, &trees)

	// Add settings fields for specific trees
	for _, t := range trees {
		fields = append(fields, settings.UIField{
			ID:    fmt.Sprintf("AlertStreams-%s", t.Name),
			Title: fmt.Sprintf("%s Alert Streams", t.DisplayName),
			Type:  settings.UIFieldText,
			Help:  "Alert streams for this tree. Defaults to tree name if blank. streamA,streamB",
		})
		fields = append(fields, settings.UIField{
			ID:    fmt.Sprintf("HelpLink-%s", t.Name),
			Title: fmt.Sprintf("%s Help Link", t.DisplayName),
			Type:  settings.UIFieldText,
			Help:  "A link to help documentation for this tree. ie. a playbook",
		})
	}

	return fields, nil
}

func (settingsUIPage) ReadSettings(c context.Context) (map[string]string, error) {
	q := datastore.NewQuery("Tree")
	results := []*Tree{}
	datastore.GetAll(c, q, &results)
	trees := make([]string, len(results))
	queues := make([]string, len(results))

	values := make(map[string]string)

	for i, t := range results {
		trees[i] = fmt.Sprintf("%s:%s", t.Name, t.DisplayName)
		queues[i] = fmt.Sprintf("%s:%s", t.Name, t.BugQueueLabel)

		values[fmt.Sprintf("AlertStreams-%s", t.Name)] = strings.Join(t.AlertStreams, ",")
		values[fmt.Sprintf("HelpLink-%s", t.Name)] = t.HelpLink
	}

	values["Trees"] = strings.Join(trees, ",")
	values["BugQueueLabels"] = strings.Join(queues, ",")

	return values, nil
}

func deleteAllTrees(c context.Context) error {
	q := datastore.NewQuery("Tree")
	trees := []*Tree{}
	err := datastore.GetAll(c, q, &trees)
	if err != nil {
		return err
	}

	err = datastore.Delete(c, trees)

	if err != nil {
		return err
	}
	return nil
}

func initializeTrees(c context.Context, treeStr string) ([]*Tree, error) {
	toMake := strings.Split(treeStr, ",")
	trees := make([]*Tree, len(toMake))
	for i, it := range toMake {
		it = strings.TrimSpace(it)
		if len(it) == 0 {
			continue
		}

		nameParts := strings.Split(it, ":")
		name := nameParts[0]
		displayName := strings.Replace(strings.Title(name), "_", " ", -1)
		if len(nameParts) == 2 {
			displayName = nameParts[1]
		}
		trees[i] = &Tree{
			Name:        name,
			DisplayName: displayName,
		}
	}
	return trees, nil
}

// bugQueueLabels format is treeA:queueA,treeB:queueB
func splitBugQueueLabels(c context.Context, bugQueueLabels string) (map[string]string, error) {
	result := make(map[string]string)
	queueLabels := strings.Split(bugQueueLabels, ",")
	for _, label := range queueLabels {
		split := strings.Split(label, ":")
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid bugQueueLabels: %q", bugQueueLabels)
		}
		result[split[0]] = split[1]
	}

	return result, nil
}

func writeAllValues(c context.Context, values map[string]string) error {
	trees := []*Tree{}
	if treeStr, ok := values["Trees"]; ok {
		// Always replace the existing list of trees. Otherwise there's no "delete"
		// capability.
		err := deleteAllTrees(c)
		if err != nil {
			return err
		}

		trees, err = initializeTrees(c, treeStr)
		if err != nil {
			return err
		}
	} else {
		q := datastore.NewQuery("Tree")
		trees = []*Tree{}
		datastore.GetAll(c, q, &trees)
	}

	labels := make(map[string]string)
	if bugQueueLabels, ok := values["BugQueueLabels"]; ok && bugQueueLabels != "" {
		// Split the bug queue labels first and write later to help minimize writes
		l, err := splitBugQueueLabels(c, bugQueueLabels)
		if err != nil {
			return err
		}
		labels = l
	}

	for _, t := range trees {
		if bugQueueLabel, ok := labels[t.Name]; ok {
			t.BugQueueLabel = bugQueueLabel
		}

		if alertStreams, ok := values[fmt.Sprintf("AlertStreams-%s", t.Name)]; ok {
			if alertStreams != "" {
				t.AlertStreams = strings.Split(alertStreams, ",")
			} else {
				t.AlertStreams = []string(nil)
			}
		}

		if helpLink, ok := values[fmt.Sprintf("HelpLink-%s", t.Name)]; ok {
			t.HelpLink = helpLink
		}

		// Try to do only write per tree each save.
		if err := datastore.Put(c, t); err != nil {
			return err
		}
	}
	return nil
}

func (settingsUIPage) WriteSettings(c context.Context, values map[string]string, who, why string) error {
	// Putting the write logic in a function outside of WriteSettings makes unit testing easier.
	return writeAllValues(c, values)
}
