// Code generated by cproto. DO NOT EDIT.

package api

import "go.chromium.org/luci/grpc/discovery"

import "google.golang.org/protobuf/types/descriptorpb"

func init() {
	discovery.RegisterDescriptorSetCompressed(
		[]string{
			"chrome.fleet.labservice.LabService",
		},
		[]byte{31, 139,
			8, 0, 0, 0, 0, 0, 0, 255, 140, 83, 221, 78, 20, 77,
			16, 157, 158, 30, 246, 27, 138, 221, 47, 208, 223, 242, 129, 68,
			66, 133, 43, 18, 195, 172, 89, 188, 240, 22, 16, 127, 16, 69,
			89, 184, 241, 198, 244, 204, 212, 48, 29, 102, 187, 215, 238, 222,
			141, 155, 248, 0, 62, 135, 247, 190, 134, 207, 101, 122, 151, 141,
			63, 81, 227, 93, 157, 154, 170, 115, 78, 78, 77, 195, 199, 127,
			224, 158, 210, 149, 149, 189, 194, 26, 215, 43, 134, 101, 175, 145,
			185, 35, 59, 81, 5, 245, 228, 72, 245, 110, 235, 108, 100, 141,
			55, 98, 163, 168, 173, 25, 82, 86, 53, 68, 62, 251, 54, 186,
			251, 63, 116, 159, 144, 191, 188, 49, 135, 101, 105, 201, 185, 11,
			122, 55, 38, 231, 119, 31, 194, 250, 79, 125, 55, 50, 218, 145,
			216, 129, 21, 103, 138, 27, 242, 111, 71, 210, 215, 155, 12, 217,
			222, 242, 5, 204, 91, 175, 164, 175, 251, 31, 0, 206, 100, 62,
			152, 243, 11, 13, 157, 31, 120, 196, 126, 246, 27, 43, 217, 175,
			124, 108, 101, 127, 59, 62, 183, 119, 180, 243, 102, 251, 143, 185,
			156, 126, 73, 160, 37, 146, 36, 90, 103, 240, 153, 1, 107, 11,
			158, 68, 162, 255, 137, 225, 177, 25, 77, 173, 186, 174, 61, 246,
			239, 247, 251, 120, 89, 19, 30, 7, 105, 53, 30, 226, 249, 0,
			15, 199, 190, 54, 214, 101, 120, 216, 52, 56, 155, 115, 104, 41,
			48, 83, 153, 1, 94, 57, 66, 83, 161, 175, 149, 67, 103, 198,
			182, 32, 44, 76, 73, 168, 28, 94, 155, 9, 89, 77, 37, 230,
			83, 148, 120, 52, 120, 180, 239, 252, 180, 33, 108, 84, 65, 218,
			17, 250, 90, 122, 44, 164, 198, 156, 0, 43, 51, 214, 37, 42,
			141, 190, 38, 60, 123, 118, 124, 242, 114, 112, 130, 149, 106, 40,
			3, 72, 129, 197, 130, 183, 34, 12, 85, 42, 120, 26, 61, 128,
			101, 136, 211, 149, 121, 57, 132, 184, 21, 137, 164, 29, 253, 203,
			182, 36, 206, 111, 160, 38, 5, 225, 200, 154, 137, 42, 201, 161,
			212, 83, 172, 198, 186, 240, 202, 104, 217, 40, 63, 237, 41, 93,
			25, 244, 6, 115, 66, 122, 63, 50, 142, 176, 50, 118, 166, 253,
			56, 228, 141, 67, 169, 229, 53, 149, 128, 141, 204, 145, 244, 68,
			89, 163, 135, 164, 125, 6, 0, 192, 91, 17, 19, 188, 157, 10,
			88, 129, 164, 21, 197, 145, 224, 157, 248, 20, 218, 176, 20, 0,
			19, 188, 211, 250, 111, 129, 98, 193, 59, 221, 189, 5, 226, 130,
			119, 14, 158, 6, 243, 73, 36, 248, 106, 180, 19, 216, 146, 176,
			178, 154, 222, 5, 128, 56, 97, 34, 89, 11, 55, 10, 125, 198,
			4, 95, 75, 183, 33, 135, 36, 97, 65, 165, 27, 223, 217, 186,
			194, 239, 126, 195, 16, 115, 48, 45, 115, 103, 154, 177, 167, 89,
			98, 110, 234, 60, 13, 113, 246, 221, 27, 148, 248, 98, 58, 120,
			125, 118, 187, 22, 58, 97, 227, 242, 249, 57, 96, 41, 189, 204,
			165, 11, 25, 183, 97, 41, 104, 44, 5, 145, 116, 129, 152, 224,
			221, 229, 238, 2, 113, 193, 187, 27, 155, 121, 107, 246, 178, 14,
			190, 6, 0, 0, 255, 255, 157, 182, 138, 198, 139, 3, 0, 0,
		},
	)
}

// FileDescriptorSet returns a descriptor set for this proto package, which
// includes all defined services, and all transitive dependencies.
//
// Will not return nil.
//
// Do NOT modify the returned descriptor.
func FileDescriptorSet() *descriptorpb.FileDescriptorSet {
	// We just need ONE of the service names to look up the FileDescriptorSet.
	ret, err := discovery.GetDescriptorSet("chrome.fleet.labservice.LabService")
	if err != nil {
		panic(err)
	}
	return ret
}
