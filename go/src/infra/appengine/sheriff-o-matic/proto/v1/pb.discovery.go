// Code generated by cproto. DO NOT EDIT.

package sompb

import "go.chromium.org/luci/grpc/discovery"

import "google.golang.org/protobuf/types/descriptorpb"

func init() {
	discovery.RegisterDescriptorSetCompressed(
		[]string{
			"infra.appengine.som.v1.Alerts",
		},
		[]byte{31, 139,
			8, 0, 0, 0, 0, 0, 0, 255, 140, 84, 207, 111, 212, 198,
			23, 247, 120, 156, 213, 230, 45, 36, 48, 124, 19, 146, 229, 27,
			120, 74, 161, 161, 37, 120, 55, 65, 72, 85, 57, 133, 52, 149,
			64, 168, 66, 129, 170, 82, 47, 196, 235, 125, 27, 79, 177, 61,
			238, 204, 120, 33, 84, 61, 244, 216, 83, 47, 253, 223, 122, 235,
			173, 151, 170, 255, 67, 171, 25, 219, 73, 40, 84, 226, 100, 127,
			102, 222, 143, 207, 251, 188, 55, 15, 254, 94, 134, 251, 178, 156,
			233, 100, 148, 84, 21, 149, 199, 178, 164, 145, 201, 72, 203, 217,
			236, 174, 186, 91, 36, 86, 166, 163, 74, 43, 171, 70, 243, 157,
			81, 146, 147, 182, 38, 246, 88, 172, 122, 183, 248, 212, 45, 54,
			170, 136, 231, 59, 155, 199, 112, 249, 137, 52, 118, 207, 219, 30,
			210, 247, 53, 25, 43, 86, 161, 87, 37, 154, 74, 187, 198, 144,
			221, 94, 60, 108, 145, 184, 6, 139, 85, 114, 76, 47, 140, 124,
			67, 107, 33, 178, 219, 11, 135, 125, 119, 240, 76, 190, 33, 177,
			1, 224, 47, 173, 122, 73, 229, 26, 247, 142, 222, 252, 185, 59,
			216, 52, 32, 206, 39, 50, 149, 42, 13, 137, 251, 208, 107, 104,
			174, 49, 228, 183, 7, 187, 27, 241, 251, 121, 198, 222, 239, 176,
			53, 22, 31, 195, 114, 73, 175, 237, 139, 115, 9, 67, 159, 240,
			162, 59, 126, 122, 154, 244, 51, 88, 240, 142, 226, 18, 240, 151,
			116, 210, 150, 227, 126, 29, 93, 31, 236, 197, 119, 70, 117, 222,
			139, 254, 228, 177, 81, 229, 174, 130, 94, 67, 85, 16, 192, 25,
			113, 241, 201, 127, 17, 124, 71, 197, 225, 167, 31, 98, 218, 232,
			176, 25, 60, 28, 127, 27, 127, 104, 103, 31, 24, 85, 84, 147,
			199, 127, 14, 160, 39, 162, 40, 216, 97, 240, 43, 3, 118, 65,
			240, 40, 16, 187, 63, 51, 220, 87, 213, 137, 150, 199, 153, 197,
			221, 241, 238, 61, 124, 158, 17, 238, 103, 90, 21, 178, 46, 112,
			175, 182, 153, 210, 6, 240, 107, 67, 168, 102, 104, 51, 105, 208,
			168, 90, 167, 132, 169, 154, 18, 74, 131, 199, 106, 78, 186, 164,
			41, 78, 78, 48, 193, 135, 207, 190, 184, 107, 236, 73, 78, 152,
			203, 148, 74, 67, 104, 179, 196, 98, 154, 148, 56, 33, 192, 153,
			170, 203, 41, 202, 18, 109, 70, 248, 228, 209, 254, 193, 87, 207,
			14, 112, 38, 115, 138, 1, 250, 192, 66, 193, 123, 193, 13, 247,
			215, 23, 188, 31, 28, 192, 34, 132, 253, 65, 243, 91, 65, 216,
			11, 68, 180, 20, 92, 98, 187, 83, 124, 170, 213, 92, 78, 201,
			96, 65, 54, 83, 83, 131, 86, 161, 38, 171, 37, 205, 9, 219,
			89, 6, 192, 189, 178, 1, 152, 42, 173, 189, 124, 141, 105, 130,
			198, 82, 133, 179, 68, 230, 181, 38, 84, 165, 59, 145, 229, 113,
			78, 56, 169, 101, 62, 37, 29, 3, 0, 240, 94, 192, 4, 95,
			234, 47, 193, 0, 162, 94, 16, 6, 130, 47, 71, 7, 112, 1,
			22, 28, 96, 130, 47, 247, 69, 135, 66, 193, 151, 175, 220, 236,
			16, 23, 124, 121, 180, 7, 0, 97, 20, 136, 72, 4, 31, 49,
			23, 46, 114, 62, 162, 191, 14, 79, 33, 138, 124, 184, 149, 104,
			117, 184, 239, 69, 47, 147, 162, 213, 152, 208, 106, 162, 247, 84,
			132, 51, 165, 1, 191, 84, 186, 72, 236, 231, 222, 200, 140, 126,
			112, 159, 31, 193, 229, 117, 17, 23, 92, 200, 83, 196, 4, 95,
			25, 92, 238, 16, 23, 124, 229, 127, 43, 240, 7, 243, 201, 153,
			224, 215, 162, 181, 225, 111, 204, 103, 47, 146, 215, 178, 168, 11,
			44, 235, 98, 66, 218, 241, 104, 83, 54, 44, 106, 93, 198, 232,
			13, 13, 233, 185, 76, 157, 195, 73, 123, 129, 51, 122, 69, 218,
			245, 185, 132, 102, 66, 230, 73, 94, 83, 12, 248, 104, 134, 117,
			105, 42, 74, 229, 76, 210, 116, 27, 19, 139, 133, 50, 22, 119,
			198, 227, 113, 23, 255, 149, 204, 115, 156, 80, 27, 139, 166, 49,
			188, 197, 199, 135, 114, 99, 230, 124, 30, 52, 208, 96, 50, 81,
			115, 106, 194, 116, 254, 169, 34, 157, 210, 212, 241, 117, 231, 113,
			167, 1, 91, 112, 117, 46, 118, 200, 85, 13, 87, 58, 196, 5,
			191, 182, 122, 21, 254, 106, 20, 9, 5, 223, 140, 134, 195, 223,
			25, 238, 161, 219, 21, 232, 119, 197, 54, 106, 74, 73, 206, 105,
			138, 51, 173, 10, 76, 176, 210, 52, 151, 170, 54, 120, 116, 246,
			54, 143, 48, 77, 242, 60, 134, 110, 48, 27, 33, 206, 183, 208,
			245, 213, 212, 19, 227, 222, 123, 105, 125, 2, 55, 161, 223, 100,
			84, 58, 32, 203, 196, 202, 242, 120, 27, 147, 60, 71, 101, 51,
			210, 88, 37, 58, 41, 200, 146, 54, 88, 53, 81, 125, 121, 111,
			101, 45, 106, 99, 177, 72, 108, 154, 129, 207, 224, 88, 52, 79,
			238, 204, 37, 163, 115, 229, 156, 10, 19, 46, 184, 114, 79, 17,
			19, 124, 115, 176, 210, 33, 46, 248, 230, 218, 186, 31, 95, 38,
			162, 91, 193, 118, 51, 190, 78, 190, 91, 253, 33, 220, 132, 40,
			98, 110, 124, 183, 162, 235, 195, 171, 190, 97, 186, 217, 99, 52,
			61, 123, 125, 46, 22, 11, 131, 200, 153, 93, 232, 80, 79, 240,
			173, 139, 162, 67, 76, 240, 173, 43, 235, 29, 226, 130, 111, 253,
			127, 3, 126, 97, 62, 60, 19, 252, 78, 116, 99, 248, 147, 107,
			71, 219, 137, 87, 153, 76, 179, 118, 153, 160, 113, 50, 38, 6,
			143, 206, 246, 250, 209, 59, 138, 187, 13, 223, 106, 237, 198, 209,
			119, 101, 38, 41, 159, 186, 153, 82, 133, 180, 214, 13, 166, 147,
			155, 48, 209, 132, 165, 250, 119, 147, 206, 42, 113, 163, 116, 39,
			58, 69, 142, 223, 96, 216, 33, 46, 248, 157, 141, 235, 94, 177,
			80, 68, 177, 219, 179, 78, 49, 87, 69, 220, 191, 232, 246, 71,
			20, 58, 197, 70, 81, 51, 125, 161, 127, 171, 163, 54, 92, 232,
			165, 24, 13, 150, 58, 196, 5, 31, 93, 22, 173, 27, 19, 124,
			28, 13, 219, 43, 199, 98, 124, 234, 230, 88, 140, 219, 190, 133,
			158, 197, 120, 109, 125, 210, 243, 203, 255, 222, 63, 1, 0, 0,
			255, 255, 152, 55, 217, 155, 2, 8, 0, 0},
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
	ret, err := discovery.GetDescriptorSet("infra.appengine.som.v1.Alerts")
	if err != nil {
		panic(err)
	}
	return ret
}
