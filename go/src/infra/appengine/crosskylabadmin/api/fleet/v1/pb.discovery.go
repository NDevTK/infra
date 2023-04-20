// Code generated by cproto. DO NOT EDIT.

package fleet

import "go.chromium.org/luci/grpc/discovery"

import "google.golang.org/protobuf/types/descriptorpb"

func init() {
	discovery.RegisterDescriptorSetCompressed(
		[]string{
			"crosskylabadmin.fleet.Inventory", "crosskylabadmin.fleet.Tracker",
		},
		[]byte{31, 139,
			8, 0, 0, 0, 0, 0, 0, 255, 220, 58, 75, 111, 28, 201,
			121, 93, 93, 61, 195, 102, 81, 18, 135, 197, 167, 134, 20, 89,
			162, 184, 50, 165, 37, 135, 187, 212, 131, 15, 173, 180, 226, 75,
			43, 74, 92, 145, 154, 161, 180, 214, 26, 107, 169, 57, 93, 156,
			233, 229, 76, 247, 184, 31, 164, 184, 202, 194, 137, 99, 3, 70,
			14, 137, 147, 75, 144, 131, 145, 179, 29, 248, 20, 4, 8, 16,
			32, 200, 15, 72, 128, 0, 49, 144, 92, 114, 200, 33, 183, 156,
			114, 75, 14, 65, 240, 213, 163, 103, 134, 230, 80, 90, 199, 151,
			152, 23, 206, 87, 85, 223, 163, 190, 87, 215, 247, 85, 145, 127,
			30, 34, 75, 158, 191, 31, 58, 115, 78, 163, 193, 253, 138, 231,
			243, 185, 114, 24, 68, 209, 193, 113, 205, 217, 115, 220, 186, 231,
			207, 57, 13, 111, 110, 191, 198, 121, 60, 119, 248, 225, 92, 57,
			168, 215, 3, 191, 208, 8, 131, 56, 160, 131, 39, 150, 22, 196,
			178, 201, 47, 73, 207, 106, 16, 151, 120, 141, 151, 227, 32, 164,
			131, 36, 235, 38, 241, 75, 207, 29, 65, 12, 77, 119, 23, 51,
			110, 18, 111, 186, 116, 157, 16, 215, 171, 115, 63, 242, 2, 63,
			26, 49, 25, 154, 238, 153, 159, 42, 156, 74, 177, 176, 26, 196,
			235, 233, 218, 98, 11, 222, 228, 115, 114, 190, 109, 146, 14, 144,
			76, 35, 8, 106, 209, 8, 98, 24, 152, 9, 0, 70, 235, 129,
			203, 107, 130, 79, 119, 81, 2, 244, 34, 177, 65, 50, 223, 169,
			243, 17, 44, 38, 186, 220, 36, 126, 226, 212, 249, 228, 107, 146,
			45, 241, 240, 144, 135, 52, 79, 236, 106, 16, 197, 98, 145, 220,
			64, 10, 211, 5, 146, 9, 131, 26, 7, 241, 241, 244, 133, 249,
			203, 29, 196, 151, 148, 138, 65, 141, 23, 229, 122, 205, 57, 241,
			220, 104, 4, 11, 65, 129, 243, 51, 207, 141, 174, 255, 5, 34,
			164, 137, 64, 115, 228, 92, 113, 123, 107, 227, 229, 230, 147, 231,
			43, 91, 155, 235, 57, 131, 14, 146, 62, 49, 178, 86, 92, 41,
			61, 124, 89, 218, 40, 62, 223, 40, 230, 76, 74, 201, 5, 49,
			188, 190, 241, 92, 141, 101, 232, 8, 25, 16, 99, 165, 103, 155,
			187, 27, 47, 75, 107, 15, 55, 214, 159, 109, 109, 20, 115, 231,
			82, 34, 165, 199, 47, 182, 86, 86, 95, 174, 23, 183, 159, 108,
			228, 206, 211, 81, 50, 220, 58, 44, 241, 118, 138, 219, 223, 126,
			145, 187, 144, 114, 40, 238, 124, 170, 56, 244, 174, 222, 252, 124,
			254, 155, 248, 207, 29, 241, 227, 209, 191, 244, 145, 44, 181, 44,
			99, 13, 145, 191, 66, 4, 157, 163, 216, 50, 232, 252, 207, 17,
			91, 11, 26, 199, 161, 87, 169, 198, 108, 254, 131, 15, 23, 217,
			110, 149, 179, 173, 103, 107, 155, 108, 37, 137, 171, 65, 24, 21,
			216, 74, 173, 198, 196, 130, 136, 133, 60, 2, 53, 185, 5, 194,
			158, 69, 156, 5, 251, 44, 174, 122, 17, 139, 130, 36, 44, 115,
			86, 14, 92, 206, 188, 136, 85, 130, 67, 30, 250, 220, 101, 137,
			239, 242, 144, 197, 85, 206, 86, 26, 78, 25, 8, 123, 101, 238,
			71, 124, 134, 61, 231, 33, 248, 14, 155, 47, 124, 64, 88, 92,
			117, 98, 86, 118, 124, 182, 199, 217, 126, 144, 248, 46, 243, 124,
			129, 181, 181, 185, 182, 241, 164, 180, 193, 246, 189, 26, 47, 16,
			98, 19, 100, 82, 156, 53, 198, 225, 151, 77, 177, 109, 108, 146,
			110, 98, 218, 61, 242, 231, 79, 76, 98, 90, 6, 181, 114, 198,
			8, 202, 255, 190, 201, 90, 2, 2, 164, 74, 34, 46, 8, 31,
			58, 161, 23, 36, 17, 19, 106, 97, 197, 157, 181, 136, 197, 1,
			176, 136, 149, 172, 165, 35, 39, 172, 123, 126, 133, 237, 5, 113,
			36, 132, 35, 98, 188, 184, 179, 198, 156, 70, 163, 230, 113, 192,
			40, 16, 194, 30, 4, 33, 227, 175, 157, 122, 163, 198, 103, 152,
			23, 167, 92, 226, 128, 69, 130, 179, 192, 147, 100, 2, 22, 37,
			245, 186, 19, 122, 95, 113, 182, 119, 44, 38, 118, 67, 167, 124,
			192, 67, 33, 195, 12, 97, 142, 127, 42, 166, 83, 113, 60, 63,
			138, 217, 81, 213, 43, 87, 153, 176, 47, 139, 157, 232, 32, 98,
			78, 200, 89, 221, 241, 157, 10, 119, 83, 146, 78, 116, 192, 67,
			34, 72, 22, 8, 33, 4, 91, 6, 162, 56, 103, 247, 147, 143,
			136, 101, 25, 166, 65, 49, 53, 7, 242, 115, 76, 230, 7, 197,
			45, 98, 14, 48, 211, 84, 212, 92, 26, 243, 5, 66, 206, 145,
			12, 96, 103, 0, 221, 214, 16, 162, 152, 118, 247, 106, 8, 83,
			76, 105, 63, 249, 49, 18, 140, 16, 197, 195, 230, 68, 254, 77,
			147, 74, 164, 183, 38, 182, 181, 119, 220, 84, 116, 115, 9, 104,
			21, 252, 109, 223, 227, 53, 55, 210, 126, 144, 206, 179, 58, 143,
			34, 167, 194, 89, 61, 137, 98, 86, 119, 226, 114, 149, 237, 7,
			161, 146, 62, 14, 192, 131, 36, 19, 240, 81, 37, 24, 202, 130,
			40, 90, 76, 4, 130, 229, 242, 26, 194, 20, 15, 95, 26, 39,
			107, 196, 180, 16, 181, 70, 141, 9, 148, 95, 96, 109, 233, 13,
			172, 234, 176, 40, 217, 139, 120, 12, 30, 223, 234, 30, 237, 146,
			131, 182, 129, 252, 168, 61, 72, 122, 136, 101, 33, 208, 246, 152,
			57, 38, 120, 33, 211, 176, 0, 34, 26, 202, 80, 60, 214, 211,
			167, 33, 68, 241, 24, 29, 214, 16, 166, 120, 44, 63, 170, 136,
			32, 138, 47, 153, 253, 106, 10, 101, 0, 178, 53, 4, 115, 221,
			23, 52, 132, 41, 190, 212, 71, 21, 154, 73, 241, 184, 57, 164,
			166, 204, 12, 64, 26, 13, 72, 142, 119, 107, 222, 38, 166, 120,
			124, 96, 144, 252, 212, 36, 102, 198, 160, 214, 183, 140, 69, 148,
			255, 137, 201, 154, 25, 81, 168, 192, 117, 26, 49, 119, 217, 126,
			24, 212, 89, 185, 26, 6, 117, 94, 144, 255, 130, 232, 165, 200,
			71, 5, 153, 129, 228, 135, 170, 224, 249, 135, 220, 143, 131, 240,
			152, 176, 105, 73, 169, 32, 72, 113, 63, 169, 95, 43, 176, 77,
			17, 45, 50, 234, 184, 203, 220, 224, 200, 7, 251, 5, 126, 237,
			152, 137, 28, 19, 50, 145, 182, 101, 114, 112, 66, 78, 100, 236,
			149, 157, 189, 26, 135, 149, 37, 201, 140, 16, 182, 22, 212, 235,
			220, 143, 185, 59, 27, 36, 177, 246, 28, 136, 141, 184, 26, 68,
			92, 18, 224, 175, 189, 40, 214, 238, 148, 138, 198, 132, 168, 44,
			42, 87, 121, 221, 153, 97, 123, 73, 76, 152, 27, 48, 63, 16,
			188, 142, 129, 141, 200, 111, 43, 59, 155, 210, 188, 25, 176, 211,
			183, 50, 189, 160, 226, 140, 8, 166, 105, 105, 153, 140, 140, 134,
			105, 243, 130, 134, 76, 138, 167, 251, 40, 25, 19, 11, 17, 197,
			215, 205, 145, 124, 47, 19, 201, 124, 229, 193, 6, 187, 203, 62,
			188, 67, 212, 90, 36, 166, 53, 29, 200, 113, 215, 135, 134, 201,
			29, 129, 105, 82, 60, 107, 14, 230, 11, 18, 115, 125, 101, 119,
			101, 117, 165, 4, 232, 55, 238, 144, 246, 177, 151, 165, 173, 149,
			231, 48, 115, 51, 37, 12, 124, 103, 205, 156, 134, 128, 86, 255,
			0, 249, 3, 36, 40, 99, 138, 111, 154, 99, 249, 175, 21, 21,
			248, 26, 177, 187, 236, 182, 38, 251, 201, 246, 214, 182, 252, 12,
			177, 187, 108, 65, 143, 62, 220, 46, 237, 54, 63, 106, 236, 46,
			91, 212, 51, 173, 131, 75, 233, 224, 198, 147, 221, 205, 39, 27,
			91, 176, 221, 15, 210, 193, 135, 43, 197, 117, 24, 105, 106, 0,
			35, 16, 102, 72, 67, 38, 197, 55, 47, 142, 42, 37, 91, 20,
			223, 50, 47, 170, 41, 11, 1, 164, 85, 101, 153, 20, 223, 26,
			26, 81, 11, 51, 20, 223, 54, 39, 212, 84, 6, 1, 164, 209,
			50, 38, 197, 183, 199, 198, 213, 194, 44, 197, 11, 41, 179, 44,
			2, 72, 235, 40, 107, 82, 188, 208, 63, 72, 254, 21, 17, 160,
			110, 221, 53, 214, 80, 254, 31, 145, 138, 133, 223, 76, 28, 164,
			73, 44, 62, 110, 240, 183, 132, 130, 19, 53, 100, 122, 78, 125,
			159, 52, 157, 127, 70, 124, 50, 130, 125, 230, 249, 128, 26, 137,
			12, 88, 174, 121, 220, 143, 35, 249, 109, 230, 236, 196, 201, 0,
			92, 89, 167, 42, 240, 142, 187, 246, 5, 145, 46, 76, 240, 229,
			123, 42, 93, 152, 34, 57, 221, 83, 233, 194, 20, 158, 125, 79,
			165, 11, 83, 36, 167, 123, 3, 131, 10, 13, 81, 124, 223, 100,
			106, 10, 89, 0, 17, 13, 101, 41, 190, 223, 51, 168, 33, 88,
			57, 52, 170, 33, 76, 241, 253, 241, 9, 69, 196, 164, 120, 85,
			90, 14, 0, 11, 32, 77, 4, 172, 186, 218, 163, 121, 3, 187,
			85, 154, 215, 16, 166, 120, 245, 210, 248, 94, 86, 104, 250, 6,
			249, 217, 18, 249, 232, 27, 157, 172, 83, 219, 156, 125, 184, 254,
			25, 34, 195, 159, 240, 184, 20, 131, 254, 213, 201, 165, 200, 191,
			151, 240, 40, 166, 140, 244, 236, 37, 94, 205, 221, 117, 194, 10,
			143, 213, 105, 181, 117, 168, 195, 57, 184, 245, 136, 139, 79, 28,
			113, 151, 201, 72, 228, 196, 53, 103, 111, 211, 223, 15, 194, 186,
			19, 123, 129, 239, 212, 158, 38, 60, 60, 30, 177, 24, 154, 182,
			139, 29, 231, 39, 255, 30, 145, 145, 95, 149, 53, 106, 4, 126,
			196, 65, 88, 216, 161, 26, 214, 194, 182, 12, 209, 105, 210, 187,
			239, 133, 245, 35, 39, 212, 200, 74, 236, 147, 195, 64, 107, 223,
			217, 143, 245, 42, 185, 135, 214, 33, 122, 157, 228, 32, 147, 7,
			107, 45, 44, 45, 177, 236, 87, 198, 233, 16, 201, 134, 220, 137,
			2, 127, 36, 35, 86, 40, 104, 146, 147, 209, 146, 216, 234, 106,
			224, 132, 238, 138, 239, 126, 10, 218, 43, 197, 161, 19, 243, 202,
			49, 104, 81, 105, 98, 93, 23, 10, 26, 6, 189, 239, 1, 146,
			214, 187, 0, 154, 214, 192, 45, 214, 152, 220, 33, 67, 146, 205,
			67, 101, 131, 119, 226, 208, 106, 67, 179, 221, 134, 147, 255, 99,
			146, 75, 37, 30, 75, 170, 167, 122, 206, 33, 25, 141, 58, 111,
			77, 48, 235, 153, 159, 239, 84, 222, 116, 198, 124, 104, 20, 207,
			34, 76, 43, 100, 40, 58, 117, 175, 170, 32, 156, 61, 147, 229,
			73, 164, 135, 70, 177, 3, 185, 147, 222, 134, 223, 201, 219, 172,
			211, 189, 109, 138, 156, 215, 67, 155, 117, 167, 194, 149, 155, 180,
			15, 174, 18, 98, 71, 138, 251, 36, 35, 227, 157, 244, 47, 163,
			97, 242, 128, 92, 110, 81, 163, 80, 210, 58, 175, 113, 136, 166,
			181, 208, 139, 121, 8, 140, 127, 83, 30, 246, 130, 76, 180, 43,
			240, 155, 177, 58, 203, 213, 254, 214, 36, 76, 80, 227, 103, 120,
			219, 143, 17, 185, 28, 189, 109, 183, 202, 233, 22, 223, 238, 116,
			167, 227, 63, 52, 138, 111, 103, 66, 127, 128, 200, 68, 116, 182,
			46, 148, 43, 222, 126, 39, 87, 60, 77, 140, 183, 49, 104, 115,
			149, 23, 228, 242, 25, 250, 83, 185, 115, 140, 116, 187, 158, 43,
			215, 9, 61, 217, 197, 230, 64, 75, 254, 50, 219, 242, 215, 123,
			228, 202, 122, 82, 111, 180, 145, 220, 13, 214, 157, 216, 137, 226,
			32, 228, 202, 58, 147, 87, 201, 212, 217, 203, 164, 16, 243, 127,
			105, 145, 238, 77, 253, 209, 162, 17, 201, 157, 76, 245, 180, 208,
			65, 101, 29, 190, 95, 249, 185, 119, 94, 175, 244, 240, 35, 68,
			134, 78, 15, 44, 122, 179, 99, 47, 230, 140, 60, 152, 191, 245,
			13, 177, 148, 28, 127, 136, 200, 197, 142, 86, 163, 11, 29, 136,
			190, 45, 78, 242, 157, 124, 255, 237, 14, 242, 167, 136, 140, 157,
			101, 68, 186, 220, 137, 244, 219, 29, 36, 127, 231, 215, 194, 149,
			146, 253, 154, 13, 163, 127, 186, 74, 186, 104, 198, 50, 126, 132,
			230, 126, 91, 59, 70, 119, 136, 153, 53, 168, 117, 206, 184, 132,
			242, 115, 44, 13, 44, 81, 128, 122, 101, 206, 170, 188, 214, 136,
			84, 139, 133, 173, 137, 67, 254, 118, 137, 173, 63, 219, 85, 71,
			232, 44, 156, 140, 207, 217, 125, 228, 175, 17, 177, 178, 162, 30,
			204, 153, 165, 252, 207, 17, 59, 25, 62, 44, 228, 142, 27, 9,
			17, 35, 49, 206, 14, 213, 132, 108, 95, 52, 156, 48, 246, 202,
			73, 205, 9, 129, 188, 44, 46, 92, 109, 72, 66, 64, 199, 161,
			208, 20, 28, 245, 125, 121, 224, 171, 29, 67, 165, 90, 220, 89,
			19, 52, 64, 131, 71, 161, 23, 123, 126, 229, 52, 54, 71, 94,
			92, 101, 78, 184, 231, 197, 161, 19, 30, 179, 114, 32, 232, 200,
			30, 73, 86, 22, 175, 185, 236, 144, 134, 76, 138, 115, 195, 5,
			13, 97, 138, 115, 75, 79, 201, 159, 200, 77, 34, 138, 7, 76,
			158, 255, 33, 20, 69, 167, 69, 40, 139, 120, 124, 234, 78, 15,
			157, 90, 194, 217, 52, 8, 43, 123, 89, 107, 225, 118, 73, 86,
			50, 43, 162, 54, 241, 20, 154, 244, 13, 112, 149, 48, 137, 171,
			215, 102, 136, 216, 160, 204, 232, 204, 229, 96, 154, 72, 148, 71,
			169, 248, 80, 92, 12, 100, 199, 52, 100, 82, 60, 112, 105, 89,
			67, 152, 226, 129, 141, 50, 249, 169, 20, 223, 164, 120, 196, 244,
			243, 127, 132, 58, 127, 51, 153, 43, 102, 148, 52, 146, 237, 169,
			123, 17, 102, 58, 185, 11, 185, 183, 211, 247, 242, 246, 157, 128,
			122, 71, 178, 19, 26, 2, 105, 217, 125, 13, 97, 138, 71, 30,
			215, 200, 27, 177, 17, 76, 241, 152, 121, 148, 247, 217, 89, 57,
			224, 12, 191, 3, 217, 99, 29, 186, 229, 192, 223, 247, 42, 34,
			108, 100, 35, 18, 92, 137, 71, 204, 19, 53, 101, 234, 137, 169,
			152, 80, 176, 143, 101, 39, 53, 100, 82, 60, 118, 229, 129, 134,
			64, 178, 167, 9, 185, 40, 219, 177, 204, 184, 129, 242, 231, 153,
			207, 95, 199, 44, 118, 42, 203, 236, 86, 179, 47, 201, 236, 9,
			18, 235, 190, 228, 21, 107, 52, 95, 97, 187, 219, 235, 219, 211,
			149, 144, 87, 130, 240, 216, 247, 162, 61, 30, 95, 91, 102, 33,
			175, 7, 135, 156, 69, 73, 163, 17, 132, 177, 80, 98, 45, 8,
			14, 192, 215, 147, 6, 57, 185, 53, 175, 89, 22, 177, 189, 99,
			214, 82, 140, 177, 247, 153, 56, 142, 181, 246, 51, 175, 88, 164,
			165, 159, 121, 165, 103, 168, 165, 159, 121, 229, 98, 94, 148, 168,
			194, 235, 167, 172, 65, 221, 53, 204, 0, 68, 90, 58, 138, 83,
			61, 185, 150, 142, 226, 84, 255, 0, 217, 20, 104, 38, 197, 87,
			173, 145, 252, 71, 236, 201, 246, 238, 198, 50, 123, 163, 15, 109,
			95, 11, 37, 191, 105, 17, 109, 70, 74, 246, 181, 232, 92, 57,
			181, 152, 135, 190, 19, 123, 135, 60, 210, 108, 160, 18, 190, 154,
			50, 5, 129, 174, 246, 244, 107, 8, 83, 124, 117, 104, 152, 252,
			141, 37, 184, 98, 138, 231, 173, 171, 249, 95, 88, 132, 149, 58,
			20, 138, 204, 115, 33, 143, 236, 123, 60, 98, 71, 85, 30, 87,
			69, 218, 245, 32, 97, 139, 47, 142, 104, 117, 180, 169, 114, 86,
			52, 36, 244, 116, 57, 168, 123, 126, 133, 52, 221, 72, 57, 245,
			218, 214, 102, 129, 16, 86, 146, 150, 2, 3, 189, 42, 204, 169,
			185, 10, 143, 103, 165, 169, 102, 181, 169, 102, 197, 166, 89, 161,
			80, 96, 179, 226, 248, 12, 63, 95, 205, 52, 131, 200, 231, 101,
			30, 69, 144, 178, 192, 232, 73, 196, 195, 72, 246, 121, 9, 211,
			157, 192, 106, 112, 116, 138, 131, 203, 22, 160, 244, 235, 36, 228,
			174, 192, 143, 171, 220, 11, 117, 224, 205, 136, 205, 120, 33, 143,
			88, 162, 136, 166, 52, 149, 118, 142, 137, 222, 112, 36, 240, 133,
			136, 239, 75, 153, 117, 99, 82, 166, 100, 87, 182, 253, 129, 68,
			20, 5, 101, 207, 137, 185, 43, 147, 110, 123, 180, 131, 114, 54,
			229, 183, 202, 7, 205, 214, 216, 126, 45, 56, 210, 27, 6, 167,
			142, 152, 19, 51, 237, 39, 50, 59, 204, 176, 35, 206, 92, 47,
			2, 125, 38, 94, 84, 101, 142, 166, 26, 132, 204, 15, 252, 89,
			5, 105, 227, 36, 145, 176, 205, 171, 42, 79, 66, 192, 42, 71,
			133, 45, 160, 188, 229, 29, 168, 124, 183, 46, 164, 153, 214, 108,
			174, 189, 130, 15, 154, 244, 37, 156, 1, 239, 209, 61, 126, 8,
			244, 249, 238, 203, 26, 2, 207, 154, 122, 79, 132, 54, 162, 214,
			109, 99, 173, 61, 180, 111, 55, 155, 224, 183, 109, 214, 108, 130,
			47, 88, 163, 45, 109, 239, 5, 139, 180, 180, 189, 23, 122, 134,
			90, 218, 222, 11, 42, 226, 68, 219, 123, 209, 154, 104, 105, 123,
			47, 166, 104, 64, 127, 177, 39, 223, 210, 246, 94, 188, 52, 222,
			108, 123, 47, 165, 220, 32, 102, 150, 82, 52, 32, 185, 148, 114,
			131, 40, 89, 74, 185, 97, 138, 151, 45, 166, 166, 64, 5, 203,
			41, 26, 168, 96, 185, 71, 147, 4, 21, 44, 143, 79, 144, 255,
			70, 2, 207, 162, 120, 213, 26, 202, 255, 7, 98, 69, 113, 226,
			87, 113, 195, 95, 55, 106, 142, 47, 83, 80, 176, 207, 174, 31,
			85, 143, 175, 203, 248, 106, 249, 210, 183, 59, 44, 97, 71, 78,
			196, 26, 94, 249, 128, 187, 5, 182, 19, 68, 145, 39, 102, 225,
			75, 35, 156, 121, 153, 16, 246, 225, 53, 54, 9, 70, 155, 141,
			26, 188, 236, 237, 123, 229, 73, 194, 230, 175, 177, 73, 25, 59,
			144, 83, 100, 185, 41, 142, 12, 94, 164, 190, 83, 250, 220, 208,
			116, 82, 175, 53, 33, 200, 62, 34, 247, 68, 18, 168, 5, 149,
			10, 119, 193, 177, 26, 112, 118, 243, 99, 137, 3, 123, 242, 93,
			17, 128, 5, 173, 24, 43, 3, 155, 79, 33, 212, 236, 212, 33,
			211, 194, 20, 175, 14, 12, 146, 127, 83, 141, 212, 135, 198, 22,
			202, 255, 18, 177, 51, 26, 23, 242, 162, 165, 28, 212, 247, 188,
			166, 230, 14, 248, 177, 186, 3, 112, 249, 190, 231, 115, 112, 253,
			246, 108, 207, 253, 88, 101, 7, 73, 155, 164, 161, 38, 85, 112,
			224, 201, 38, 169, 92, 7, 103, 207, 114, 208, 80, 155, 98, 39,
			244, 86, 16, 39, 45, 29, 92, 155, 235, 172, 153, 16, 185, 203,
			246, 146, 152, 149, 147, 48, 228, 126, 92, 59, 102, 94, 197, 15,
			66, 113, 205, 164, 155, 169, 15, 237, 43, 205, 102, 234, 102, 91,
			51, 117, 179, 173, 153, 186, 217, 214, 76, 221, 108, 109, 166, 62,
			82, 55, 61, 166, 112, 249, 71, 41, 26, 184, 252, 35, 117, 211,
			35, 219, 167, 143, 212, 77, 143, 56, 36, 60, 78, 209, 192, 229,
			31, 167, 104, 64, 242, 113, 138, 6, 94, 254, 184, 143, 146, 127,
			0, 163, 96, 106, 237, 24, 37, 148, 255, 59, 109, 148, 147, 13,
			28, 125, 134, 57, 224, 199, 242, 90, 83, 30, 85, 75, 167, 30,
			136, 164, 118, 133, 161, 78, 168, 88, 123, 170, 200, 105, 5, 194,
			62, 227, 105, 158, 138, 219, 117, 237, 51, 199, 117, 61, 97, 121,
			113, 243, 194, 211, 52, 40, 88, 239, 39, 113, 18, 242, 217, 70,
			24, 4, 251, 240, 97, 145, 37, 182, 238, 102, 67, 148, 238, 216,
			50, 11, 96, 48, 192, 83, 101, 0, 44, 12, 240, 84, 169, 4,
			11, 3, 60, 85, 6, 192, 194, 0, 79, 149, 1, 48, 104, 171,
			152, 162, 129, 1, 138, 41, 26, 24, 160, 152, 162, 129, 1, 138,
			3, 131, 132, 16, 211, 178, 168, 245, 204, 248, 46, 18, 82, 64,
			16, 60, 179, 223, 35, 231, 136, 101, 89, 182, 65, 173, 231, 230,
			11, 44, 112, 44, 27, 24, 63, 183, 115, 130, 149, 5, 18, 126,
			102, 221, 151, 83, 166, 145, 5, 104, 66, 67, 136, 226, 207, 216,
			29, 13, 97, 138, 63, 187, 247, 177, 66, 67, 20, 127, 219, 186,
			173, 166, 80, 22, 160, 188, 134, 96, 110, 244, 67, 13, 97, 138,
			191, 125, 243, 150, 66, 51, 41, 254, 92, 222, 134, 0, 144, 1,
			200, 214, 16, 162, 248, 243, 238, 1, 13, 97, 138, 63, 31, 30,
			81, 104, 152, 226, 239, 152, 151, 212, 20, 100, 197, 239, 164, 104,
			160, 239, 239, 116, 143, 104, 8, 86, 142, 142, 145, 239, 10, 52,
			139, 226, 47, 204, 209, 252, 83, 25, 128, 250, 40, 12, 118, 140,
			171, 96, 57, 71, 142, 76, 182, 116, 139, 39, 197, 245, 27, 156,
			55, 147, 72, 157, 188, 125, 126, 4, 135, 18, 30, 214, 61, 63,
			168, 5, 21, 117, 72, 6, 250, 25, 96, 160, 37, 1, 157, 127,
			209, 61, 164, 33, 76, 241, 23, 23, 243, 194, 50, 25, 106, 189,
			50, 28, 105, 153, 12, 162, 248, 149, 125, 149, 252, 23, 248, 126,
			150, 90, 85, 163, 134, 32, 109, 191, 181, 169, 37, 252, 217, 171,
			123, 144, 177, 227, 224, 172, 4, 54, 67, 228, 21, 34, 252, 57,
			215, 68, 21, 12, 251, 80, 55, 146, 226, 202, 122, 143, 179, 70,
			24, 28, 122, 110, 243, 64, 210, 140, 6, 145, 153, 143, 120, 173,
			54, 11, 201, 89, 212, 204, 240, 183, 119, 13, 98, 166, 146, 56,
			161, 227, 199, 92, 93, 104, 198, 162, 32, 60, 242, 106, 53, 64,
			226, 175, 157, 50, 228, 164, 192, 231, 108, 143, 167, 87, 227, 112,
			64, 1, 33, 188, 58, 151, 97, 66, 176, 149, 69, 20, 87, 237,
			107, 194, 192, 89, 240, 66, 79, 57, 124, 86, 196, 137, 167, 212,
			154, 21, 94, 232, 41, 135, 207, 10, 47, 244, 84, 156, 100, 193,
			101, 190, 84, 25, 39, 43, 226, 228, 203, 20, 13, 188, 240, 75,
			149, 113, 178, 194, 11, 191, 84, 137, 42, 11, 94, 120, 144, 162,
			129, 23, 30, 164, 104, 64, 242, 32, 69, 3, 199, 59, 232, 163,
			228, 63, 193, 88, 93, 212, 138, 140, 67, 148, 255, 247, 147, 137,
			234, 221, 76, 117, 50, 173, 253, 191, 176, 82, 23, 162, 56, 178,
			191, 37, 244, 214, 5, 86, 138, 149, 149, 186, 132, 149, 98, 165,
			183, 46, 97, 165, 88, 89, 169, 75, 88, 41, 86, 86, 234, 2,
			149, 38, 41, 26, 88, 41, 73, 209, 192, 74, 73, 138, 6, 86,
			74, 84, 54, 179, 169, 245, 218, 248, 90, 198, 140, 141, 40, 126,
			109, 79, 139, 108, 102, 67, 54, 59, 54, 127, 71, 102, 51, 91,
			100, 179, 99, 149, 205, 108, 144, 240, 43, 107, 75, 78, 137, 108,
			246, 149, 245, 158, 134, 16, 197, 95, 93, 253, 68, 67, 152, 226,
			175, 30, 61, 86, 104, 136, 226, 55, 214, 67, 53, 5, 217, 236,
			141, 117, 69, 67, 48, 55, 181, 166, 33, 76, 241, 155, 7, 159,
			144, 30, 98, 90, 221, 52, 243, 125, 227, 7, 72, 138, 216, 141,
			40, 254, 190, 125, 77, 136, 216, 109, 26, 212, 250, 93, 100, 14,
			146, 243, 36, 3, 80, 70, 128, 89, 13, 34, 0, 187, 114, 26,
			196, 0, 246, 15, 40, 84, 68, 173, 223, 67, 230, 128, 154, 68,
			25, 1, 218, 26, 20, 179, 221, 189, 26, 196, 0, 210, 126, 161,
			48, 66, 173, 31, 34, 163, 64, 122, 8, 182, 8, 2, 192, 126,
			95, 76, 244, 80, 235, 71, 200, 152, 19, 19, 61, 8, 0, 123,
			38, 189, 171, 252, 229, 48, 89, 254, 70, 119, 149, 177, 124, 44,
			116, 246, 77, 165, 71, 198, 118, 146, 168, 186, 26, 196, 209, 131,
			32, 20, 253, 136, 93, 39, 58, 136, 244, 45, 192, 38, 201, 197,
			162, 218, 124, 233, 38, 241, 203, 40, 118, 84, 47, 251, 194, 252,
			68, 199, 6, 99, 92, 130, 101, 197, 11, 18, 81, 195, 147, 19,
			228, 82, 7, 86, 234, 122, 229, 61, 114, 5, 22, 20, 121, 195,
			241, 194, 71, 193, 30, 44, 219, 114, 246, 128, 169, 120, 82, 216,
			108, 125, 159, 189, 76, 145, 235, 39, 125, 69, 14, 53, 37, 112,
			212, 200, 3, 132, 182, 14, 170, 165, 47, 200, 229, 147, 162, 173,
			36, 174, 23, 183, 169, 226, 38, 177, 98, 39, 58, 80, 219, 103,
			29, 182, 159, 226, 21, 197, 234, 201, 41, 50, 121, 22, 105, 41,
			192, 245, 95, 32, 98, 107, 69, 209, 126, 210, 171, 127, 111, 250,
			135, 78, 205, 115, 115, 6, 237, 38, 153, 34, 119, 220, 227, 28,
			162, 57, 114, 238, 9, 231, 110, 180, 86, 227, 142, 159, 52, 114,
			38, 237, 37, 61, 98, 68, 234, 36, 135, 233, 5, 66, 212, 64,
			196, 227, 156, 37, 30, 40, 138, 185, 7, 142, 87, 227, 110, 46,
			67, 7, 72, 78, 163, 212, 156, 50, 175, 115, 63, 206, 101, 233,
			32, 233, 19, 163, 159, 58, 126, 226, 212, 20, 185, 174, 148, 254,
			58, 111, 212, 130, 227, 156, 125, 125, 139, 116, 167, 187, 128, 89,
			248, 223, 148, 181, 151, 244, 148, 120, 120, 24, 60, 43, 173, 62,
			230, 32, 241, 5, 66, 214, 159, 237, 150, 226, 32, 116, 42, 60,
			103, 210, 243, 164, 187, 184, 243, 233, 154, 40, 181, 115, 120, 254,
			207, 45, 210, 165, 158, 182, 209, 31, 32, 50, 120, 170, 155, 208,
			27, 29, 52, 126, 150, 255, 230, 59, 221, 48, 156, 233, 137, 162,
			51, 127, 150, 143, 117, 236, 204, 191, 131, 255, 118, 236, 204, 191,
			139, 83, 83, 135, 144, 166, 255, 210, 233, 14, 164, 126, 197, 239,
			243, 215, 222, 97, 165, 98, 241, 199, 136, 228, 59, 187, 44, 237,
			116, 223, 241, 214, 0, 202, 47, 253, 26, 152, 255, 167, 11, 137,
			63, 59, 255, 219, 254, 130, 213, 211, 247, 17, 99, 40, 255, 69,
			250, 60, 84, 223, 70, 240, 215, 141, 0, 206, 198, 97, 163, 124,
			226, 61, 169, 234, 126, 197, 98, 63, 233, 61, 133, 108, 74, 147,
			246, 7, 139, 210, 70, 173, 183, 23, 189, 112, 104, 23, 151, 23,
			86, 175, 153, 43, 228, 119, 216, 3, 249, 22, 54, 72, 98, 81,
			25, 127, 47, 225, 137, 228, 177, 22, 110, 151, 154, 175, 98, 117,
			251, 172, 237, 77, 170, 116, 121, 192, 35, 76, 164, 171, 182, 43,
			134, 222, 236, 104, 203, 21, 67, 239, 216, 98, 235, 21, 67, 239,
			12, 121, 161, 110, 24, 172, 126, 115, 224, 86, 254, 113, 103, 73,
			106, 105, 48, 189, 147, 60, 109, 23, 5, 253, 105, 223, 26, 140,
			209, 159, 246, 173, 197, 69, 65, 239, 13, 114, 69, 223, 19, 12,
			155, 43, 249, 33, 38, 99, 74, 178, 169, 243, 56, 244, 202, 81,
			91, 183, 126, 56, 155, 107, 233, 214, 15, 247, 77, 182, 116, 235,
			135, 103, 63, 38, 91, 170, 91, 111, 229, 205, 209, 27, 249, 123,
			239, 164, 221, 128, 133, 137, 223, 182, 15, 17, 69, 109, 237, 247,
			124, 150, 181, 180, 223, 243, 151, 87, 91, 218, 239, 163, 189, 31,
			138, 35, 136, 65, 173, 113, 131, 161, 180, 223, 62, 110, 79, 169,
			126, 182, 65, 241, 132, 122, 183, 101, 136, 3, 219, 132, 217, 218,
			6, 159, 80, 253, 46, 105, 154, 137, 241, 9, 112, 83, 40, 180,
			38, 141, 233, 180, 197, 55, 105, 191, 39, 134, 77, 138, 167, 140,
			66, 218, 6, 153, 178, 223, 23, 195, 152, 226, 171, 198, 92, 90,
			156, 95, 181, 103, 196, 176, 69, 241, 180, 49, 150, 86, 203, 211,
			246, 69, 49, 156, 161, 248, 186, 113, 41, 45, 213, 174, 219, 178,
			132, 203, 82, 107, 198, 40, 160, 180, 116, 153, 105, 45, 93, 102,
			205, 65, 93, 158, 100, 1, 234, 105, 41, 93, 102, 207, 229, 90,
			74, 151, 217, 254, 1, 193, 165, 139, 226, 15, 140, 217, 244, 136,
			253, 129, 125, 29, 184, 100, 12, 106, 205, 27, 31, 161, 244, 137,
			231, 124, 230, 124, 243, 137, 231, 141, 244, 173, 32, 76, 221, 48,
			251, 72, 243, 137, 231, 13, 121, 216, 150, 79, 60, 111, 154, 231,
			72, 243, 69, 231, 77, 179, 139, 52, 95, 116, 222, 36, 61, 106,
			161, 217, 250, 130, 209, 20, 239, 25, 245, 163, 81, 49, 39, 171,
			37, 249, 64, 243, 182, 73, 73, 243, 133, 228, 109, 243, 60, 105,
			190, 144, 188, 157, 235, 107, 190, 144, 92, 72, 165, 178, 196, 123,
			70, 45, 8, 216, 102, 161, 55, 215, 124, 33, 185, 152, 178, 6,
			45, 47, 166, 172, 51, 38, 197, 139, 41, 235, 44, 197, 75, 230,
			48, 105, 190, 144, 92, 74, 5, 201, 154, 20, 47, 13, 14, 169,
			133, 93, 20, 47, 155, 35, 106, 10, 52, 186, 156, 210, 239, 50,
			41, 94, 30, 26, 86, 11, 109, 138, 239, 164, 52, 160, 174, 184,
			147, 110, 198, 54, 41, 190, 147, 235, 19, 134, 64, 212, 186, 7,
			9, 30, 12, 1, 74, 188, 151, 17, 239, 19, 51, 162, 139, 252,
			177, 34, 32, 251, 198, 31, 43, 2, 72, 24, 226, 99, 165, 13,
			36, 95, 36, 234, 133, 226, 213, 97, 186, 16, 12, 113, 63, 93,
			104, 82, 188, 162, 212, 38, 123, 195, 43, 74, 109, 72, 24, 98,
			69, 169, 77, 244, 134, 87, 213, 11, 81, 217, 13, 94, 149, 158,
			6, 144, 73, 241, 234, 133, 94, 125, 172, 255, 223, 0, 0, 0,
			255, 255, 81, 95, 63, 172, 221, 51, 0, 0},
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
	ret, err := discovery.GetDescriptorSet("crosskylabadmin.fleet.Inventory")
	if err != nil {
		panic(err)
	}
	return ret
}
