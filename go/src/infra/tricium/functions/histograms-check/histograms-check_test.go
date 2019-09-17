package main

import (
	"bufio"
	tricium "infra/tricium/api/v1"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func analyzeTestFile(t *testing.T, name string) []*tricium.Data_Comment {
	f, err := os.Open("test/src/" + name)
	if err != nil {
		t.Errorf("Failed to open %s: %v", name, err)
		return nil
	}
	defer f.Close()
	return analyzeFile(bufio.NewScanner(f), name)
}

func TestHistogramsCheck(t *testing.T) {
	Convey("Analyze XML file with no errors", t, func() {
		results := analyzeTestFile(t, "good.xml")
		So(results, ShouldBeNil)
	})

	Convey("Analyze XML file with error: only one owner", t, func() {
		results := analyzeTestFile(t, "one_owner.xml")
		So(results, ShouldResemble, []*tricium.Data_Comment{{
			Category:  "HistogramsXMLCheck/Owners",
			Message:   oneOwnerError,
			StartLine: 4,
			EndLine:   4,
			Path:      "one_owner.xml",
		}})
	})

	Convey("Analyze XML file with error: no owners", t, func() {
		results := analyzeTestFile(t, "no_owners.xml")
		So(results, ShouldResemble, []*tricium.Data_Comment{{
			Category:  "HistogramsXMLCheck/Owners",
			Message:   oneOwnerError,
			StartLine: 5,
			EndLine:   5,
			Path:      "no_owners.xml",
		}})
	})

	Convey("Analyze XML file with error: first owner is team", t, func() {
		results := analyzeTestFile(t, "first_team_owner.xml")
		So(results, ShouldResemble, []*tricium.Data_Comment{{
			Category:  "HistogramsXMLCheck/Owners",
			Message:   firstOwnerTeamError,
			StartLine: 4,
			EndLine:   4,
			Path:      "first_team_owner.xml",
		}})
	})

	Convey("Analyze XML file with multiple owner errors", t, func() {
		results := analyzeTestFile(t, "first_team_one_owner.xml")
		So(results, ShouldResemble, []*tricium.Data_Comment{
			{
				Category:  "HistogramsXMLCheck/Owners",
				Message:   oneOwnerError,
				StartLine: 4,
				EndLine:   4,
				Path:      "first_team_one_owner.xml",
			},
			{
				Category:  "HistogramsXMLCheck/Owners",
				Message:   firstOwnerTeamError,
				StartLine: 4,
				EndLine:   4,
				Path:      "first_team_one_owner.xml",
			},
		})
	})
}
