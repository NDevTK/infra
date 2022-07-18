package cros

import (
	"context"
	"testing"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/tlw"
)

var isSmartHubExpectedExecTests = []struct {
	testName    string
	chromeos    *tlw.ChromeOS
	expectedErr bool
}{
	{
		"SmartHub is specified",
		&tlw.ChromeOS{
			Servo: &tlw.ServoHost{
				SmartUsbhubPresent: true,
			},
		},
		false,
	},
	{
		"SmartHub is not",
		nil,
		true,
	},
}

func TestIsSmartHubExpectedExec(t *testing.T) {
	t.Parallel()
	for _, tt := range isSmartHubExpectedExecTests {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			info := &execs.ExecInfo{
				RunArgs: &execs.RunArgs{
					DUT: &tlw.Dut{
						Chromeos: tt.chromeos,
					},
				},
			}
			err := isSmartHubExpectedExec(ctx, info)
			if err == nil && tt.expectedErr {
				t.Errorf("%q -> error expected but not received", tt.testName)
			}
			if err != nil && !tt.expectedErr {
				t.Errorf("%q -> received error even not expected %v", tt.testName, err)
			}
		})
	}
}
