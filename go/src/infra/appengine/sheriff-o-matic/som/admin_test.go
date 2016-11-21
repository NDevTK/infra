// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package som

import (
	"testing"

	"github.com/luci/gae/service/datastore"
	"github.com/luci/luci-go/appengine/gaetesting"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRenderSettingsPage(t *testing.T) {
	t.Parallel()

	Convey("render settings", t, func() {
		c := gaetesting.TestingContext()
		s := settingsUIPage{}

		Convey("Title", func() {
			title, err := settingsUIPage.Title(s, c)
			So(err, ShouldBeNil)
			So(title, ShouldEqual, "Admin SOM settings")
		})

		tree := &Tree{
			Name:          "oak",
			DisplayName:   "Great Oaakk",
			BugQueueLabel: "test",
			AlertStreams:  []string{"hello", "world"},
			HelpLink:      "http://google.com/",
		}

		So(datastore.Put(c, tree), ShouldBeNil)
		datastore.GetTestable(c).CatchupIndexes()

		Convey("Fields", func() {
			fields, err := settingsUIPage.Fields(s, c)
			So(err, ShouldBeNil)
			So(len(fields), ShouldEqual, 4)
		})

		Convey("ReadSettings", func() {
			settings, err := settingsUIPage.ReadSettings(s, c)
			So(err, ShouldBeNil)
			So(len(settings), ShouldEqual, 4)
			So(settings["Trees"], ShouldEqual, "oak:Great Oaakk")
			So(settings["BugQueueLabels"], ShouldEqual, "oak:test")
			So(settings["AlertStreams-oak"], ShouldEqual, "hello,world")
			So(settings["HelpLink-oak"], ShouldEqual, "http://google.com/")
		})
	})
}

func TestWriteAllValues(t *testing.T) {
	t.Parallel()

	Convey("write settings", t, func() {
		c := gaetesting.TestingContext()

		Convey("writeTrees", func() {
			Convey("basic", func() {
				values := map[string]string{
					"Trees": "foo",
				}
				err := writeAllValues(c, values)
				So(err, ShouldBeNil)
				datastore.GetTestable(c).CatchupIndexes()

				t := &Tree{
					Name: "foo",
				}
				So(datastore.Get(c, t), ShouldBeNil)
				So(t.DisplayName, ShouldEqual, "Foo")
			})

			tree := &Tree{
				Name:        "oak",
				DisplayName: "Great Oaakk",
			}

			So(datastore.Put(c, tree), ShouldBeNil)
			datastore.GetTestable(c).CatchupIndexes()

			Convey("overwrite tree", func() {
				values := map[string]string{
					"Trees": "oak",
				}
				err := writeAllValues(c, values)
				So(err, ShouldBeNil)
				datastore.GetTestable(c).CatchupIndexes()

				So(datastore.Get(c, tree), ShouldBeNil)
				So(tree.DisplayName, ShouldEqual, "Oak")
			})

			Convey("overwrite tree with new display name", func() {
				values := map[string]string{
					"Trees": "oak:Oaakk",
				}
				err := writeAllValues(c, values)
				So(err, ShouldBeNil)
				datastore.GetTestable(c).CatchupIndexes()

				So(datastore.Get(c, tree), ShouldBeNil)
				So(tree.DisplayName, ShouldEqual, "Oaakk")
			})
		})

		Convey("update AlertStreams", func() {
			tree := &Tree{
				Name:        "oak",
				DisplayName: "Oak",
			}

			So(datastore.Put(c, tree), ShouldBeNil)
			datastore.GetTestable(c).CatchupIndexes()

			Convey("basic", func() {
				values := map[string]string{
					"AlertStreams-oak": "thing,hello",
				}
				err := writeAllValues(c, values)
				So(err, ShouldBeNil)
				datastore.GetTestable(c).CatchupIndexes()

				So(datastore.Get(c, tree), ShouldBeNil)
				So(tree.DisplayName, ShouldEqual, "Oak")
				So(tree.AlertStreams, ShouldResemble, []string{"thing", "hello"})
			})

			Convey("delete", func() {
				values := map[string]string{
					"AlertStreams-oak": "",
				}
				err := writeAllValues(c, values)
				So(err, ShouldBeNil)
				datastore.GetTestable(c).CatchupIndexes()

				So(datastore.Get(c, tree), ShouldBeNil)
				So(tree.DisplayName, ShouldEqual, "Oak")
				So(tree.AlertStreams, ShouldResemble, []string(nil))
			})
		})

		Convey("splitBugQueueLabels", func() {
			Convey("single", func() {
				labelMap, err := splitBugQueueLabels(c, "oak:thing")
				So(err, ShouldBeNil)

				So(labelMap["oak"], ShouldEqual, "thing")
			})

			Convey("mutiple", func() {
				labelMap, err := splitBugQueueLabels(c, "oak:thing,maple:syrup,haha:haha")
				So(err, ShouldBeNil)

				So(labelMap["oak"], ShouldEqual, "thing")
				So(labelMap["maple"], ShouldEqual, "syrup")
				So(labelMap["haha"], ShouldEqual, "haha")
			})
		})

		Convey("update BugQueueLabel", func() {
			tree := &Tree{
				Name:          "oak",
				DisplayName:   "Oak",
				BugQueueLabel: "test",
			}

			So(datastore.Put(c, tree), ShouldBeNil)
			datastore.GetTestable(c).CatchupIndexes()

			Convey("basic", func() {
				values := map[string]string{
					"BugQueueLabels": "oak:thing",
				}
				err := writeAllValues(c, values)
				So(err, ShouldBeNil)
				datastore.GetTestable(c).CatchupIndexes()

				So(datastore.Get(c, tree), ShouldBeNil)
				So(tree.Name, ShouldEqual, "oak")
				So(tree.DisplayName, ShouldEqual, "Oak")
				So(tree.BugQueueLabel, ShouldEqual, "thing")
			})

			Convey("remove label", func() {
				values := map[string]string{
					"BugQueueLabels": "oak:",
				}
				err := writeAllValues(c, values)
				So(err, ShouldBeNil)
				datastore.GetTestable(c).CatchupIndexes()

				So(datastore.Get(c, tree), ShouldBeNil)
				So(tree.Name, ShouldEqual, "oak")
				So(tree.DisplayName, ShouldEqual, "Oak")
				So(tree.BugQueueLabel, ShouldEqual, "")
			})
		})

		Convey("update HelpLink", func() {
			tree := &Tree{
				Name:          "oak",
				DisplayName:   "Oak",
				HelpLink:      "Mwuhaha",
				BugQueueLabel: "ShouldNotChange",
			}

			So(datastore.Put(c, tree), ShouldBeNil)
			datastore.GetTestable(c).CatchupIndexes()

			Convey("basic", func() {
				values := map[string]string{
					"HelpLink-oak": "http://google.com",
				}
				err := writeAllValues(c, values)
				So(err, ShouldBeNil)
				datastore.GetTestable(c).CatchupIndexes()

				So(datastore.Get(c, tree), ShouldBeNil)
				So(tree.DisplayName, ShouldEqual, "Oak")
				So(tree.HelpLink, ShouldEqual, "http://google.com")
				So(tree.BugQueueLabel, ShouldEqual, "ShouldNotChange")
			})

			Convey("delete", func() {
				values := map[string]string{
					"HelpLink-oak": "",
				}
				err := writeAllValues(c, values)
				So(err, ShouldBeNil)
				datastore.GetTestable(c).CatchupIndexes()

				So(datastore.Get(c, tree), ShouldBeNil)
				So(tree.DisplayName, ShouldEqual, "Oak")
				So(tree.HelpLink, ShouldEqual, "")
				So(tree.BugQueueLabel, ShouldEqual, "ShouldNotChange")
			})
		})
	})
}
