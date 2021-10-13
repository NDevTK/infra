package re

import (
	"testing"

	cpb "infra/appengine/weetbix/internal/clustering/proto"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAlgorithm(t *testing.T) {
	Convey(`Does not cluster test result without failure reason`, t, func() {
		re := &Algorithm{}
		id := re.Cluster(&cpb.Failure{})
		So(id, ShouldBeNil)
	})
	Convey(`ID of appropriate length`, t, func() {
		re := &Algorithm{}
		id := re.Cluster(&cpb.Failure{
			FailureReason: &cpb.FailureReason{
				PrimaryErrorMessage: "abcd this is a test failure message",
			},
		})
		// IDs may be 16 bytes at most.
		So(len(id), ShouldBeGreaterThan, 0)
		So(len(id), ShouldBeLessThanOrEqualTo, 16)
	})
	Convey(`Same ID for same cluster with different numbers`, t, func() {
		re := &Algorithm{}
		id1 := re.Cluster(&cpb.Failure{
			FailureReason: &cpb.FailureReason{
				PrimaryErrorMessage: "Null pointer exception at ip 0x45637271",
			},
		})
		id2 := re.Cluster(&cpb.Failure{
			FailureReason: &cpb.FailureReason{
				PrimaryErrorMessage: "Null pointer exception at ip 0x12345678",
			},
		})
		So(id2, ShouldResemble, id1)
	})
	Convey(`Different ID for different clusters`, t, func() {
		re := &Algorithm{}
		id1 := re.Cluster(&cpb.Failure{
			FailureReason: &cpb.FailureReason{
				PrimaryErrorMessage: "Exception in TestMethod",
			},
		})
		id2 := re.Cluster(&cpb.Failure{
			FailureReason: &cpb.FailureReason{
				PrimaryErrorMessage: "Exception in MethodUnderTest",
			},
		})
		So(id2, ShouldNotResemble, id1)
	})
}
