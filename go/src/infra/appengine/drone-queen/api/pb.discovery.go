// Code generated by cproto. DO NOT EDIT.

package api

import "go.chromium.org/luci/grpc/discovery"

import "google.golang.org/protobuf/types/descriptorpb"

func init() {
	discovery.RegisterDescriptorSetCompressed(
		[]string{
			"drone_queen.Drone", "drone_queen.InventoryProvider", "drone_queen.Inspect",
		},
		[]byte{31, 139,
			8, 0, 0, 0, 0, 0, 0, 255, 164, 122, 75, 108, 27, 73,
			122, 48, 187, 171, 37, 145, 37, 191, 84, 246, 120, 52, 180, 199,
			254, 76, 143, 118, 164, 89, 170, 41, 201, 143, 249, 101, 239, 252,
			179, 148, 72, 89, 244, 74, 164, 126, 62, 172, 181, 189, 187, 118,
			139, 93, 36, 123, 166, 217, 77, 119, 23, 37, 203, 30, 207, 236,
			191, 185, 100, 129, 32, 1, 114, 8, 2, 4, 11, 36, 57, 36,
			192, 230, 28, 228, 144, 83, 238, 185, 5, 8, 144, 123, 246, 190,
			192, 94, 114, 12, 190, 170, 238, 38, 245, 240, 120, 103, 87, 24,
			99, 250, 171, 199, 247, 174, 239, 81, 69, 250, 31, 5, 122, 189,
			235, 251, 93, 151, 23, 6, 129, 47, 252, 189, 97, 167, 32, 156,
			62, 15, 133, 213, 31, 152, 114, 136, 157, 87, 11, 204, 120, 65,
			238, 62, 205, 52, 227, 53, 108, 150, 78, 133, 188, 237, 123, 118,
			56, 171, 129, 54, 79, 234, 49, 200, 46, 209, 9, 207, 242, 252,
			112, 86, 7, 109, 126, 162, 174, 128, 181, 175, 233, 197, 182, 223,
			55, 143, 225, 92, 59, 151, 96, 220, 193, 161, 29, 237, 201, 247,
			187, 142, 232, 13, 247, 204, 182, 223, 47, 116, 125, 215, 242, 186,
			35, 22, 7, 226, 112, 192, 195, 17, 167, 255, 163, 105, 127, 175,
			147, 7, 59, 107, 191, 214, 175, 61, 80, 152, 119, 162, 181, 230,
			46, 119, 221, 31, 121, 254, 129, 215, 196, 61, 15, 255, 107, 145,
			78, 177, 137, 107, 169, 95, 106, 26, 253, 247, 51, 84, 59, 195,
			200, 181, 20, 91, 249, 183, 51, 32, 119, 180, 125, 23, 214, 134,
			157, 14, 15, 66, 88, 4, 133, 235, 227, 16, 108, 75, 88, 224,
			120, 130, 7, 237, 158, 229, 117, 57, 116, 252, 160, 111, 9, 10,
			235, 254, 224, 48, 112, 186, 61, 1, 43, 75, 75, 255, 39, 218,
			0, 21, 175, 109, 2, 20, 93, 23, 228, 92, 8, 1, 15, 121,
			176, 207, 109, 147, 66, 79, 136, 65, 120, 175, 80, 176, 249, 62,
			119, 253, 1, 15, 194, 88, 25, 40, 233, 32, 98, 98, 113, 79,
			49, 81, 160, 20, 234, 220, 118, 66, 17, 56, 123, 67, 225, 248,
			30, 88, 158, 13, 195, 144, 131, 227, 65, 232, 15, 131, 54, 151,
			35, 123, 142, 103, 5, 135, 146, 175, 48, 15, 7, 142, 232, 129,
			31, 200, 255, 251, 67, 65, 161, 239, 219, 78, 199, 105, 91, 136,
			33, 15, 86, 192, 97, 192, 131, 190, 35, 4, 183, 97, 16, 248,
			251, 142, 205, 109, 16, 61, 75, 128, 232, 161, 116, 174, 235, 31,
			56, 94, 23, 208, 148, 14, 110, 10, 113, 19, 133, 62, 23, 247,
			40, 5, 252, 251, 228, 24, 99, 33, 248, 157, 152, 163, 182, 111,
			115, 232, 15, 67, 1, 1, 23, 150, 227, 73, 172, 214, 158, 191,
			143, 83, 145, 198, 40, 120, 190, 112, 218, 60, 15, 162, 231, 132,
			224, 58, 161, 64, 12, 227, 20, 61, 251, 24, 59, 182, 19, 182,
			93, 203, 233, 243, 192, 124, 27, 19, 142, 55, 174, 139, 152, 137,
			65, 224, 219, 195, 54, 31, 241, 65, 71, 140, 252, 81, 124, 80,
			136, 164, 179, 253, 246, 176, 207, 61, 97, 197, 70, 42, 248, 1,
			248, 162, 199, 3, 232, 91, 130, 7, 142, 229, 134, 35, 85, 75,
			3, 137, 30, 167, 48, 206, 125, 34, 84, 149, 59, 114, 39, 34,
			246, 172, 62, 71, 134, 198, 125, 203, 243, 71, 115, 82, 239, 142,
			8, 81, 34, 79, 161, 242, 131, 16, 250, 214, 33, 236, 113, 244,
			20, 27, 132, 15, 220, 179, 253, 32, 228, 232, 20, 131, 192, 239,
			251, 130, 131, 210, 137, 8, 193, 230, 129, 179, 207, 109, 232, 4,
			126, 159, 42, 45, 132, 126, 71, 28, 160, 155, 68, 30, 4, 225,
			128, 183, 209, 131, 96, 16, 56, 232, 88, 1, 250, 142, 167, 188,
			40, 12, 37, 239, 20, 154, 155, 149, 6, 52, 106, 27, 205, 221,
			98, 189, 12, 149, 6, 236, 212, 107, 143, 42, 165, 114, 9, 214,
			30, 67, 115, 179, 12, 235, 181, 157, 199, 245, 202, 131, 205, 38,
			108, 214, 182, 74, 229, 122, 3, 138, 213, 18, 172, 215, 170, 205,
			122, 101, 173, 213, 172, 213, 27, 20, 114, 197, 6, 84, 26, 57,
			57, 83, 172, 62, 134, 242, 143, 119, 234, 229, 70, 3, 106, 117,
			168, 108, 239, 108, 85, 202, 37, 216, 45, 214, 235, 197, 106, 179,
			82, 110, 228, 161, 82, 93, 223, 106, 149, 42, 213, 7, 121, 88,
			107, 53, 161, 90, 107, 82, 216, 170, 108, 87, 154, 229, 18, 52,
			107, 121, 73, 246, 228, 62, 168, 109, 192, 118, 185, 190, 190, 89,
			172, 54, 139, 107, 149, 173, 74, 243, 177, 36, 184, 81, 105, 86,
			145, 216, 70, 173, 78, 161, 8, 59, 197, 122, 179, 178, 222, 218,
			42, 214, 97, 167, 85, 223, 169, 53, 202, 128, 146, 149, 42, 141,
			245, 173, 98, 101, 187, 92, 50, 161, 82, 133, 106, 13, 202, 143,
			202, 213, 38, 52, 54, 139, 91, 91, 71, 5, 165, 80, 219, 173,
			150, 235, 200, 253, 184, 152, 176, 86, 134, 173, 74, 113, 109, 171,
			140, 164, 164, 156, 165, 74, 189, 188, 222, 68, 129, 70, 95, 235,
			149, 82, 185, 218, 44, 110, 229, 41, 52, 118, 202, 235, 149, 226,
			86, 30, 202, 63, 46, 111, 239, 108, 21, 235, 143, 243, 17, 210,
			70, 249, 255, 181, 202, 213, 102, 165, 184, 5, 165, 226, 118, 241,
			65, 185, 1, 243, 239, 210, 202, 78, 189, 182, 222, 170, 151, 183,
			145, 235, 218, 6, 52, 90, 107, 141, 102, 165, 217, 106, 150, 225,
			65, 173, 86, 146, 202, 110, 148, 235, 143, 42, 235, 229, 198, 125,
			216, 170, 53, 164, 194, 90, 141, 114, 158, 66, 169, 216, 44, 74,
			210, 59, 245, 218, 70, 165, 217, 184, 143, 223, 107, 173, 70, 69,
			42, 174, 82, 109, 150, 235, 245, 214, 78, 179, 82, 171, 46, 192,
			102, 109, 183, 252, 168, 92, 135, 245, 98, 171, 81, 46, 73, 13,
			215, 170, 40, 45, 250, 74, 185, 86, 127, 140, 104, 81, 15, 210,
			2, 121, 216, 221, 44, 55, 55, 203, 117, 84, 170, 212, 86, 17,
			213, 208, 104, 214, 43, 235, 205, 241, 101, 181, 58, 52, 107, 245,
			38, 29, 147, 19, 170, 229, 7, 91, 149, 7, 229, 234, 122, 25,
			167, 107, 136, 102, 183, 210, 40, 47, 64, 177, 94, 105, 224, 130,
			138, 36, 12, 187, 197, 199, 80, 107, 73, 169, 209, 80, 173, 70,
			153, 170, 239, 49, 215, 205, 75, 123, 66, 101, 3, 138, 165, 71,
			21, 228, 60, 90, 189, 83, 107, 52, 42, 145, 187, 72, 181, 173,
			111, 70, 58, 55, 41, 77, 83, 77, 103, 4, 82, 179, 248, 149,
			102, 36, 151, 186, 79, 51, 84, 79, 207, 169, 79, 53, 120, 51,
			117, 93, 14, 94, 87, 159, 106, 240, 163, 212, 154, 28, 156, 86,
			159, 106, 112, 46, 149, 151, 131, 154, 250, 84, 131, 223, 75, 21,
			228, 96, 244, 169, 6, 63, 78, 229, 228, 32, 85, 159, 106, 112,
			62, 117, 67, 14, 126, 164, 62, 127, 119, 133, 234, 70, 138, 77,
			124, 141, 153, 47, 251, 155, 43, 80, 132, 36, 229, 202, 248, 200,
			67, 238, 137, 16, 44, 24, 248, 142, 39, 100, 84, 115, 250, 152,
			101, 108, 62, 224, 158, 205, 61, 25, 21, 45, 239, 80, 141, 191,
			242, 61, 25, 76, 92, 191, 109, 185, 20, 218, 150, 203, 61, 219,
			10, 242, 192, 61, 12, 254, 54, 88, 136, 171, 237, 15, 213, 190,
			168, 40, 144, 161, 180, 19, 88, 237, 81, 194, 136, 39, 48, 31,
			96, 133, 32, 97, 76, 152, 190, 171, 98, 34, 52, 123, 60, 66,
			228, 96, 38, 117, 45, 225, 236, 115, 140, 105, 150, 7, 124, 224,
			183, 123, 96, 9, 104, 53, 215, 161, 239, 216, 158, 12, 232, 190,
			71, 225, 161, 229, 13, 49, 11, 44, 231, 97, 121, 245, 211, 165,
			124, 28, 167, 7, 129, 239, 242, 129, 112, 218, 240, 32, 224, 93,
			63, 112, 44, 47, 225, 30, 14, 122, 78, 187, 7, 252, 165, 224,
			200, 147, 140, 207, 167, 172, 218, 179, 218, 95, 30, 88, 1, 174,
			240, 225, 144, 91, 1, 248, 30, 199, 248, 135, 25, 191, 239, 120,
			67, 193, 101, 186, 132, 187, 75, 137, 124, 174, 239, 117, 77, 216,
			226, 214, 96, 36, 114, 192, 33, 23, 246, 185, 21, 112, 59, 7,
			161, 175, 242, 175, 231, 131, 203, 173, 1, 141, 150, 129, 176, 246,
			92, 142, 146, 123, 156, 163, 94, 59, 126, 160, 42, 145, 1, 166,
			86, 149, 207, 135, 33, 38, 37, 11, 158, 174, 220, 94, 236, 249,
			195, 0, 92, 199, 227, 86, 64, 65, 98, 255, 233, 252, 183, 215,
			28, 104, 207, 130, 92, 185, 32, 131, 120, 143, 67, 32, 139, 28,
			39, 148, 41, 1, 150, 150, 150, 150, 23, 229, 127, 205, 165, 165,
			123, 242, 191, 39, 40, 250, 234, 234, 234, 234, 226, 242, 202, 226,
			173, 229, 230, 202, 173, 123, 119, 86, 239, 221, 89, 53, 87, 227,
			191, 39, 38, 172, 29, 82, 52, 164, 8, 156, 182, 64, 6, 69,
			36, 162, 196, 158, 135, 3, 14, 220, 11, 135, 1, 87, 163, 7,
			28, 218, 168, 101, 223, 219, 231, 129, 80, 246, 85, 57, 9, 158,
			214, 55, 214, 41, 220, 186, 117, 107, 117, 36, 203, 193, 193, 129,
			233, 112, 209, 49, 253, 160, 91, 8, 58, 109, 252, 135, 43, 76,
			241, 82, 44, 96, 193, 198, 1, 41, 123, 221, 16, 133, 186, 9,
			229, 151, 86, 127, 224, 242, 144, 210, 248, 19, 150, 239, 193, 186,
			223, 31, 12, 5, 31, 59, 11, 146, 224, 78, 173, 81, 249, 49,
			60, 71, 205, 204, 47, 60, 55, 163, 138, 103, 180, 40, 169, 60,
			239, 171, 153, 81, 205, 28, 114, 241, 44, 50, 240, 188, 220, 94,
			109, 109, 109, 45, 44, 156, 186, 78, 250, 251, 252, 210, 194, 253,
			49, 158, 86, 222, 197, 83, 151, 11, 196, 226, 119, 108, 235, 112,
			140, 183, 80, 4, 195, 182, 144, 4, 246, 45, 23, 196, 126, 68,
			241, 200, 242, 239, 137, 253, 60, 72, 134, 238, 255, 161, 34, 237,
			155, 98, 31, 161, 111, 147, 72, 45, 26, 134, 188, 13, 159, 192,
			242, 210, 210, 81, 9, 111, 189, 85, 194, 93, 199, 187, 181, 2,
			207, 31, 112, 209, 56, 12, 5, 239, 227, 116, 49, 220, 112, 92,
			222, 60, 106, 136, 141, 202, 86, 185, 89, 217, 46, 67, 71, 68,
			108, 188, 109, 207, 247, 58, 34, 230, 180, 85, 169, 54, 239, 222,
			6, 225, 180, 191, 12, 225, 51, 152, 159, 159, 87, 35, 11, 29,
			97, 218, 7, 155, 78, 183, 87, 178, 132, 220, 181, 0, 63, 248,
			1, 220, 90, 89, 128, 175, 64, 206, 109, 249, 7, 241, 84, 172,
			183, 66, 1, 138, 200, 175, 237, 31, 132, 18, 37, 30, 150, 229,
			165, 165, 177, 24, 22, 154, 201, 2, 21, 165, 150, 239, 158, 60,
			70, 9, 54, 220, 190, 124, 247, 246, 237, 219, 159, 222, 186, 187,
			52, 10, 27, 123, 188, 227, 7, 28, 90, 158, 243, 50, 198, 178,
			250, 233, 210, 113, 44, 230, 31, 102, 204, 121, 37, 63, 204, 207,
			43, 165, 20, 164, 177, 240, 111, 1, 22, 199, 217, 121, 135, 7,
			35, 30, 84, 87, 140, 103, 110, 12, 143, 116, 128, 133, 35, 14,
			112, 251, 173, 14, 240, 208, 218, 183, 224, 185, 50, 164, 217, 30,
			6, 1, 247, 4, 46, 217, 118, 92, 215, 9, 199, 28, 0, 163,
			41, 244, 229, 40, 124, 6, 111, 223, 240, 45, 110, 14, 159, 141,
			70, 77, 143, 31, 172, 13, 29, 215, 230, 193, 252, 2, 10, 214,
			136, 52, 20, 145, 80, 138, 89, 80, 184, 240, 15, 215, 84, 149,
			236, 142, 39, 80, 242, 104, 165, 18, 61, 18, 91, 106, 96, 193,
			220, 67, 204, 146, 151, 145, 14, 238, 188, 85, 7, 145, 20, 113,
			246, 133, 157, 67, 209, 83, 213, 245, 17, 245, 143, 179, 63, 191,
			112, 220, 54, 15, 184, 88, 31, 105, 99, 126, 65, 70, 192, 135,
			141, 90, 21, 182, 173, 193, 192, 241, 186, 148, 66, 197, 83, 35,
			170, 149, 205, 203, 228, 56, 166, 167, 195, 129, 76, 0, 71, 210,
			185, 10, 168, 81, 38, 165, 50, 44, 127, 167, 168, 172, 72, 97,
			70, 183, 48, 153, 231, 21, 26, 53, 138, 196, 114, 175, 49, 155,
			190, 89, 124, 221, 247, 61, 209, 123, 179, 248, 218, 182, 14, 223,
			52, 95, 99, 74, 123, 115, 239, 117, 223, 241, 222, 220, 123, 29,
			242, 246, 155, 167, 230, 107, 44, 34, 208, 145, 223, 252, 244, 73,
			142, 194, 65, 143, 7, 28, 212, 110, 68, 100, 185, 7, 214, 97,
			8, 252, 37, 214, 53, 216, 1, 169, 12, 217, 193, 220, 104, 59,
			93, 71, 132, 152, 234, 93, 14, 17, 165, 60, 72, 82, 121, 10,
			138, 88, 30, 36, 181, 188, 76, 65, 146, 164, 204, 214, 175, 120,
			224, 47, 14, 44, 219, 86, 61, 149, 56, 240, 99, 108, 220, 106,
			247, 84, 165, 18, 87, 55, 88, 21, 69, 7, 45, 31, 213, 21,
			152, 222, 186, 62, 12, 7, 50, 121, 198, 91, 231, 29, 147, 155,
			209, 224, 242, 233, 53, 208, 66, 158, 74, 250, 254, 64, 97, 86,
			148, 114, 79, 114, 16, 14, 59, 29, 231, 37, 86, 105, 216, 220,
			115, 89, 179, 72, 63, 144, 245, 217, 124, 174, 213, 92, 207, 45,
			220, 63, 50, 74, 85, 25, 245, 98, 232, 4, 220, 54, 161, 8,
			242, 206, 225, 150, 114, 134, 80, 54, 170, 206, 43, 30, 64, 216,
			243, 135, 174, 29, 171, 114, 24, 114, 89, 99, 205, 91, 97, 66,
			205, 134, 189, 67, 138, 108, 44, 160, 1, 60, 108, 13, 61, 149,
			232, 79, 186, 18, 42, 210, 58, 66, 106, 96, 5, 225, 136, 204,
			30, 167, 32, 43, 29, 204, 251, 237, 54, 31, 8, 216, 243, 69,
			79, 210, 196, 189, 170, 147, 142, 101, 8, 79, 240, 129, 197, 160,
			223, 233, 132, 92, 200, 34, 102, 195, 15, 128, 171, 179, 150, 135,
			220, 202, 210, 242, 167, 24, 51, 151, 239, 52, 151, 150, 239, 221,
			90, 186, 183, 124, 199, 92, 90, 126, 146, 139, 188, 59, 4, 9,
			39, 65, 119, 96, 133, 130, 130, 92, 41, 233, 251, 222, 168, 154,
			188, 147, 7, 196, 102, 70, 7, 200, 218, 183, 26, 237, 192, 25,
			136, 60, 214, 128, 71, 10, 24, 11, 48, 105, 128, 191, 247, 5,
			111, 11, 85, 251, 96, 65, 165, 156, 93, 249, 163, 116, 255, 80,
			88, 88, 85, 218, 20, 158, 10, 191, 210, 168, 53, 228, 33, 155,
			95, 56, 165, 108, 51, 251, 254, 43, 199, 117, 45, 121, 186, 184,
			183, 216, 106, 20, 108, 191, 29, 22, 118, 249, 94, 97, 196, 74,
			161, 206, 59, 60, 224, 94, 155, 23, 30, 184, 254, 158, 229, 62,
			171, 73, 30, 194, 2, 50, 84, 24, 35, 178, 32, 47, 116, 122,
			190, 109, 162, 48, 42, 210, 228, 229, 57, 87, 44, 193, 115, 172,
			163, 80, 233, 102, 252, 241, 60, 22, 8, 69, 221, 227, 177, 180,
			220, 166, 167, 138, 72, 225, 233, 243, 80, 4, 29, 185, 117, 76,
			34, 191, 29, 154, 3, 21, 217, 80, 150, 149, 130, 235, 236, 5,
			86, 112, 40, 139, 81, 179, 39, 250, 238, 77, 249, 21, 239, 93,
			144, 23, 17, 52, 113, 228, 152, 72, 56, 224, 109, 248, 120, 238,
			241, 226, 92, 127, 113, 206, 110, 206, 109, 222, 155, 219, 190, 55,
			215, 48, 231, 58, 79, 62, 54, 97, 203, 249, 146, 31, 56, 33,
			151, 197, 63, 42, 104, 100, 165, 97, 200, 21, 182, 135, 190, 109,
			73, 103, 253, 56, 132, 167, 207, 43, 141, 90, 156, 234, 55, 84,
			176, 178, 35, 112, 126, 225, 249, 79, 231, 213, 245, 93, 20, 231,
			190, 240, 109, 101, 9, 252, 88, 148, 85, 180, 53, 112, 164, 65,
			226, 81, 85, 91, 43, 94, 11, 39, 113, 75, 57, 99, 2, 115,
			43, 165, 185, 149, 18, 133, 5, 84, 164, 191, 39, 175, 205, 172,
			72, 78, 193, 3, 104, 91, 3, 121, 64, 252, 14, 116, 185, 199,
			3, 75, 29, 181, 248, 152, 133, 42, 44, 39, 250, 55, 169, 252,
			35, 70, 74, 99, 228, 235, 244, 12, 253, 149, 70, 13, 35, 165,
			167, 152, 241, 11, 77, 191, 148, 253, 115, 13, 234, 163, 182, 47,
			118, 125, 191, 35, 61, 94, 170, 56, 116, 188, 246, 120, 233, 65,
			79, 175, 61, 96, 123, 24, 10, 116, 133, 111, 235, 21, 232, 105,
			205, 194, 19, 112, 188, 182, 59, 12, 157, 125, 236, 158, 206, 210,
			9, 100, 111, 66, 242, 55, 21, 131, 26, 130, 233, 243, 49, 72,
			16, 100, 23, 233, 111, 148, 48, 26, 51, 254, 84, 211, 89, 246,
			63, 53, 168, 250, 222, 162, 199, 187, 170, 57, 60, 210, 98, 90,
			113, 43, 133, 221, 213, 233, 45, 102, 53, 218, 152, 116, 93, 251,
			150, 59, 228, 161, 186, 166, 27, 33, 147, 151, 137, 161, 112, 92,
			23, 122, 214, 62, 7, 111, 156, 166, 68, 29, 109, 164, 170, 165,
			81, 93, 107, 199, 15, 176, 91, 140, 91, 234, 227, 10, 139, 58,
			169, 124, 244, 143, 158, 162, 20, 109, 66, 202, 25, 43, 69, 147,
			98, 167, 207, 198, 32, 65, 240, 194, 204, 222, 164, 10, 175, 244,
			191, 111, 211, 69, 199, 235, 4, 86, 193, 26, 12, 184, 215, 117,
			60, 94, 176, 3, 223, 227, 139, 47, 134, 156, 123, 232, 165, 133,
			144, 7, 251, 78, 59, 186, 129, 103, 211, 114, 250, 153, 156, 206,
			190, 235, 69, 32, 247, 11, 157, 178, 58, 31, 248, 129, 40, 225,
			182, 58, 127, 49, 228, 161, 96, 31, 82, 170, 208, 12, 135, 142,
			45, 95, 3, 50, 245, 140, 28, 105, 13, 29, 155, 237, 210, 243,
			174, 111, 217, 207, 162, 168, 237, 7, 234, 101, 96, 122, 197, 52,
			199, 168, 155, 39, 17, 155, 91, 190, 101, 87, 146, 93, 245, 115,
			238, 17, 152, 125, 159, 206, 40, 4, 54, 15, 101, 0, 116, 124,
			111, 150, 72, 242, 23, 228, 68, 105, 52, 206, 24, 53, 122, 206,
			62, 159, 53, 228, 188, 252, 206, 222, 162, 231, 142, 146, 96, 55,
			232, 25, 123, 40, 158, 225, 145, 107, 59, 226, 80, 10, 115, 182,
			62, 109, 15, 197, 122, 52, 148, 251, 87, 157, 94, 60, 194, 107,
			56, 240, 189, 144, 179, 207, 233, 100, 40, 44, 49, 84, 239, 33,
			231, 86, 62, 126, 187, 116, 106, 135, 217, 144, 203, 235, 209, 182,
			99, 106, 212, 143, 171, 113, 157, 158, 231, 47, 7, 78, 32, 91,
			255, 103, 104, 26, 41, 235, 244, 74, 246, 248, 163, 138, 153, 164,
			224, 250, 185, 209, 22, 28, 100, 55, 233, 89, 43, 12, 157, 174,
			199, 237, 103, 246, 80, 132, 179, 6, 144, 249, 76, 253, 76, 60,
			88, 26, 138, 16, 23, 217, 129, 229, 120, 142, 215, 85, 139, 38,
			212, 162, 120, 16, 23, 229, 238, 208, 73, 197, 63, 155, 161, 103,
			91, 213, 31, 85, 107, 187, 213, 103, 229, 122, 189, 86, 191, 144,
			98, 147, 84, 175, 253, 232, 130, 198, 46, 208, 51, 241, 84, 171,
			85, 41, 93, 208, 115, 15, 208, 131, 92, 110, 133, 28, 177, 252,
			158, 30, 196, 168, 33, 249, 208, 37, 31, 242, 59, 247, 30, 90,
			97, 12, 145, 210, 105, 238, 111, 52, 202, 74, 188, 237, 90, 193,
			17, 2, 15, 233, 57, 107, 223, 114, 92, 12, 164, 207, 18, 92,
			211, 43, 55, 143, 24, 233, 228, 70, 179, 52, 20, 245, 179, 201,
			86, 156, 201, 46, 82, 82, 26, 10, 100, 202, 179, 250, 60, 226,
			86, 126, 39, 78, 166, 143, 156, 236, 161, 145, 214, 46, 232, 35,
			166, 143, 208, 136, 152, 190, 72, 103, 182, 156, 80, 121, 71, 76,
			57, 247, 59, 141, 178, 241, 209, 200, 205, 62, 163, 147, 146, 101,
			116, 51, 148, 96, 238, 136, 4, 39, 55, 152, 202, 231, 162, 77,
			217, 95, 105, 116, 66, 142, 176, 115, 84, 79, 116, 173, 159, 238,
			95, 250, 119, 246, 175, 63, 246, 72, 230, 102, 232, 121, 41, 195,
			200, 4, 185, 127, 209, 232, 133, 209, 88, 164, 134, 59, 145, 75,
			40, 37, 220, 56, 169, 132, 177, 197, 210, 136, 114, 121, 214, 85,
			182, 59, 46, 251, 28, 61, 55, 58, 22, 136, 41, 178, 96, 114,
			88, 148, 202, 178, 52, 29, 159, 1, 41, 84, 186, 158, 192, 167,
			9, 179, 242, 79, 137, 178, 119, 232, 244, 88, 4, 96, 215, 223,
			17, 249, 178, 240, 174, 224, 161, 48, 38, 254, 127, 2, 227, 241,
			35, 118, 2, 227, 137, 163, 179, 194, 233, 76, 197, 219, 231, 158,
			240, 131, 195, 29, 245, 94, 21, 32, 153, 49, 143, 61, 70, 230,
			228, 121, 57, 70, 230, 20, 103, 95, 249, 59, 141, 78, 85, 60,
			172, 223, 4, 219, 166, 116, 228, 177, 236, 218, 91, 93, 89, 225,
			190, 254, 14, 87, 103, 15, 104, 58, 182, 61, 187, 250, 22, 151,
			80, 168, 62, 252, 86, 135, 89, 187, 241, 228, 250, 59, 242, 232,
			195, 223, 206, 209, 41, 54, 97, 164, 126, 174, 105, 244, 159, 53,
			249, 158, 108, 164, 216, 202, 175, 181, 35, 79, 195, 203, 171, 178,
			99, 219, 106, 173, 87, 160, 56, 20, 61, 63, 8, 205, 183, 188,
			15, 183, 66, 89, 226, 69, 175, 112, 163, 215, 84, 39, 132, 174,
			191, 207, 3, 15, 187, 89, 207, 142, 30, 7, 139, 3, 171, 141,
			136, 157, 54, 247, 176, 206, 125, 196, 131, 208, 241, 61, 88, 49,
			151, 226, 26, 68, 213, 233, 29, 127, 232, 217, 241, 29, 248, 86,
			101, 189, 92, 109, 148, 161, 227, 184, 88, 100, 100, 168, 78, 82,
			140, 76, 166, 22, 162, 55, 140, 116, 234, 82, 244, 138, 64, 83,
			119, 227, 151, 9, 252, 164, 84, 159, 76, 49, 227, 76, 234, 178,
			134, 181, 229, 36, 214, 150, 103, 210, 103, 233, 63, 104, 212, 152,
			196, 218, 146, 48, 189, 148, 253, 107, 89, 90, 198, 174, 138, 156,
			183, 45, 215, 85, 109, 154, 138, 63, 88, 243, 4, 114, 9, 184,
			206, 62, 247, 120, 168, 158, 6, 186, 92, 64, 169, 213, 164, 160,
			14, 92, 31, 107, 83, 108, 181, 26, 92, 61, 221, 214, 203, 197,
			210, 118, 89, 222, 129, 219, 92, 88, 142, 27, 98, 115, 38, 228,
			3, 129, 39, 176, 78, 27, 61, 98, 75, 74, 178, 100, 163, 209,
			203, 173, 73, 233, 25, 58, 49, 41, 171, 74, 194, 38, 103, 98,
			72, 103, 132, 177, 143, 98, 136, 48, 194, 10, 107, 116, 75, 74,
			164, 49, 242, 158, 94, 202, 126, 14, 99, 39, 229, 237, 2, 201,
			37, 224, 31, 120, 60, 8, 123, 206, 0, 237, 88, 106, 53, 195,
			132, 174, 134, 232, 18, 186, 168, 233, 247, 18, 186, 26, 97, 228,
			189, 194, 154, 84, 177, 198, 140, 217, 212, 85, 165, 98, 220, 51,
			155, 254, 128, 238, 81, 99, 82, 67, 13, 95, 209, 75, 217, 22,
			140, 29, 41, 16, 220, 117, 85, 231, 31, 85, 117, 96, 237, 249,
			67, 1, 150, 235, 42, 87, 226, 146, 13, 72, 242, 151, 44, 200,
			149, 138, 145, 113, 37, 66, 196, 165, 38, 181, 115, 37, 226, 82,
			147, 218, 185, 18, 113, 169, 73, 237, 92, 41, 172, 209, 191, 210,
			168, 62, 169, 51, 3, 82, 55, 181, 236, 47, 53, 136, 78, 114,
			194, 64, 244, 208, 29, 66, 125, 103, 61, 28, 189, 89, 96, 33,
			189, 207, 193, 81, 171, 29, 223, 43, 216, 124, 111, 216, 237, 58,
			94, 215, 148, 47, 15, 33, 87, 59, 162, 242, 58, 121, 106, 129,
			182, 223, 31, 88, 194, 217, 115, 92, 71, 28, 130, 31, 96, 143,
			26, 1, 221, 161, 21, 88, 158, 224, 82, 4, 84, 25, 90, 13,
			210, 231, 233, 52, 53, 38, 117, 84, 217, 13, 189, 40, 249, 215,
			165, 108, 55, 38, 47, 196, 144, 206, 200, 141, 153, 92, 12, 17,
			70, 110, 44, 126, 30, 109, 211, 24, 201, 233, 247, 163, 41, 52,
			66, 110, 242, 92, 12, 233, 140, 228, 206, 95, 139, 33, 194, 72,
			110, 97, 21, 13, 103, 164, 152, 49, 151, 186, 163, 37, 125, 215,
			92, 58, 75, 255, 44, 238, 187, 200, 188, 62, 155, 253, 6, 70,
			21, 14, 58, 18, 26, 7, 107, 34, 136, 83, 140, 106, 163, 35,
			247, 53, 1, 170, 252, 32, 246, 49, 117, 85, 66, 193, 229, 168,
			29, 25, 33, 120, 127, 32, 14, 239, 131, 5, 30, 63, 80, 120,
			14, 176, 59, 217, 227, 111, 193, 39, 109, 172, 218, 44, 50, 175,
			167, 99, 72, 99, 100, 62, 115, 49, 134, 8, 35, 243, 151, 223,
			167, 247, 163, 22, 139, 124, 162, 207, 101, 77, 56, 86, 188, 203,
			11, 41, 249, 235, 130, 142, 124, 6, 180, 108, 216, 179, 92, 203,
			107, 75, 91, 70, 168, 180, 73, 220, 125, 33, 134, 16, 215, 12,
			196, 16, 97, 228, 147, 155, 31, 209, 71, 146, 140, 206, 72, 94,
			191, 158, 173, 192, 137, 186, 65, 222, 231, 65, 111, 216, 183, 60,
			232, 4, 14, 247, 108, 247, 16, 198, 231, 35, 23, 143, 47, 78,
			143, 10, 170, 79, 32, 226, 88, 80, 148, 38, 159, 201, 198, 16,
			97, 36, 255, 33, 218, 209, 48, 82, 36, 197, 140, 69, 125, 153,
			168, 57, 130, 42, 89, 164, 179, 52, 164, 147, 8, 161, 249, 150,
			140, 171, 89, 27, 198, 251, 2, 197, 90, 232, 200, 43, 93, 169,
			130, 68, 63, 42, 14, 141, 238, 229, 122, 254, 1, 244, 45, 239,
			144, 130, 240, 133, 229, 170, 3, 57, 10, 83, 24, 165, 195, 225,
			0, 35, 162, 73, 233, 57, 58, 165, 136, 78, 32, 213, 49, 88,
			99, 100, 105, 250, 253, 17, 76, 24, 89, 202, 94, 161, 127, 161,
			92, 140, 48, 114, 91, 103, 217, 255, 175, 1, 150, 29, 170, 21,
			149, 214, 25, 209, 177, 186, 220, 147, 23, 176, 142, 12, 99, 137,
			253, 74, 173, 102, 33, 90, 209, 233, 56, 158, 35, 14, 77, 170,
			120, 148, 45, 112, 104, 245, 249, 56, 210, 211, 157, 204, 9, 143,
			41, 159, 76, 32, 71, 177, 242, 137, 198, 200, 237, 204, 217, 24,
			66, 110, 47, 204, 200, 99, 163, 49, 227, 110, 170, 170, 142, 13,
			58, 201, 221, 244, 21, 106, 81, 195, 144, 241, 110, 85, 191, 148,
			109, 130, 106, 142, 162, 164, 17, 5, 59, 53, 20, 155, 223, 114,
			93, 19, 160, 34, 47, 146, 157, 62, 46, 179, 60, 121, 239, 214,
			238, 241, 246, 151, 52, 185, 27, 1, 30, 4, 152, 127, 21, 147,
			154, 158, 154, 68, 26, 233, 24, 210, 24, 89, 205, 156, 143, 33,
			194, 200, 42, 187, 40, 61, 68, 195, 211, 125, 79, 95, 83, 30,
			162, 201, 243, 125, 111, 234, 44, 253, 185, 78, 39, 17, 68, 94,
			63, 55, 46, 103, 127, 171, 193, 145, 62, 40, 190, 225, 244, 124,
			145, 252, 32, 199, 243, 131, 190, 229, 186, 135, 9, 195, 210, 66,
			188, 99, 13, 93, 65, 35, 29, 59, 157, 113, 41, 157, 16, 228,
			15, 109, 188, 46, 6, 191, 161, 247, 165, 231, 31, 120, 38, 28,
			189, 232, 84, 91, 104, 18, 133, 135, 33, 15, 163, 216, 192, 189,
			97, 63, 66, 156, 100, 200, 182, 235, 200, 3, 227, 243, 80, 114,
			135, 56, 105, 148, 59, 14, 121, 244, 36, 16, 45, 146, 22, 31,
			134, 124, 156, 83, 133, 47, 242, 87, 45, 138, 35, 159, 27, 51,
			35, 88, 103, 228, 243, 75, 239, 209, 179, 145, 134, 52, 70, 126,
			104, 76, 39, 211, 154, 132, 39, 71, 176, 206, 200, 15, 51, 52,
			89, 174, 51, 82, 52, 222, 75, 166, 113, 123, 209, 184, 48, 130,
			113, 254, 226, 37, 250, 183, 154, 116, 21, 141, 145, 13, 125, 54,
			251, 151, 218, 119, 141, 176, 149, 206, 248, 142, 3, 43, 68, 5,
			138, 184, 86, 10, 84, 169, 24, 253, 58, 172, 227, 112, 215, 86,
			202, 136, 126, 162, 213, 83, 177, 152, 171, 51, 162, 52, 236, 7,
			20, 77, 237, 171, 31, 216, 37, 158, 166, 77, 32, 139, 177, 167,
			161, 244, 27, 81, 208, 213, 100, 52, 220, 184, 252, 62, 221, 144,
			178, 232, 140, 108, 234, 75, 217, 85, 56, 214, 138, 161, 60, 242,
			170, 125, 60, 224, 141, 106, 37, 181, 156, 143, 124, 91, 159, 68,
			68, 87, 98, 72, 99, 100, 243, 234, 247, 99, 136, 48, 178, 105,
			22, 232, 15, 37, 69, 194, 200, 67, 253, 163, 236, 45, 56, 114,
			47, 32, 131, 252, 168, 126, 248, 150, 148, 162, 233, 196, 64, 20,
			9, 52, 193, 200, 195, 233, 153, 24, 210, 24, 121, 200, 174, 199,
			16, 18, 203, 221, 164, 129, 164, 108, 48, 178, 173, 127, 148, 69,
			116, 99, 151, 13, 71, 41, 31, 43, 234, 162, 19, 37, 55, 152,
			16, 71, 51, 26, 191, 85, 88, 16, 14, 247, 208, 132, 126, 231,
			168, 56, 9, 175, 134, 36, 154, 64, 19, 140, 108, 39, 188, 26,
			26, 35, 219, 9, 175, 6, 97, 100, 59, 119, 83, 134, 41, 157,
			25, 59, 169, 71, 42, 76, 161, 46, 119, 210, 89, 250, 3, 106,
			24, 178, 198, 168, 235, 179, 217, 194, 119, 115, 61, 69, 95, 151,
			97, 190, 30, 249, 133, 42, 81, 234, 145, 95, 168, 162, 164, 126,
			249, 125, 250, 84, 210, 209, 24, 105, 233, 87, 178, 85, 144, 42,
			26, 213, 156, 73, 28, 193, 99, 108, 121, 42, 196, 97, 56, 176,
			80, 127, 201, 196, 136, 11, 122, 10, 27, 154, 129, 216, 19, 104,
			130, 145, 86, 164, 20, 85, 0, 181, 216, 229, 24, 34, 140, 180,
			62, 200, 98, 103, 128, 250, 217, 77, 93, 147, 58, 65, 43, 239,
			166, 175, 72, 93, 25, 204, 120, 156, 114, 148, 174, 80, 163, 143,
			211, 89, 250, 127, 241, 59, 195, 200, 83, 253, 108, 118, 69, 137,
			128, 41, 131, 15, 2, 46, 159, 113, 76, 144, 221, 207, 209, 27,
			26, 44, 22, 5, 183, 240, 20, 77, 83, 195, 48, 50, 41, 70,
			158, 78, 159, 145, 156, 24, 25, 84, 214, 24, 164, 43, 72, 18,
			165, 140, 252, 68, 103, 106, 19, 77, 49, 242, 19, 41, 140, 97,
			24, 152, 233, 127, 166, 219, 42, 142, 27, 50, 211, 255, 140, 158,
			165, 55, 232, 36, 66, 104, 203, 231, 198, 165, 44, 75, 126, 120,
			25, 121, 97, 20, 231, 140, 40, 47, 63, 55, 198, 96, 141, 145,
			231, 211, 231, 71, 48, 97, 228, 57, 187, 72, 255, 68, 139, 112,
			106, 140, 180, 141, 75, 89, 49, 158, 67, 199, 48, 195, 239, 153,
			144, 155, 106, 189, 44, 59, 198, 60, 202, 138, 142, 197, 105, 169,
			122, 140, 107, 180, 104, 123, 140, 107, 180, 105, 123, 140, 107, 180,
			106, 155, 93, 196, 54, 214, 48, 12, 212, 67, 79, 207, 101, 255,
			81, 59, 97, 16, 60, 97, 241, 207, 100, 229, 241, 236, 91, 54,
			63, 210, 93, 196, 45, 133, 116, 75, 108, 205, 44, 199, 11, 199,
			187, 58, 112, 60, 245, 76, 129, 5, 28, 202, 107, 69, 138, 144,
			248, 162, 192, 169, 46, 200, 71, 191, 202, 141, 170, 14, 170, 14,
			62, 183, 101, 203, 104, 115, 151, 143, 130, 172, 161, 167, 12, 228,
			59, 129, 38, 25, 233, 77, 159, 139, 33, 141, 145, 222, 249, 15,
			99, 136, 48, 210, 3, 249, 43, 57, 140, 0, 95, 68, 94, 60,
			161, 49, 242, 69, 250, 138, 28, 158, 100, 196, 77, 93, 149, 195,
			147, 26, 35, 110, 250, 3, 233, 220, 83, 204, 232, 167, 134, 202,
			185, 167, 52, 70, 250, 233, 172, 116, 173, 41, 116, 45, 79, 15,
			149, 107, 77, 73, 215, 242, 232, 121, 153, 208, 166, 148, 107, 249,
			6, 147, 10, 159, 138, 220, 200, 143, 12, 50, 21, 185, 145, 63,
			125, 118, 4, 19, 70, 252, 11, 51, 201, 118, 141, 145, 129, 177,
			146, 76, 99, 113, 61, 48, 62, 28, 193, 56, 127, 109, 113, 4,
			19, 70, 6, 75, 203, 201, 118, 157, 145, 23, 198, 141, 100, 26,
			43, 227, 23, 99, 212, 17, 253, 139, 233, 171, 35, 152, 48, 242,
			226, 58, 36, 219, 9, 35, 129, 113, 41, 153, 198, 0, 31, 140,
			109, 199, 195, 31, 68, 222, 36, 97, 92, 207, 46, 202, 243, 55,
			133, 146, 11, 253, 170, 82, 139, 52, 145, 136, 76, 52, 37, 77,
			36, 166, 47, 196, 144, 198, 136, 152, 121, 63, 134, 8, 35, 34,
			171, 108, 145, 102, 228, 32, 149, 149, 58, 79, 107, 140, 28, 164,
			223, 167, 211, 84, 55, 50, 108, 226, 165, 188, 125, 193, 137, 140,
			198, 200, 203, 244, 172, 52, 70, 6, 141, 113, 168, 127, 173, 140,
			145, 145, 198, 56, 164, 103, 165, 60, 25, 101, 140, 87, 145, 49,
			50, 145, 49, 94, 69, 242, 100, 34, 99, 188, 138, 140, 145, 137,
			140, 241, 42, 50, 70, 70, 25, 227, 181, 113, 45, 153, 198, 195,
			245, 122, 108, 59, 26, 227, 245, 244, 7, 35, 152, 48, 242, 250,
			234, 135, 201, 118, 157, 145, 175, 140, 203, 201, 52, 26, 227, 43,
			35, 61, 130, 53, 70, 190, 202, 204, 140, 96, 194, 200, 87, 81,
			37, 149, 81, 198, 120, 19, 25, 35, 19, 25, 227, 205, 24, 117,
			52, 198, 155, 200, 24, 153, 200, 24, 111, 34, 99, 100, 80, 242,
			111, 244, 89, 165, 22, 105, 140, 111, 34, 99, 100, 164, 49, 190,
			137, 206, 75, 70, 234, 224, 155, 243, 23, 99, 136, 48, 242, 205,
			229, 247, 227, 103, 166, 255, 13, 0, 0, 255, 255, 195, 197, 224,
			80, 61, 50, 0, 0},
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
	ret, err := discovery.GetDescriptorSet("drone_queen.Drone")
	if err != nil {
		panic(err)
	}
	return ret
}