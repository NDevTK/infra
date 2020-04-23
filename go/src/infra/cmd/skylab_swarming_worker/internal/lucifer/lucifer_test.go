package lucifer

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConvertActionsTaskArgs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		task     string
		inDir    string
		inArgs   ActionsTaskArgs
		expected []string
	}{
		{
			"test1",
			"som_path",
			ActionsTaskArgs{
				Task:     "task1",
				TaskArgs: TaskArgs{},
				Host:     "host1",
			},
			[]string{
				"task1",
				"-autotestdir", "som_path",
				"-abortsock", "",
				"-gcp-project", "",
				"-resultsdir", "",
				"-host", "host1",
			},
		},
		{
			"test2",
			"my_path",
			ActionsTaskArgs{
				Task: "task2",
				TaskArgs: TaskArgs{
					AbortSock:  "AbortSock1",
					GCPProject: "GCPProject1",
					ResultsDir: "ResultsDir1",
					LogDogFile: "LogDogFile1",
				},
				Host:    "host1",
				Actions: "action1,action2, my super space haha",
			},
			[]string{
				"task2",
				"-autotestdir", "my_path",
				"-abortsock", "AbortSock1",
				"-gcp-project", "GCPProject1",
				"-resultsdir", "ResultsDir1",
				"-logdog-file", "LogDogFile1",
				"-host", "host1",
				"-actions", "action1,action2, my super space haha",
				"-gcp-project", "GCPProject1",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.task, func(t *testing.T) {
			t.Parallel()
			output := convertActionsTaskArgs(tc.inArgs, tc.inDir)
			diff := cmp.Diff(tc.expected, output)
			if diff != "" {
				t.Errorf("Input task was %s - check was incorrect, dif: %v", tc.task, diff)
			}
		})
	}
}
