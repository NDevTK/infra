// Code generated by cproto. DO NOT EDIT.

package adminpb

import "go.chromium.org/luci/grpc/discovery"

import "google.golang.org/protobuf/types/descriptorpb"

func init() {
	discovery.RegisterDescriptorSetCompressed(
		[]string{
			"weetbix.internal.admin.Admin",
		},
		[]byte{31, 139,
			8, 0, 0, 0, 0, 0, 0, 255, 236, 123, 77, 140, 27, 73,
			150, 30, 147, 201, 170, 46, 69, 75, 45, 41, 165, 233, 213, 176,
			213, 210, 107, 142, 74, 34, 101, 86, 178, 126, 244, 91, 26, 13,
			134, 197, 202, 146, 82, 83, 34, 107, 72, 86, 105, 164, 70, 183,
			148, 204, 12, 146, 49, 74, 102, 176, 51, 130, 69, 177, 101, 1,
			134, 1, 99, 129, 93, 96, 97, 27, 48, 96, 31, 108, 3, 246,
			201, 139, 5, 124, 48, 96, 248, 176, 192, 194, 128, 13, 95, 22,
			240, 205, 7, 31, 236, 147, 111, 134, 125, 53, 96, 195, 48, 226,
			69, 38, 201, 146, 74, 82, 207, 216, 48, 176, 64, 23, 84, 221,
			25, 127, 47, 222, 123, 241, 226, 189, 239, 69, 68, 145, 191, 176,
			200, 23, 61, 206, 123, 33, 173, 12, 99, 46, 121, 103, 212, 173,
			208, 193, 80, 78, 108, 44, 90, 167, 117, 163, 157, 54, 22, 62,
			33, 11, 142, 106, 223, 122, 67, 206, 249, 124, 96, 191, 213, 190,
			69, 176, 117, 79, 21, 247, 140, 103, 105, 115, 143, 135, 94, 212,
			179, 121, 220, 155, 77, 35, 39, 67, 42, 42, 47, 35, 62, 142,
			244, 148, 195, 206, 255, 48, 140, 127, 146, 53, 31, 236, 109, 253,
			105, 246, 210, 3, 61, 114, 47, 233, 110, 63, 161, 97, 248, 43,
			213, 185, 173, 198, 61, 250, 223, 103, 200, 162, 149, 187, 148, 217,
			56, 67, 254, 242, 36, 49, 78, 90, 230, 165, 140, 181, 254, 175,
			79, 2, 14, 240, 121, 8, 91, 163, 110, 151, 198, 2, 86, 64,
			147, 186, 38, 32, 240, 164, 7, 44, 146, 52, 246, 251, 94, 212,
			163, 208, 229, 241, 192, 147, 4, 106, 124, 56, 137, 89, 175, 47,
			97, 125, 117, 245, 78, 50, 0, 220, 200, 183, 1, 170, 97, 8,
			216, 38, 32, 166, 130, 198, 135, 52, 176, 9, 244, 165, 28, 138,
			205, 74, 37, 160, 135, 52, 228, 67, 26, 139, 84, 86, 159, 15,
			180, 144, 62, 15, 87, 58, 154, 137, 10, 33, 208, 164, 1, 19,
			50, 102, 157, 145, 100, 60, 2, 47, 10, 96, 36, 40, 176, 8,
			4, 31, 197, 62, 197, 154, 14, 139, 188, 120, 130, 124, 137, 50,
			140, 153, 236, 3, 143, 241, 255, 124, 36, 9, 12, 120, 192, 186,
			204, 247, 20, 133, 50, 120, 49, 133, 33, 141, 7, 76, 74, 26,
			192, 48, 230, 135, 44, 160, 1, 200, 190, 39, 65, 246, 149, 116,
			97, 200, 199, 44, 234, 129, 207, 163, 128, 169, 65, 66, 13, 34,
			48, 160, 114, 147, 16, 80, 63, 215, 223, 98, 76, 0, 239, 166,
			28, 249, 60, 160, 48, 24, 9, 9, 49, 149, 30, 139, 144, 170,
			215, 225, 135, 170, 41, 209, 24, 129, 136, 75, 230, 211, 50, 200,
			62, 19, 16, 50, 33, 21, 133, 249, 25, 163, 224, 45, 118, 2,
			38, 252, 208, 99, 3, 26, 219, 239, 99, 130, 69, 243, 186, 72,
			153, 24, 198, 60, 24, 249, 116, 198, 7, 153, 49, 242, 127, 197,
			7, 129, 68, 186, 128, 251, 163, 1, 141, 164, 151, 46, 82, 133,
			199, 192, 101, 159, 198, 48, 240, 36, 141, 153, 23, 138, 153, 170,
			113, 129, 100, 159, 18, 152, 231, 126, 42, 84, 157, 50, 28, 169,
			8, 71, 222, 128, 42, 134, 230, 109, 43, 226, 179, 54, 212, 59,
			147, 66, 73, 20, 105, 82, 60, 22, 48, 240, 38, 208, 161, 202,
			82, 2, 144, 28, 104, 20, 240, 88, 80, 101, 20, 195, 152, 15,
			184, 164, 160, 117, 34, 5, 4, 52, 102, 135, 52, 128, 110, 204,
			7, 68, 107, 65, 240, 174, 28, 43, 51, 73, 44, 8, 196, 144,
			250, 202, 130, 96, 24, 51, 101, 88, 177, 178, 157, 72, 91, 145,
			16, 200, 59, 129, 246, 67, 183, 5, 173, 198, 78, 251, 73, 181,
			233, 128, 219, 130, 189, 102, 227, 192, 221, 118, 182, 97, 235, 41,
			180, 31, 58, 80, 107, 236, 61, 109, 186, 15, 30, 182, 225, 97,
			99, 119, 219, 105, 182, 160, 90, 223, 134, 90, 163, 222, 110, 186,
			91, 251, 237, 70, 179, 69, 160, 80, 109, 129, 219, 42, 96, 75,
			181, 254, 20, 156, 223, 236, 53, 157, 86, 11, 26, 77, 112, 31,
			239, 237, 186, 206, 54, 60, 169, 54, 155, 213, 122, 219, 117, 90,
			101, 112, 235, 181, 221, 253, 109, 183, 254, 160, 12, 91, 251, 109,
			168, 55, 218, 4, 118, 221, 199, 110, 219, 217, 134, 118, 163, 140,
			211, 190, 59, 14, 26, 59, 240, 216, 105, 214, 30, 86, 235, 237,
			234, 150, 187, 235, 182, 159, 226, 132, 59, 110, 187, 174, 38, 219,
			105, 52, 9, 84, 97, 175, 218, 108, 187, 181, 253, 221, 106, 19,
			246, 246, 155, 123, 141, 150, 3, 74, 178, 109, 183, 85, 219, 173,
			186, 143, 157, 109, 27, 220, 58, 212, 27, 224, 28, 56, 245, 54,
			180, 30, 86, 119, 119, 143, 10, 74, 160, 241, 164, 238, 52, 21,
			247, 243, 98, 194, 150, 3, 187, 110, 117, 107, 215, 81, 83, 161,
			156, 219, 110, 211, 169, 181, 149, 64, 179, 175, 154, 187, 237, 212,
			219, 213, 221, 50, 129, 214, 158, 83, 115, 171, 187, 101, 112, 126,
			227, 60, 222, 219, 173, 54, 159, 150, 19, 162, 45, 231, 215, 251,
			78, 189, 237, 86, 119, 97, 187, 250, 184, 250, 192, 105, 65, 241,
			99, 90, 217, 107, 54, 106, 251, 77, 231, 177, 226, 186, 177, 3,
			173, 253, 173, 86, 219, 109, 239, 183, 29, 120, 208, 104, 108, 163,
			178, 91, 78, 243, 192, 173, 57, 173, 123, 176, 219, 104, 161, 194,
			246, 91, 78, 153, 192, 118, 181, 93, 197, 169, 247, 154, 141, 29,
			183, 221, 186, 167, 190, 183, 246, 91, 46, 42, 206, 173, 183, 157,
			102, 115, 127, 175, 237, 54, 234, 37, 120, 216, 120, 226, 28, 56,
			77, 168, 85, 247, 91, 206, 54, 106, 184, 81, 87, 210, 42, 91,
			113, 26, 205, 167, 138, 172, 210, 3, 174, 64, 25, 158, 60, 116,
			218, 15, 157, 166, 82, 42, 106, 171, 170, 212, 208, 106, 55, 221,
			90, 123, 190, 91, 163, 9, 237, 70, 179, 77, 230, 228, 132, 186,
			243, 96, 215, 125, 224, 212, 107, 142, 106, 110, 40, 50, 79, 220,
			150, 83, 130, 106, 211, 109, 169, 14, 46, 78, 12, 79, 170, 79,
			161, 177, 143, 82, 171, 133, 218, 111, 57, 68, 127, 207, 153, 110,
			25, 215, 19, 220, 29, 168, 110, 31, 184, 138, 243, 164, 247, 94,
			163, 213, 114, 19, 115, 65, 181, 213, 30, 38, 58, 183, 9, 89,
			34, 70, 214, 50, 33, 115, 65, 125, 45, 89, 102, 33, 115, 143,
			156, 32, 217, 165, 101, 253, 169, 43, 127, 150, 113, 176, 242, 83,
			253, 169, 43, 175, 100, 202, 88, 105, 232, 79, 93, 185, 156, 249,
			107, 88, 153, 124, 234, 202, 171, 153, 2, 86, 18, 253, 169, 43,
			175, 101, 190, 194, 202, 43, 250, 83, 87, 22, 51, 151, 177, 242,
			178, 254, 252, 159, 89, 146, 205, 101, 44, 115, 35, 115, 38, 255,
			223, 179, 80, 133, 30, 141, 104, 204, 124, 192, 8, 10, 3, 42,
			132, 215, 163, 58, 4, 76, 248, 8, 124, 47, 130, 152, 174, 168,
			64, 35, 57, 120, 135, 156, 5, 16, 208, 46, 139, 208, 253, 141,
			134, 161, 10, 38, 52, 32, 71, 199, 163, 251, 157, 240, 81, 12,
			213, 61, 87, 216, 80, 5, 57, 25, 50, 223, 11, 129, 190, 242,
			6, 195, 144, 2, 19, 138, 30, 198, 47, 9, 158, 64, 47, 22,
			211, 239, 70, 84, 72, 2, 137, 87, 139, 169, 24, 242, 72, 205,
			60, 25, 162, 235, 243, 34, 69, 79, 5, 159, 62, 15, 108, 216,
			225, 49, 176, 72, 72, 47, 242, 105, 26, 141, 84, 124, 101, 62,
			133, 29, 206, 225, 181, 174, 2, 136, 135, 62, 108, 121, 113, 241,
			45, 172, 97, 35, 212, 40, 169, 216, 52, 138, 35, 1, 239, 105,
			191, 167, 201, 188, 81, 142, 173, 79, 225, 81, 171, 81, 199, 72,
			66, 197, 212, 205, 119, 121, 12, 47, 176, 247, 11, 37, 153, 214,
			5, 118, 228, 157, 223, 82, 95, 194, 139, 215, 111, 94, 216, 132,
			16, 98, 230, 50, 134, 101, 110, 44, 157, 234, 44, 226, 52, 27,
			228, 223, 173, 145, 203, 111, 35, 40, 201, 6, 84, 72, 111, 48,
			124, 31, 138, 186, 71, 78, 180, 211, 62, 214, 5, 242, 137, 160,
			42, 78, 137, 11, 6, 24, 69, 179, 153, 22, 173, 243, 100, 33,
			242, 34, 46, 46, 100, 193, 40, 46, 52, 117, 97, 235, 111, 25,
			199, 67, 175, 207, 166, 36, 83, 248, 181, 254, 3, 225, 215, 148,
			223, 223, 9, 130, 253, 219, 10, 249, 196, 90, 184, 148, 249, 187,
			134, 241, 35, 6, 251, 17, 131, 253, 136, 193, 126, 196, 96, 63,
			98, 176, 31, 49, 216, 255, 71, 12, 54, 133, 70, 234, 51, 197,
			96, 110, 10, 204, 212, 103, 138, 193, 166, 192, 108, 121, 10, 204,
			174, 102, 42, 41, 48, 83, 159, 41, 6, 155, 2, 179, 107, 83,
			96, 86, 156, 1, 51, 245, 249, 159, 190, 68, 12, 182, 248, 135,
			134, 10, 125, 249, 127, 255, 37, 84, 97, 26, 122, 103, 208, 66,
			128, 7, 67, 206, 34, 137, 110, 141, 13, 84, 152, 9, 232, 144,
			70, 1, 141, 164, 134, 67, 19, 93, 255, 61, 143, 208, 155, 132,
			220, 247, 66, 2, 190, 23, 210, 40, 240, 226, 50, 208, 72, 121,
			255, 64, 1, 44, 15, 124, 62, 210, 227, 18, 116, 128, 190, 180,
			27, 123, 254, 44, 98, 164, 13, 42, 32, 40, 168, 128, 101, 21,
			49, 121, 168, 157, 34, 34, 32, 77, 136, 169, 80, 26, 122, 146,
			29, 106, 104, 24, 1, 29, 114, 191, 15, 158, 132, 253, 118, 13,
			6, 44, 136, 208, 163, 243, 136, 192, 35, 47, 26, 169, 48, 176,
			86, 134, 181, 187, 183, 87, 203, 169, 163, 30, 198, 60, 164, 67,
			201, 124, 120, 16, 211, 30, 143, 153, 23, 77, 185, 135, 113, 159,
			249, 125, 160, 175, 36, 85, 60, 161, 131, 62, 166, 87, 199, 243,
			95, 142, 189, 56, 64, 60, 57, 161, 94, 12, 60, 162, 202, 1,
			170, 144, 63, 96, 209, 72, 82, 140, 151, 112, 107, 117, 42, 95,
			200, 163, 158, 13, 187, 212, 27, 206, 68, 142, 41, 20, 196, 128,
			122, 49, 13, 10, 32, 184, 14, 192, 17, 135, 144, 122, 67, 146,
			116, 3, 233, 117, 52, 118, 141, 40, 85, 122, 237, 34, 2, 149,
			52, 30, 170, 216, 170, 3, 250, 72, 168, 168, 228, 193, 215, 235,
			55, 86, 250, 10, 2, 135, 44, 162, 94, 76, 0, 169, 127, 83,
			252, 48, 232, 80, 235, 89, 193, 158, 37, 59, 1, 156, 49, 162,
			28, 38, 48, 38, 192, 234, 234, 234, 218, 10, 254, 107, 175, 174,
			110, 226, 191, 103, 74, 244, 187, 119, 239, 222, 93, 89, 91, 95,
			217, 88, 107, 175, 111, 108, 222, 188, 187, 121, 243, 174, 125, 55,
			253, 121, 102, 195, 214, 132, 168, 133, 148, 49, 243, 165, 98, 80,
			38, 34, 34, 245, 50, 140, 41, 208, 72, 140, 226, 4, 250, 143,
			41, 34, 127, 159, 71, 135, 52, 150, 122, 125, 117, 80, 130, 175,
			155, 59, 53, 2, 27, 27, 27, 119, 103, 178, 140, 199, 99, 155,
			81, 217, 69, 132, 24, 119, 125, 245, 171, 122, 216, 242, 149, 44,
			41, 196, 70, 65, 205, 28, 245, 132, 18, 234, 103, 224, 232, 44,
			64, 16, 146, 126, 194, 218, 38, 212, 248, 96, 56, 146, 116, 110,
			47, 224, 132, 123, 141, 150, 251, 27, 120, 161, 52, 83, 44, 41,
			20, 141, 129, 121, 214, 105, 10, 62, 19, 160, 62, 3, 207, 130,
			202, 231, 201, 2, 23, 113, 120, 125, 127, 119, 183, 84, 58, 182,
			31, 218, 123, 113, 181, 116, 111, 142, 167, 245, 143, 241, 212, 163,
			82, 81, 225, 221, 192, 155, 204, 241, 38, 100, 60, 242, 37, 78,
			112, 232, 133, 32, 15, 147, 25, 143, 116, 191, 42, 15, 203, 128,
			12, 221, 251, 125, 69, 58, 180, 229, 161, 42, 125, 72, 34, 221,
			105, 36, 168, 15, 215, 97, 109, 117, 245, 168, 132, 27, 239, 149,
			240, 9, 139, 54, 214, 225, 197, 3, 42, 91, 19, 33, 233, 64,
			53, 87, 197, 14, 11, 105, 251, 232, 66, 236, 184, 187, 78, 219,
			125, 236, 64, 87, 38, 108, 188, 111, 204, 213, 174, 76, 57, 221,
			119, 235, 237, 91, 55, 64, 50, 255, 165, 128, 251, 80, 44, 22,
			117, 77, 169, 43, 237, 96, 252, 144, 245, 250, 219, 158, 196, 81,
			37, 248, 249, 207, 97, 99, 189, 4, 127, 29, 176, 109, 151, 143,
			211, 166, 84, 111, 149, 10, 84, 21, 191, 1, 31, 11, 36, 169,
			54, 203, 218, 234, 234, 156, 15, 19, 246, 180, 131, 246, 82, 107,
			183, 222, 221, 70, 83, 106, 106, 248, 218, 173, 27, 55, 110, 220,
			222, 184, 181, 58, 115, 27, 29, 218, 229, 49, 133, 253, 136, 189,
			74, 169, 220, 189, 189, 250, 54, 21, 251, 247, 91, 204, 162, 150,
			31, 138, 69, 173, 148, 10, 46, 150, 250, 41, 193, 202, 60, 59,
			31, 177, 96, 69, 71, 169, 43, 165, 179, 60, 71, 7, 13, 160,
			116, 196, 0, 110, 188, 215, 0, 30, 121, 135, 30, 188, 208, 11,
			105, 251, 163, 56, 166, 145, 84, 93, 30, 179, 48, 100, 98, 206,
			0, 148, 55, 133, 1, 214, 194, 125, 120, 255, 128, 15, 152, 57,
			220, 159, 213, 218, 17, 29, 111, 141, 88, 24, 208, 184, 88, 82,
			130, 181, 18, 13, 37, 83, 104, 197, 148, 210, 220, 30, 64, 245,
			169, 107, 217, 89, 36, 149, 228, 73, 79, 45, 122, 34, 54, 106,
			160, 100, 119, 20, 101, 228, 101, 166, 131, 155, 31, 209, 129, 139,
			103, 12, 210, 142, 248, 120, 78, 236, 164, 22, 34, 62, 134, 251,
			112, 164, 207, 7, 37, 157, 49, 254, 113, 145, 35, 62, 182, 123,
			84, 58, 202, 216, 116, 93, 177, 52, 39, 249, 81, 233, 147, 206,
			170, 80, 124, 143, 164, 183, 222, 43, 105, 178, 94, 41, 206, 128,
			189, 137, 236, 235, 68, 226, 136, 161, 205, 47, 84, 177, 244, 182,
			21, 62, 160, 178, 54, 91, 247, 98, 9, 125, 61, 30, 131, 60,
			246, 134, 67, 22, 245, 8, 1, 55, 210, 53, 58, 107, 47, 35,
			12, 152, 211, 211, 100, 136, 161, 238, 8, 112, 209, 161, 35, 193,
			12, 4, 3, 208, 239, 20, 127, 244, 84, 10, 187, 120, 10, 182,
			148, 53, 25, 93, 171, 38, 43, 188, 86, 184, 225, 205, 202, 235,
			1, 143, 100, 255, 205, 202, 235, 192, 155, 188, 105, 191, 86, 193,
			251, 205, 230, 235, 1, 139, 222, 108, 190, 22, 212, 127, 243, 181,
			253, 90, 193, 37, 181, 101, 223, 124, 243, 172, 64, 96, 220, 167,
			49, 5, 61, 90, 17, 242, 194, 177, 55, 17, 64, 95, 41, 4,
			167, 146, 61, 141, 5, 186, 10, 5, 4, 172, 199, 164, 80, 160,
			38, 164, 144, 204, 84, 6, 156, 170, 76, 64, 79, 86, 6, 156,
			173, 140, 193, 22, 167, 68, 92, 242, 61, 141, 249, 202, 208, 11,
			2, 157, 62, 202, 49, 79, 169, 81, 207, 239, 107, 76, 150, 226,
			56, 133, 255, 18, 151, 82, 78, 16, 148, 10, 228, 61, 14, 163,
			33, 194, 132, 116, 104, 145, 217, 212, 78, 42, 215, 142, 71, 123,
			165, 50, 193, 249, 249, 80, 83, 214, 51, 21, 158, 21, 64, 140,
			186, 93, 246, 74, 225, 81, 60, 254, 211, 199, 119, 202, 14, 16,
			137, 22, 11, 251, 237, 90, 161, 116, 239, 72, 45, 209, 128, 241,
			187, 17, 139, 105, 96, 67, 21, 244, 241, 151, 54, 6, 129, 57,
			57, 251, 158, 198, 32, 250, 124, 20, 6, 169, 42, 71, 130, 34,
			154, 44, 122, 98, 58, 91, 0, 157, 9, 81, 108, 148, 212, 2,
			68, 42, 11, 142, 52, 164, 121, 215, 148, 148, 34, 189, 35, 83,
			13, 189, 88, 204, 166, 233, 80, 2, 136, 233, 20, 194, 241, 125,
			58, 148, 208, 225, 178, 143, 115, 170, 177, 250, 208, 32, 149, 65,
			188, 195, 135, 130, 189, 188, 219, 21, 84, 34, 92, 219, 225, 113,
			122, 194, 89, 134, 194, 250, 234, 218, 109, 21, 29, 214, 110, 182,
			87, 215, 54, 55, 86, 55, 215, 110, 218, 171, 107, 207, 10, 137,
			117, 11, 192, 242, 52, 188, 12, 61, 33, 9, 96, 79, 156, 159,
			71, 51, 220, 124, 179, 12, 138, 154, 157, 108, 32, 239, 208, 107,
			249, 49, 27, 202, 178, 66, 187, 71, 160, 154, 7, 42, 60, 166,
			231, 142, 136, 242, 20, 116, 212, 198, 174, 237, 17, 205, 95, 185,
			171, 192, 139, 3, 2, 95, 75, 238, 182, 26, 45, 220, 100, 197,
			210, 49, 0, 213, 30, 240, 239, 89, 24, 122, 184, 187, 104, 180,
			178, 223, 170, 4, 220, 23, 149, 39, 180, 83, 153, 177, 82, 105,
			210, 46, 141, 105, 228, 211, 202, 131, 144, 119, 188, 240, 121, 3,
			121, 16, 21, 197, 80, 101, 110, 146, 18, 153, 30, 225, 186, 169,
			167, 41, 227, 62, 215, 44, 193, 11, 133, 24, 149, 210, 237, 244,
			227, 69, 42, 144, 18, 181, 67, 83, 105, 105, 64, 142, 21, 145,
			192, 215, 47, 132, 140, 187, 56, 116, 78, 34, 238, 11, 123, 168,
			61, 155, 146, 101, 189, 18, 178, 78, 236, 197, 19, 132, 221, 118,
			95, 14, 194, 159, 225, 87, 58, 182, 132, 103, 46, 100, 106, 200,
			233, 36, 98, 72, 125, 184, 182, 252, 116, 101, 121, 176, 178, 28,
			180, 151, 31, 110, 46, 63, 222, 92, 110, 217, 203, 221, 103, 215,
			108, 216, 101, 47, 233, 152, 9, 138, 105, 142, 82, 208, 108, 149,
			70, 130, 106, 106, 143, 120, 224, 161, 177, 94, 19, 240, 245, 11,
			183, 213, 72, 65, 205, 142, 118, 86, 65, 82, 44, 150, 94, 124,
			83, 212, 39, 149, 137, 159, 251, 45, 15, 244, 74, 168, 143, 21,
			204, 23, 188, 33, 195, 5, 73, 107, 117, 22, 161, 121, 173, 188,
			75, 27, 229, 76, 39, 88, 94, 223, 94, 94, 223, 38, 80, 82,
			138, 228, 29, 60, 33, 244, 18, 57, 37, 141, 193, 247, 134, 184,
			65, 120, 87, 95, 21, 120, 122, 171, 165, 219, 76, 104, 183, 60,
			213, 63, 30, 114, 127, 170, 143, 185, 115, 127, 104, 44, 157, 37,
			255, 208, 32, 185, 92, 38, 155, 177, 114, 127, 108, 100, 207, 231,
			255, 196, 128, 230, 44, 195, 77, 109, 159, 119, 209, 228, 81, 199,
			130, 69, 254, 60, 202, 34, 199, 195, 44, 120, 60, 18, 82, 217,
			194, 135, 210, 34, 114, 92, 94, 244, 12, 88, 228, 135, 35, 193,
			14, 85, 162, 120, 138, 44, 40, 246, 22, 144, 191, 79, 210, 162,
			161, 138, 75, 167, 211, 162, 169, 138, 214, 57, 242, 95, 180, 48,
			134, 149, 251, 59, 70, 214, 202, 255, 7, 3, 234, 60, 90, 137,
			104, 79, 231, 193, 71, 178, 105, 47, 205, 26, 85, 34, 121, 124,
			54, 93, 79, 6, 78, 19, 204, 67, 47, 28, 81, 161, 143, 36,
			103, 196, 240, 224, 84, 72, 22, 134, 208, 247, 14, 41, 68, 243,
			115, 34, 233, 100, 32, 209, 217, 155, 78, 208, 187, 60, 86, 137,
			113, 122, 122, 240, 182, 194, 146, 164, 177, 156, 252, 146, 99, 148,
			98, 44, 160, 156, 169, 82, 12, 20, 123, 233, 84, 90, 52, 85,
			241, 204, 217, 233, 77, 198, 191, 248, 138, 172, 176, 168, 27, 123,
			21, 111, 56, 164, 81, 143, 69, 180, 50, 166, 84, 118, 216, 43,
			125, 152, 94, 57, 92, 171, 248, 124, 48, 224, 81, 114, 175, 65,
			146, 102, 251, 112, 45, 255, 177, 75, 144, 194, 88, 223, 121, 52,
			85, 194, 106, 221, 34, 75, 212, 139, 67, 70, 133, 196, 75, 143,
			79, 215, 243, 111, 223, 103, 216, 211, 88, 208, 156, 246, 181, 214,
			201, 98, 168, 34, 150, 196, 43, 145, 15, 143, 74, 122, 22, 110,
			145, 147, 109, 42, 100, 147, 138, 81, 40, 221, 192, 250, 156, 44,
			10, 68, 185, 56, 243, 137, 102, 82, 178, 62, 35, 89, 22, 32,
			221, 19, 205, 44, 11, 10, 223, 145, 79, 14, 188, 152, 121, 145,
			180, 108, 98, 6, 180, 123, 193, 0, 179, 248, 233, 250, 69, 123,
			38, 182, 157, 244, 176, 183, 105, 215, 137, 100, 60, 105, 170, 142,
			249, 91, 100, 41, 173, 176, 206, 16, 243, 37, 157, 36, 115, 169,
			79, 235, 60, 89, 192, 245, 78, 230, 210, 133, 205, 236, 29, 163,
			112, 131, 16, 237, 99, 247, 60, 22, 255, 208, 145, 133, 93, 114,
			126, 107, 212, 107, 199, 158, 255, 146, 69, 61, 133, 16, 121, 68,
			35, 249, 94, 65, 47, 146, 19, 126, 218, 41, 161, 52, 171, 40,
			220, 33, 159, 237, 197, 84, 140, 58, 3, 38, 155, 163, 232, 135,
			43, 236, 250, 11, 114, 234, 128, 198, 1, 243, 101, 75, 122, 114,
			36, 172, 75, 36, 127, 224, 52, 183, 221, 90, 251, 121, 171, 93,
			109, 239, 183, 158, 239, 215, 241, 240, 117, 199, 117, 182, 207, 100,
			172, 207, 8, 217, 175, 59, 191, 217, 115, 106, 109, 103, 251, 12,
			177, 206, 146, 83, 105, 255, 157, 221, 234, 175, 158, 158, 185, 100,
			157, 36, 75, 211, 14, 235, 91, 229, 103, 215, 63, 102, 161, 247,
			146, 138, 97, 231, 209, 127, 252, 130, 44, 90, 185, 92, 134, 26,
			228, 79, 13, 188, 159, 202, 101, 172, 245, 127, 108, 28, 185, 106,
			90, 95, 67, 88, 84, 235, 199, 124, 192, 70, 3, 168, 142, 100,
			159, 199, 194, 126, 207, 157, 211, 190, 64, 95, 154, 156, 236, 207,
			110, 104, 152, 128, 30, 63, 164, 113, 148, 224, 10, 216, 106, 109,
			175, 8, 57, 9, 41, 132, 204, 167, 120, 13, 138, 123, 27, 3,
			160, 130, 175, 163, 40, 72, 207, 209, 118, 221, 154, 83, 111, 57,
			208, 101, 33, 157, 158, 126, 46, 102, 206, 145, 19, 36, 107, 102,
			44, 115, 41, 83, 74, 142, 34, 73, 166, 154, 30, 111, 170, 207,
			43, 120, 18, 153, 59, 149, 57, 103, 228, 47, 64, 53, 57, 107,
			82, 252, 77, 29, 252, 220, 181, 229, 169, 165, 179, 228, 231, 137,
			55, 55, 79, 103, 75, 249, 10, 138, 206, 195, 128, 10, 57, 151,
			36, 72, 174, 157, 73, 64, 83, 6, 145, 174, 77, 200, 73, 237,
			78, 23, 213, 240, 47, 210, 146, 97, 153, 167, 47, 94, 73, 75,
			166, 101, 158, 190, 86, 36, 110, 226, 104, 77, 43, 123, 45, 255,
			115, 112, 19, 122, 60, 10, 39, 243, 209, 7, 117, 162, 64, 170,
			62, 217, 10, 39, 200, 77, 172, 234, 117, 92, 154, 78, 106, 44,
			42, 90, 233, 164, 134, 162, 124, 177, 144, 150, 76, 203, 180, 150,
			175, 146, 63, 55, 72, 118, 33, 99, 229, 46, 100, 174, 26, 249,
			127, 110, 128, 54, 67, 237, 204, 19, 203, 180, 9, 184, 152, 53,
			4, 84, 210, 120, 192, 210, 245, 10, 67, 141, 18, 40, 222, 113,
			41, 79, 33, 244, 58, 83, 56, 212, 35, 53, 172, 167, 175, 184,
			142, 162, 211, 123, 60, 214, 139, 120, 76, 3, 13, 200, 187, 30,
			11, 71, 177, 190, 31, 143, 41, 162, 76, 204, 129, 146, 250, 50,
			208, 67, 26, 1, 235, 2, 67, 38, 82, 106, 52, 40, 233, 117,
			90, 80, 218, 188, 176, 96, 145, 231, 36, 183, 128, 235, 244, 69,
			246, 171, 124, 19, 170, 41, 23, 58, 152, 68, 92, 234, 80, 162,
			237, 16, 197, 180, 9, 180, 85, 137, 9, 173, 101, 188, 174, 66,
			132, 221, 101, 161, 164, 152, 131, 37, 68, 18, 173, 46, 232, 197,
			251, 34, 123, 49, 45, 101, 45, 243, 139, 203, 64, 238, 226, 228,
			134, 101, 94, 202, 90, 249, 178, 222, 9, 199, 234, 4, 151, 110,
			20, 209, 87, 67, 234, 75, 181, 63, 18, 66, 6, 142, 61, 153,
			150, 178, 150, 121, 233, 244, 89, 242, 55, 12, 164, 155, 181, 204,
			66, 246, 39, 121, 129, 198, 151, 18, 234, 123, 66, 67, 247, 148,
			150, 190, 156, 157, 146, 78, 25, 80, 82, 114, 21, 5, 3, 214,
			69, 188, 42, 25, 106, 25, 67, 110, 53, 242, 194, 201, 247, 52,
			80, 238, 62, 113, 204, 218, 4, 108, 116, 39, 83, 246, 148, 104,
			133, 236, 233, 180, 164, 24, 178, 206, 147, 219, 200, 157, 105, 153,
			203, 217, 51, 249, 235, 31, 147, 250, 29, 153, 77, 67, 141, 156,
			150, 178, 150, 185, 124, 234, 52, 41, 146, 108, 206, 176, 114, 165,
			204, 13, 35, 127, 17, 220, 64, 49, 44, 39, 218, 36, 231, 140,
			45, 217, 165, 74, 111, 165, 165, 243, 228, 9, 201, 229, 12, 181,
			250, 229, 236, 249, 252, 35, 84, 212, 17, 203, 212, 14, 216, 38,
			144, 228, 235, 225, 68, 103, 226, 184, 240, 135, 94, 200, 18, 40,
			130, 233, 177, 30, 20, 116, 10, 201, 94, 50, 20, 90, 50, 203,
			217, 165, 180, 100, 88, 102, 249, 196, 233, 180, 100, 90, 102, 217,
			58, 71, 254, 126, 22, 121, 48, 44, 115, 35, 123, 38, 255, 71,
			89, 112, 183, 241, 188, 252, 173, 93, 146, 122, 136, 227, 217, 83,
			9, 213, 145, 22, 22, 129, 142, 195, 219, 91, 229, 228, 34, 56,
			73, 227, 55, 9, 20, 88, 116, 200, 245, 197, 186, 168, 188, 118,
			235, 7, 141, 90, 181, 237, 54, 234, 207, 221, 237, 55, 21, 69,
			70, 84, 94, 239, 55, 119, 159, 59, 173, 90, 117, 207, 217, 126,
			222, 118, 90, 109, 108, 75, 168, 87, 94, 55, 157, 214, 254, 46,
			214, 21, 8, 60, 193, 244, 254, 8, 153, 50, 28, 51, 30, 45,
			109, 58, 18, 23, 55, 193, 113, 248, 82, 70, 37, 41, 115, 108,
			79, 149, 104, 44, 40, 213, 164, 74, 84, 43, 183, 113, 226, 211,
			180, 100, 90, 230, 198, 103, 167, 201, 95, 26, 36, 155, 203, 90,
			185, 205, 204, 47, 140, 252, 95, 24, 144, 24, 229, 209, 75, 162,
			177, 135, 246, 16, 143, 162, 72, 95, 61, 160, 198, 124, 79, 208,
			244, 10, 65, 120, 3, 58, 171, 77, 147, 40, 250, 138, 250, 35,
			101, 251, 44, 154, 237, 6, 69, 77, 148, 113, 165, 210, 183, 58,
			60, 34, 115, 237, 141, 86, 25, 30, 236, 237, 167, 47, 27, 102,
			13, 10, 1, 176, 48, 61, 46, 16, 192, 99, 197, 146, 78, 155,
			66, 175, 151, 6, 18, 101, 17, 155, 75, 167, 201, 223, 86, 80,
			58, 171, 108, 244, 126, 246, 82, 254, 111, 26, 200, 168, 126, 90,
			132, 215, 246, 233, 150, 73, 240, 17, 56, 158, 223, 135, 151, 116,
			178, 162, 13, 115, 232, 177, 248, 136, 26, 136, 74, 237, 189, 129,
			242, 202, 16, 80, 225, 199, 172, 163, 180, 209, 231, 227, 153, 125,
			141, 61, 161, 120, 130, 34, 181, 123, 118, 42, 73, 25, 168, 244,
			237, 82, 178, 46, 89, 140, 78, 247, 179, 63, 73, 75, 134, 101,
			222, 255, 252, 167, 105, 201, 180, 204, 251, 23, 191, 36, 132, 100,
			115, 166, 149, 251, 101, 230, 129, 129, 66, 169, 189, 251, 203, 37,
			139, 252, 138, 228, 114, 166, 146, 169, 150, 61, 155, 255, 5, 52,
			105, 143, 190, 218, 132, 111, 191, 246, 86, 190, 255, 70, 253, 103,
			117, 229, 238, 243, 111, 174, 23, 43, 111, 85, 148, 174, 95, 33,
			240, 216, 123, 5, 33, 141, 122, 178, 191, 9, 183, 110, 36, 236,
			152, 184, 215, 106, 137, 153, 152, 200, 78, 237, 196, 201, 180, 100,
			90, 102, 237, 244, 25, 114, 25, 167, 53, 44, 115, 39, 123, 46,
			111, 29, 161, 180, 126, 243, 214, 148, 148, 178, 184, 157, 41, 41,
			101, 113, 59, 39, 62, 75, 75, 166, 101, 238, 156, 181, 200, 46,
			201, 230, 114, 86, 238, 81, 230, 137, 145, 255, 229, 91, 254, 166,
			51, 234, 129, 76, 80, 34, 76, 1, 31, 96, 198, 120, 164, 45,
			221, 191, 168, 155, 156, 97, 153, 143, 150, 46, 146, 127, 170, 22,
			60, 167, 148, 83, 207, 158, 207, 255, 61, 189, 224, 199, 12, 3,
			159, 199, 250, 233, 87, 48, 189, 168, 82, 225, 48, 53, 223, 178,
			138, 136, 12, 25, 235, 50, 181, 185, 58, 147, 15, 120, 144, 31,
			226, 224, 6, 60, 226, 177, 199, 194, 212, 193, 229, 80, 233, 245,
			68, 83, 57, 84, 122, 61, 113, 112, 57, 84, 122, 221, 58, 71,
			254, 87, 22, 229, 49, 44, 243, 32, 251, 7, 249, 255, 150, 125,
			87, 158, 153, 138, 254, 159, 138, 228, 234, 157, 113, 156, 234, 152,
			128, 84, 152, 228, 13, 13, 211, 135, 115, 51, 86, 188, 228, 46,
			117, 36, 104, 12, 99, 60, 5, 19, 148, 2, 147, 101, 192, 93,
			81, 112, 21, 64, 254, 133, 10, 129, 191, 216, 9, 189, 151, 44,
			162, 66, 20, 244, 107, 187, 121, 218, 200, 0, 153, 113, 48, 140,
			57, 158, 208, 232, 189, 85, 240, 19, 60, 92, 40, 225, 125, 41,
			151, 233, 153, 110, 25, 58, 35, 197, 133, 24, 13, 244, 121, 166,
			66, 179, 201, 147, 22, 26, 204, 221, 12, 43, 106, 215, 4, 60,
			209, 112, 28, 124, 30, 117, 89, 111, 164, 161, 211, 116, 161, 148,
			73, 31, 76, 23, 74, 153, 244, 193, 9, 43, 45, 153, 150, 121,
			240, 147, 207, 201, 175, 73, 54, 183, 96, 229, 158, 101, 168, 145,
			119, 222, 50, 233, 97, 154, 169, 104, 191, 224, 133, 130, 3, 190,
			105, 211, 176, 171, 80, 251, 53, 52, 71, 81, 65, 57, 179, 66,
			237, 0, 191, 19, 164, 149, 91, 48, 44, 243, 217, 210, 231, 228,
			31, 40, 187, 94, 80, 118, 253, 109, 246, 124, 254, 143, 181, 93,
			39, 235, 161, 47, 83, 61, 49, 125, 251, 51, 140, 185, 79, 133,
			72, 100, 156, 155, 251, 7, 154, 106, 56, 242, 217, 138, 127, 88,
			64, 7, 189, 187, 95, 115, 161, 198, 7, 138, 196, 1, 141, 149,
			2, 99, 2, 69, 93, 125, 144, 122, 180, 5, 180, 230, 111, 19,
			37, 45, 160, 53, 127, 155, 88, 243, 2, 90, 243, 183, 214, 57,
			242, 111, 180, 20, 134, 101, 6, 217, 51, 249, 127, 105, 28, 209,
			211, 113, 220, 186, 111, 87, 207, 76, 48, 97, 224, 72, 128, 78,
			115, 158, 84, 148, 77, 2, 0, 133, 215, 170, 235, 243, 189, 102,
			227, 145, 83, 107, 191, 169, 232, 98, 237, 0, 3, 176, 182, 71,
			236, 166, 115, 182, 59, 119, 239, 220, 185, 179, 118, 247, 198, 173,
			141, 59, 55, 111, 172, 172, 173, 116, 239, 222, 184, 189, 177, 222,
			165, 235, 171, 171, 55, 111, 117, 131, 181, 194, 84, 96, 101, 21,
			193, 84, 96, 101, 21, 65, 18, 90, 23, 208, 42, 130, 207, 78,
			79, 79, 45, 254, 213, 121, 114, 231, 125, 57, 33, 222, 237, 71,
			94, 88, 241, 130, 1, 139, 146, 20, 17, 191, 147, 3, 140, 207,
			211, 76, 62, 237, 105, 99, 107, 254, 67, 127, 19, 147, 255, 221,
			14, 73, 10, 127, 110, 144, 159, 58, 175, 134, 60, 150, 115, 176,
			84, 52, 245, 99, 89, 149, 209, 199, 212, 11, 211, 212, 90, 23,
			172, 159, 145, 83, 126, 200, 71, 193, 243, 100, 31, 37, 73, 246,
			73, 172, 220, 211, 117, 214, 5, 242, 73, 224, 73, 79, 80, 121,
			193, 196, 230, 180, 168, 136, 226, 83, 135, 11, 57, 77, 20, 11,
			214, 13, 66, 84, 52, 127, 142, 201, 220, 133, 69, 60, 63, 249,
			201, 252, 89, 198, 244, 120, 166, 121, 66, 166, 159, 235, 191, 37,
			11, 85, 165, 19, 203, 35, 214, 187, 98, 88, 107, 246, 241, 42,
			180, 223, 43, 114, 254, 243, 119, 206, 108, 240, 233, 109, 33, 179,
			117, 235, 217, 141, 223, 101, 41, 239, 225, 247, 176, 243, 232, 207,
			206, 232, 68, 255, 214, 95, 209, 68, 255, 242, 44, 209, 95, 198,
			79, 195, 50, 79, 100, 110, 39, 57, 255, 167, 153, 95, 165, 57,
			191, 250, 252, 175, 6, 201, 46, 102, 172, 220, 217, 204, 23, 70,
			254, 63, 27, 128, 171, 3, 124, 136, 71, 184, 83, 119, 59, 240,
			88, 36, 61, 22, 209, 88, 167, 131, 54, 129, 167, 201, 123, 112,
			63, 77, 118, 171, 123, 174, 72, 238, 20, 154, 123, 53, 112, 94,
			13, 67, 30, 211, 120, 147, 192, 245, 233, 219, 90, 191, 207, 135,
			98, 37, 89, 133, 149, 128, 30, 218, 222, 112, 40, 134, 92, 226,
			115, 151, 120, 232, 211, 100, 84, 37, 121, 186, 45, 42, 200, 71,
			64, 15, 223, 75, 230, 7, 146, 24, 198, 60, 64, 87, 189, 168,
			92, 222, 217, 165, 83, 228, 207, 76, 146, 91, 196, 172, 56, 159,
			61, 200, 255, 35, 19, 222, 53, 50, 144, 49, 235, 245, 148, 212,
			199, 181, 121, 226, 37, 62, 52, 162, 216, 134, 241, 153, 164, 16,
			85, 232, 184, 78, 231, 130, 25, 238, 155, 228, 210, 70, 239, 105,
			12, 247, 162, 12, 157, 239, 82, 26, 211, 155, 40, 8, 120, 68,
			193, 27, 73, 62, 240, 36, 83, 74, 158, 40, 3, 241, 99, 30,
			193, 111, 121, 39, 77, 207, 149, 166, 143, 164, 232, 42, 132, 122,
			254, 75, 101, 19, 161, 126, 29, 173, 15, 69, 194, 152, 122, 193,
			68, 89, 78, 186, 166, 173, 161, 23, 69, 52, 198, 115, 241, 45,
			214, 251, 245, 136, 198, 19, 27, 92, 9, 1, 167, 34, 186, 38,
			97, 204, 227, 151, 192, 186, 243, 175, 241, 1, 69, 198, 21, 81,
			164, 147, 55, 16, 9, 69, 50, 75, 226, 122, 84, 32, 108, 23,
			210, 139, 85, 86, 171, 226, 131, 24, 249, 253, 25, 157, 152, 161,
			228, 99, 138, 175, 167, 240, 198, 45, 80, 89, 58, 222, 177, 145,
			196, 12, 171, 123, 174, 126, 85, 37, 181, 51, 95, 212, 71, 12,
			249, 197, 11, 105, 41, 107, 153, 249, 159, 174, 167, 37, 211, 50,
			243, 247, 155, 136, 199, 51, 86, 238, 75, 181, 131, 211, 211, 170,
			47, 151, 190, 34, 219, 233, 105, 213, 229, 236, 185, 252, 109, 29,
			152, 154, 202, 67, 218, 160, 22, 118, 182, 116, 233, 45, 6, 186,
			207, 52, 63, 231, 241, 52, 63, 215, 87, 4, 230, 229, 36, 168,
			104, 174, 46, 39, 232, 89, 243, 113, 249, 172, 69, 14, 211, 83,
			171, 43, 217, 47, 242, 108, 170, 228, 228, 225, 216, 81, 195, 153,
			183, 27, 157, 168, 41, 172, 134, 29, 31, 239, 183, 218, 128, 88,
			164, 67, 241, 21, 244, 12, 24, 105, 6, 143, 3, 67, 120, 94,
			111, 94, 153, 114, 168, 194, 222, 149, 19, 159, 207, 29, 113, 93,
			249, 105, 158, 124, 138, 28, 102, 45, 115, 57, 73, 113, 50, 217,
			236, 130, 42, 165, 195, 20, 247, 203, 39, 206, 164, 37, 211, 50,
			151, 207, 157, 79, 134, 153, 150, 121, 53, 123, 46, 105, 50, 23,
			84, 41, 29, 166, 92, 206, 213, 169, 62, 76, 213, 243, 172, 69,
			254, 217, 2, 142, 203, 89, 230, 205, 236, 213, 252, 159, 228, 240,
			202, 107, 238, 148, 81, 165, 124, 104, 177, 92, 39, 164, 83, 149,
			131, 147, 156, 216, 99, 114, 189, 139, 7, 241, 115, 91, 165, 59,
			10, 67, 232, 243, 145, 114, 191, 174, 77, 109, 69, 105, 50, 215,
			190, 186, 185, 186, 90, 134, 181, 205, 213, 85, 176, 109, 155, 64,
			67, 217, 216, 152, 161, 111, 165, 19, 24, 171, 173, 210, 161, 32,
			227, 81, 164, 175, 114, 147, 173, 59, 71, 151, 16, 168, 115, 153,
			56, 99, 170, 146, 207, 152, 143, 241, 152, 201, 3, 65, 85, 174,
			41, 147, 75, 193, 244, 9, 29, 222, 182, 11, 246, 61, 250, 120,
			124, 0, 206, 195, 48, 185, 174, 86, 252, 39, 235, 221, 249, 46,
			145, 51, 182, 161, 138, 71, 67, 117, 126, 136, 241, 165, 60, 155,
			71, 13, 247, 88, 36, 96, 13, 217, 81, 59, 83, 246, 85, 95,
			165, 174, 25, 250, 154, 205, 15, 98, 232, 69, 250, 105, 99, 122,
			252, 169, 135, 106, 20, 166, 188, 6, 74, 45, 250, 94, 28, 36,
			182, 174, 198, 17, 96, 145, 218, 138, 211, 199, 246, 98, 224, 133,
			97, 114, 233, 173, 67, 189, 126, 103, 144, 76, 144, 240, 163, 86,
			69, 248, 125, 26, 140, 66, 74, 222, 239, 42, 241, 24, 65, 13,
			78, 22, 59, 37, 206, 35, 42, 108, 178, 254, 71, 198, 156, 142,
			19, 52, 169, 47, 215, 161, 203, 104, 24, 160, 163, 75, 254, 52,
			98, 182, 67, 209, 159, 216, 176, 69, 125, 15, 255, 46, 169, 207,
			4, 153, 9, 168, 171, 142, 144, 138, 249, 224, 184, 125, 3, 244,
			85, 114, 227, 165, 194, 91, 98, 185, 185, 69, 101, 171, 233, 174,
			81, 185, 235, 205, 63, 248, 42, 45, 153, 150, 121, 243, 202, 114,
			10, 29, 255, 79, 0, 0, 0, 255, 255, 108, 175, 136, 55, 253,
			60, 0, 0},
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
	ret, err := discovery.GetDescriptorSet("weetbix.internal.admin.Admin")
	if err != nil {
		panic(err)
	}
	return ret
}
