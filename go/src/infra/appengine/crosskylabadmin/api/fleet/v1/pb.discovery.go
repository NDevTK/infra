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
			8, 0, 0, 0, 0, 0, 0, 255, 220, 90, 205, 111, 28, 71,
			118, 239, 234, 234, 25, 54, 139, 146, 56, 44, 82, 20, 53, 162,
			204, 210, 144, 146, 69, 137, 28, 218, 180, 44, 139, 180, 97, 91,
			36, 245, 65, 145, 250, 34, 165, 245, 218, 11, 175, 213, 156, 174,
			153, 105, 179, 167, 123, 182, 187, 154, 20, 173, 24, 155, 56, 54,
			144, 83, 130, 228, 146, 83, 144, 107, 118, 129, 5, 2, 4, 11,
			4, 8, 16, 228, 15, 72, 78, 201, 61, 200, 33, 167, 228, 180,
			167, 36, 135, 32, 120, 245, 209, 51, 164, 56, 35, 217, 201, 37,
			203, 139, 230, 85, 213, 123, 245, 234, 189, 223, 123, 93, 245, 158,
			200, 111, 108, 178, 20, 68, 245, 196, 91, 240, 218, 109, 30, 53,
			130, 136, 47, 212, 146, 56, 77, 119, 15, 66, 111, 199, 243, 91,
			65, 180, 224, 181, 131, 133, 122, 200, 185, 88, 216, 123, 123, 161,
			22, 183, 90, 113, 84, 109, 39, 177, 136, 233, 233, 35, 75, 171,
			114, 217, 202, 181, 207, 22, 191, 143, 204, 247, 229, 143, 123, 255,
			140, 72, 145, 58, 142, 229, 174, 147, 95, 35, 130, 78, 80, 236,
			88, 116, 241, 151, 136, 173, 198, 237, 131, 36, 104, 52, 5, 91,
			124, 235, 237, 27, 236, 73, 147, 179, 205, 167, 171, 235, 236, 102,
			38, 154, 113, 146, 86, 217, 205, 48, 100, 114, 65, 202, 18, 158,
			242, 100, 143, 251, 85, 194, 158, 166, 156, 197, 117, 38, 154, 65,
			202, 210, 56, 75, 106, 156, 213, 98, 159, 179, 32, 101, 141, 120,
			143, 39, 17, 247, 89, 22, 249, 60, 97, 162, 201, 217, 205, 182,
			87, 3, 193, 65, 141, 71, 41, 159, 99, 63, 226, 73, 26, 196,
			17, 91, 172, 190, 69, 152, 104, 122, 130, 213, 188, 136, 237, 112,
			86, 143, 179, 200, 103, 65, 36, 185, 54, 215, 87, 111, 61, 216,
			190, 197, 234, 65, 200, 171, 132, 184, 4, 217, 20, 23, 173, 55,
			224, 151, 75, 177, 107, 173, 147, 65, 98, 187, 67, 242, 231, 78,
			81, 90, 237, 29, 242, 139, 37, 242, 193, 247, 50, 122, 16, 237,
			241, 72, 196, 201, 65, 95, 187, 87, 126, 129, 200, 153, 59, 92,
			108, 11, 111, 39, 228, 250, 0, 91, 252, 103, 25, 79, 5, 101,
			100, 104, 39, 11, 66, 255, 137, 151, 52, 184, 152, 64, 12, 93,
			30, 220, 234, 30, 162, 99, 164, 208, 138, 125, 30, 78, 216, 114,
			78, 17, 180, 76, 220, 102, 156, 138, 200, 107, 241, 9, 44, 39,
			114, 154, 46, 147, 137, 212, 19, 161, 183, 179, 30, 213, 227, 164,
			229, 137, 32, 142, 188, 240, 113, 198, 147, 131, 9, 135, 161, 203,
			238, 86, 207, 249, 202, 223, 35, 50, 241, 178, 174, 105, 59, 142,
			82, 14, 202, 194, 9, 245, 176, 81, 182, 107, 136, 94, 38, 195,
			245, 32, 105, 237, 123, 137, 97, 214, 106, 31, 29, 6, 89, 117,
			175, 46, 204, 42, 117, 134, 238, 33, 122, 133, 148, 0, 52, 241,
			106, 215, 150, 142, 92, 246, 210, 56, 29, 39, 197, 132, 123, 105,
			28, 77, 20, 228, 10, 77, 85, 56, 57, 183, 45, 143, 186, 18,
			123, 137, 127, 51, 242, 239, 131, 245, 182, 69, 226, 9, 222, 56,
			0, 43, 106, 75, 172, 233, 211, 228, 52, 216, 125, 7, 152, 140,
			221, 37, 209, 241, 6, 238, 242, 70, 229, 17, 25, 87, 219, 220,
			213, 62, 120, 173, 29, 186, 125, 104, 31, 246, 97, 229, 191, 109,
			114, 126, 155, 11, 37, 245, 88, 228, 236, 145, 115, 105, 239, 163,
			201, 205, 134, 22, 23, 171, 199, 66, 178, 218, 199, 40, 119, 173,
			173, 126, 130, 105, 131, 140, 167, 199, 158, 85, 158, 97, 104, 113,
			190, 239, 150, 71, 153, 238, 90, 91, 61, 196, 29, 69, 27, 126,
			45, 180, 57, 199, 163, 109, 134, 156, 52, 67, 235, 45, 175, 193,
			53, 76, 14, 15, 174, 16, 226, 166, 122, 247, 10, 35, 111, 244,
			178, 191, 138, 134, 202, 46, 185, 208, 101, 70, 105, 164, 53, 30,
			114, 136, 166, 213, 36, 16, 60, 129, 141, 255, 175, 16, 246, 41,
			153, 58, 108, 192, 239, 183, 85, 63, 168, 253, 173, 77, 152, 148,
			198, 251, 160, 237, 15, 16, 185, 144, 190, 234, 180, 26, 116, 55,
			94, 13, 186, 227, 249, 239, 90, 91, 175, 222, 132, 126, 131, 200,
			84, 218, 223, 22, 26, 138, 215, 95, 11, 138, 199, 169, 241, 170,
			13, 14, 65, 229, 83, 114, 161, 143, 253, 116, 238, 156, 36, 131,
			126, 224, 171, 117, 210, 78, 238, 86, 103, 160, 43, 127, 217, 135,
			242, 215, 69, 50, 189, 150, 181, 218, 135, 68, 62, 137, 215, 60,
			225, 165, 34, 78, 184, 246, 78, 229, 18, 153, 233, 191, 76, 41,
			177, 248, 87, 14, 25, 92, 55, 31, 45, 154, 146, 210, 209, 84,
			79, 171, 61, 76, 214, 227, 251, 85, 94, 120, 237, 245, 218, 14,
			223, 33, 50, 126, 124, 96, 209, 107, 189, 220, 213, 47, 15, 150,
			223, 253, 158, 92, 90, 143, 63, 68, 228, 108, 79, 175, 209, 247,
			122, 8, 125, 85, 156, 148, 123, 97, 255, 213, 0, 249, 83, 68,
			38, 251, 57, 145, 46, 247, 18, 253, 106, 128, 148, 223, 255, 65,
			188, 74, 179, 31, 120, 111, 252, 199, 75, 100, 128, 22, 28, 235,
			59, 180, 240, 91, 122, 113, 36, 239, 19, 187, 104, 81, 231, 132,
			117, 30, 149, 23, 88, 30, 88, 12, 84, 14, 106, 156, 53, 121,
			216, 78, 89, 203, 139, 188, 6, 103, 171, 205, 36, 110, 241, 135,
			219, 108, 237, 233, 147, 180, 74, 8, 33, 184, 104, 33, 138, 79,
			184, 35, 228, 175, 17, 113, 138, 150, 109, 81, 92, 178, 183, 203,
			191, 68, 236, 104, 248, 176, 132, 123, 126, 42, 85, 76, 229, 56,
			219, 211, 19, 245, 56, 97, 30, 107, 123, 137, 8, 106, 89, 232,
			37, 32, 158, 213, 147, 184, 197, 124, 227, 72, 66, 192, 198, 137,
			180, 84, 16, 9, 30, 169, 11, 95, 120, 192, 162, 152, 109, 61,
			90, 149, 50, 192, 130, 251, 73, 32, 130, 168, 113, 220, 54, 251,
			129, 104, 50, 47, 217, 9, 68, 226, 37, 7, 172, 22, 75, 57,
			85, 66, 78, 144, 2, 104, 142, 40, 46, 21, 199, 13, 101, 83,
			92, 58, 83, 53, 20, 166, 184, 180, 244, 152, 252, 137, 58, 36,
			162, 120, 204, 230, 229, 111, 17, 59, 62, 66, 89, 202, 197, 177,
			39, 221, 243, 194, 140, 179, 203, 160, 236, 126, 51, 168, 53, 217,
			106, 242, 112, 123, 91, 130, 240, 38, 128, 16, 142, 39, 217, 20,
			54, 0, 42, 73, 38, 154, 179, 115, 68, 30, 80, 101, 116, 230,
			115, 112, 77, 202, 226, 40, 60, 200, 213, 71, 160, 84, 113, 210,
			80, 54, 197, 99, 231, 151, 13, 133, 41, 30, 187, 85, 35, 127,
			166, 212, 183, 41, 158, 176, 163, 242, 31, 161, 222, 223, 76, 230,
			203, 25, 173, 141, 218, 246, 216, 179, 72, 55, 29, 61, 133, 58,
			219, 241, 103, 121, 245, 73, 192, 188, 19, 197, 41, 67, 129, 182,
			236, 99, 67, 97, 138, 39, 54, 66, 242, 66, 30, 4, 83, 60,
			105, 239, 151, 35, 214, 47, 7, 244, 193, 29, 232, 46, 76, 232,
			214, 226, 168, 30, 52, 100, 216, 16, 230, 69, 62, 3, 40, 241,
			148, 5, 130, 137, 184, 131, 196, 92, 77, 140, 40, 158, 44, 86,
			12, 101, 83, 60, 57, 125, 219, 80, 160, 217, 227, 140, 156, 37,
			182, 99, 81, 135, 89, 239, 160, 242, 73, 22, 241, 231, 130, 9,
			175, 177, 204, 222, 149, 177, 227, 0, 228, 152, 59, 69, 4, 113,
			28, 25, 58, 211, 206, 185, 114, 131, 61, 121, 184, 246, 240, 114,
			35, 225, 141, 56, 57, 136, 130, 116, 135, 139, 217, 101, 150, 240,
			86, 188, 199, 89, 154, 181, 219, 113, 34, 164, 17, 195, 56, 222,
			5, 172, 103, 109, 114, 244, 104, 65, 231, 89, 196, 118, 14, 88,
			215, 99, 140, 93, 101, 242, 58, 38, 143, 1, 187, 22, 96, 219,
			156, 66, 20, 79, 15, 141, 27, 10, 83, 60, 125, 182, 76, 134,
			164, 126, 136, 226, 25, 231, 180, 158, 66, 5, 160, 12, 27, 128,
			111, 102, 168, 100, 40, 76, 241, 204, 232, 24, 89, 151, 108, 54,
			197, 151, 156, 137, 242, 7, 236, 193, 195, 39, 183, 150, 217, 11,
			115, 105, 251, 90, 26, 249, 69, 151, 106, 115, 74, 179, 175, 153,
			151, 112, 230, 133, 130, 39, 145, 39, 130, 61, 158, 154, 109, 236,
			2, 200, 202, 41, 68, 241, 165, 161, 81, 67, 97, 138, 47, 141,
			159, 33, 127, 227, 200, 93, 49, 197, 139, 206, 165, 242, 175, 28,
			194, 182, 123, 60, 20, 89, 224, 67, 30, 169, 7, 60, 101, 251,
			77, 46, 154, 50, 237, 6, 144, 176, 229, 23, 7, 0, 236, 29,
			50, 229, 60, 224, 52, 159, 174, 197, 173, 32, 106, 144, 14, 140,
			52, 168, 87, 55, 215, 171, 132, 176, 109, 229, 41, 112, 208, 179,
			234, 130, 158, 107, 112, 49, 175, 92, 53, 111, 92, 53, 47, 15,
			205, 170, 213, 42, 155, 151, 215, 103, 248, 249, 108, 174, 19, 68,
			17, 175, 241, 52, 133, 148, 5, 78, 207, 82, 158, 164, 128, 200,
			29, 128, 41, 248, 92, 196, 172, 25, 239, 31, 3, 240, 84, 26,
			82, 225, 58, 75, 184, 47, 249, 69, 147, 7, 137, 9, 188, 57,
			121, 152, 32, 225, 41, 203, 180, 208, 92, 166, 182, 206, 1, 49,
			7, 78, 37, 191, 84, 241, 170, 210, 89, 126, 131, 96, 15, 153,
			146, 125, 238, 27, 17, 105, 26, 215, 2, 79, 112, 95, 37, 221,
			195, 209, 14, 198, 89, 87, 223, 170, 8, 44, 27, 178, 122, 24,
			239, 155, 3, 3, 168, 83, 230, 9, 102, 112, 162, 178, 195, 28,
			219, 231, 204, 15, 82, 176, 103, 22, 164, 77, 230, 25, 169, 113,
			194, 162, 56, 154, 215, 148, 113, 78, 150, 74, 223, 60, 107, 242,
			44, 1, 174, 90, 90, 221, 4, 201, 155, 193, 174, 206, 119, 107,
			82, 155, 203, 102, 155, 217, 103, 240, 65, 83, 88, 194, 5, 64,
			143, 107, 40, 68, 241, 226, 224, 5, 67, 1, 178, 102, 46, 202,
			208, 70, 212, 185, 110, 173, 30, 14, 237, 235, 42, 180, 33, 34,
			174, 187, 76, 134, 14, 130, 208, 126, 207, 57, 39, 37, 32, 25,
			113, 239, 105, 20, 35, 25, 113, 239, 233, 136, 67, 50, 226, 222,
			211, 17, 135, 0, 224, 55, 156, 41, 61, 5, 17, 119, 35, 103,
			3, 249, 55, 134, 202, 134, 194, 20, 223, 56, 255, 134, 102, 179,
			41, 94, 202, 119, 131, 152, 89, 202, 217, 64, 228, 82, 190, 27,
			68, 201, 82, 190, 27, 166, 120, 217, 97, 122, 10, 76, 176, 156,
			179, 129, 9, 150, 135, 140, 72, 48, 193, 242, 27, 83, 228, 191,
			144, 228, 115, 40, 94, 113, 198, 203, 255, 142, 216, 150, 188, 241,
			235, 184, 225, 207, 219, 161, 23, 169, 20, 20, 215, 217, 149, 253,
			230, 193, 21, 21, 95, 93, 95, 250, 195, 128, 37, 108, 223, 75,
			89, 59, 168, 237, 114, 191, 202, 30, 197, 105, 26, 200, 89, 248,
			210, 72, 48, 47, 19, 194, 222, 158, 101, 21, 112, 218, 124, 218,
			230, 181, 160, 30, 212, 42, 132, 45, 206, 178, 138, 138, 29, 200,
			41, 234, 185, 41, 175, 12, 65, 170, 191, 83, 230, 222, 208, 1,
			105, 208, 157, 16, 230, 36, 35, 15, 100, 18, 8, 227, 70, 131,
			251, 0, 172, 54, 220, 221, 34, 161, 120, 224, 76, 145, 47, 3,
			176, 106, 12, 227, 20, 224, 240, 57, 133, 40, 94, 25, 26, 49,
			20, 166, 120, 101, 236, 52, 249, 23, 68, 108, 199, 166, 206, 93,
			107, 19, 149, 255, 9, 177, 62, 133, 11, 105, 58, 72, 44, 59,
			65, 199, 114, 187, 252, 32, 85, 161, 230, 243, 122, 16, 113, 128,
			254, 225, 108, 207, 35, 161, 179, 131, 146, 77, 242, 80, 83, 38,
			216, 13, 34, 31, 36, 169, 117, 112, 247, 172, 197, 109, 125, 40,
			118, 196, 110, 85, 121, 211, 50, 193, 181, 190, 198, 58, 9, 145,
			251, 108, 39, 19, 172, 150, 37, 9, 143, 68, 120, 192, 130, 70,
			20, 39, 112, 177, 149, 144, 7, 108, 221, 117, 167, 37, 154, 108,
			128, 252, 186, 173, 128, 102, 75, 200, 175, 219, 174, 161, 16, 197,
			235, 131, 35, 134, 194, 20, 175, 143, 157, 214, 108, 136, 226, 123,
			246, 168, 158, 2, 200, 223, 203, 217, 0, 242, 247, 6, 79, 25,
			10, 83, 124, 111, 132, 106, 54, 155, 226, 141, 156, 13, 32, 191,
			145, 179, 129, 200, 141, 156, 13, 80, 190, 49, 66, 201, 63, 128,
			83, 48, 117, 30, 89, 219, 168, 252, 119, 198, 41, 71, 11, 56,
			230, 14, 179, 203, 15, 192, 239, 190, 190, 170, 110, 31, 123, 33,
			82, 214, 149, 142, 58, 98, 98, 131, 84, 153, 211, 170, 132, 125,
			194, 243, 60, 37, 14, 219, 58, 98, 158, 239, 7, 210, 243, 34,
			150, 147, 38, 63, 201, 173, 235, 153, 200, 18, 62, 223, 78, 226,
			184, 14, 31, 22, 245, 196, 214, 87, 113, 7, 162, 244, 145, 171,
			178, 0, 6, 7, 60, 214, 14, 192, 210, 1, 143, 181, 73, 176,
			116, 192, 99, 237, 0, 44, 29, 240, 88, 59, 0, 131, 181, 182,
			114, 54, 112, 192, 86, 206, 6, 14, 216, 202, 217, 192, 1, 91,
			99, 167, 9, 33, 182, 227, 80, 231, 169, 245, 83, 36, 181, 128,
			32, 120, 234, 94, 36, 39, 136, 227, 56, 174, 69, 157, 31, 217,
			159, 98, 201, 227, 184, 176, 241, 143, 220, 146, 220, 202, 1, 13,
			63, 113, 62, 86, 83, 182, 85, 4, 106, 202, 80, 136, 226, 79,
			216, 251, 134, 194, 20, 127, 242, 225, 71, 154, 13, 81, 252, 99,
			231, 186, 158, 66, 69, 160, 202, 134, 130, 185, 115, 111, 27, 10,
			83, 252, 227, 107, 239, 106, 54, 155, 226, 207, 236, 179, 122, 10,
			32, 242, 153, 62, 152, 66, 221, 103, 131, 99, 134, 194, 20, 127,
			118, 102, 66, 179, 97, 138, 127, 98, 159, 215, 83, 144, 21, 127,
			146, 179, 129, 189, 127, 50, 56, 97, 40, 88, 121, 110, 146, 252,
			84, 178, 57, 20, 127, 110, 159, 43, 63, 86, 1, 104, 174, 194,
			224, 71, 209, 4, 207, 121, 106, 164, 210, 85, 45, 174, 204, 201,
			240, 10, 224, 227, 165, 111, 222, 17, 223, 135, 75, 9, 79, 90,
			65, 20, 135, 113, 67, 95, 146, 65, 126, 1, 54, 48, 154, 128,
			205, 63, 31, 28, 55, 20, 166, 248, 243, 179, 101, 233, 153, 2,
			117, 158, 89, 158, 242, 76, 1, 81, 252, 204, 189, 68, 254, 19,
			176, 95, 164, 78, 211, 10, 17, 164, 237, 87, 22, 181, 36, 158,
			131, 86, 0, 25, 91, 196, 253, 18, 216, 28, 129, 51, 16, 6,
			127, 222, 172, 124, 5, 195, 57, 234, 1, 15, 253, 148, 181, 178,
			84, 64, 234, 109, 39, 241, 94, 224, 119, 46, 36, 157, 104, 144,
			153, 121, 159, 135, 225, 60, 36, 103, 249, 102, 134, 191, 157, 89,
			136, 153, 70, 230, 37, 94, 36, 56, 87, 65, 38, 228, 131, 112,
			63, 8, 67, 96, 226, 207, 189, 26, 228, 164, 56, 226, 108, 7,
			100, 181, 60, 81, 107, 194, 69, 194, 3, 37, 130, 22, 87, 97,
			66, 176, 83, 68, 20, 55, 221, 89, 233, 224, 34, 160, 48, 208,
			128, 47, 202, 56, 9, 180, 89, 139, 18, 133, 129, 6, 124, 81,
			162, 48, 208, 113, 82, 4, 200, 124, 169, 51, 78, 81, 198, 201,
			151, 57, 27, 160, 240, 75, 157, 113, 138, 18, 133, 95, 234, 68,
			85, 4, 20, 238, 230, 108, 128, 194, 221, 156, 13, 68, 238, 230,
			108, 0, 188, 221, 17, 74, 126, 3, 206, 26, 160, 78, 106, 237,
			161, 242, 191, 30, 77, 84, 175, 231, 170, 163, 105, 237, 255, 133,
			151, 6, 16, 197, 169, 251, 166, 180, 219, 0, 120, 73, 104, 47,
			13, 72, 47, 9, 109, 183, 1, 233, 37, 161, 189, 52, 32, 189,
			36, 180, 151, 6, 192, 164, 89, 206, 6, 94, 202, 114, 54, 240,
			82, 150, 179, 129, 151, 50, 157, 205, 92, 234, 60, 183, 190, 86,
			49, 227, 34, 138, 159, 187, 151, 101, 54, 115, 33, 155, 29, 216,
			191, 163, 178, 153, 43, 179, 217, 129, 206, 102, 46, 104, 248, 149,
			179, 169, 166, 100, 54, 251, 202, 185, 104, 40, 68, 241, 87, 151,
			238, 24, 10, 83, 252, 213, 189, 13, 205, 134, 40, 126, 225, 220,
			213, 83, 144, 205, 94, 56, 211, 134, 130, 185, 153, 85, 67, 97,
			138, 95, 220, 190, 67, 134, 136, 237, 12, 210, 194, 207, 173, 111,
			144, 82, 113, 16, 81, 252, 115, 119, 86, 170, 56, 104, 91, 212,
			249, 93, 100, 159, 38, 39, 73, 1, 168, 130, 36, 139, 134, 68,
			64, 14, 148, 12, 137, 129, 28, 29, 211, 172, 136, 58, 191, 135,
			236, 49, 61, 137, 10, 146, 116, 13, 41, 103, 7, 135, 13, 137,
			129, 164, 163, 210, 96, 132, 58, 223, 34, 171, 74, 134, 8, 118,
			8, 2, 194, 189, 42, 39, 134, 168, 243, 29, 178, 22, 228, 196,
			16, 2, 194, 157, 203, 123, 149, 127, 57, 69, 150, 191, 87, 175,
			82, 36, 94, 109, 151, 39, 253, 59, 149, 191, 182, 201, 228, 163,
			44, 109, 174, 196, 34, 189, 29, 39, 178, 32, 241, 196, 75, 119,
			83, 211, 6, 88, 39, 37, 33, 159, 155, 95, 248, 153, 248, 34,
			21, 158, 46, 102, 159, 90, 156, 234, 89, 97, 20, 219, 176, 108,
			235, 148, 98, 52, 52, 221, 39, 163, 105, 173, 201, 253, 44, 12,
			162, 198, 23, 105, 119, 19, 233, 212, 226, 237, 30, 210, 250, 41,
			87, 221, 206, 197, 153, 88, 221, 162, 233, 75, 99, 149, 199, 132,
			190, 188, 146, 158, 39, 103, 215, 120, 221, 203, 66, 241, 242, 100,
			201, 162, 147, 100, 98, 211, 219, 129, 3, 203, 154, 237, 78, 28,
			139, 124, 22, 85, 166, 200, 249, 30, 154, 233, 94, 209, 69, 50,
			13, 11, 182, 120, 219, 11, 146, 123, 241, 14, 44, 235, 200, 75,
			187, 234, 248, 253, 151, 105, 113, 163, 100, 100, 139, 195, 3, 25,
			118, 52, 204, 99, 132, 118, 15, 234, 165, 159, 146, 11, 71, 85,
			187, 153, 249, 129, 56, 228, 214, 107, 196, 17, 94, 186, 171, 93,
			201, 122, 24, 63, 231, 219, 146, 171, 43, 51, 164, 210, 79, 180,
			82, 224, 202, 175, 16, 113, 115, 167, 143, 146, 97, 243, 123, 61,
			218, 243, 194, 192, 47, 89, 116, 144, 20, 182, 184, 231, 31, 148,
			16, 45, 145, 19, 15, 56, 247, 211, 213, 144, 123, 81, 214, 46,
			217, 116, 152, 12, 201, 17, 101, 147, 18, 166, 167, 8, 209, 3,
			41, 23, 37, 7, 88, 212, 220, 109, 47, 8, 185, 95, 42, 208,
			49, 82, 50, 44, 161, 87, 227, 45, 30, 137, 82, 145, 158, 38,
			35, 114, 244, 190, 23, 101, 94, 168, 197, 13, 228, 242, 215, 120,
			59, 140, 15, 74, 238, 149, 77, 50, 152, 159, 2, 102, 225, 223,
			142, 174, 195, 100, 104, 155, 39, 123, 241, 211, 237, 149, 13, 14,
			26, 159, 34, 100, 237, 233, 147, 109, 17, 39, 94, 131, 151, 108,
			122, 146, 12, 110, 61, 186, 191, 42, 235, 6, 37, 188, 248, 231,
			14, 25, 120, 162, 66, 143, 126, 131, 200, 233, 99, 97, 66, 223,
			249, 1, 112, 47, 247, 106, 151, 244, 69, 162, 108, 51, 244, 195,
			88, 207, 54, 195, 107, 224, 183, 103, 155, 225, 117, 64, 77, 61,
			66, 58, 248, 165, 151, 123, 136, 122, 9, 247, 229, 217, 215, 88,
			169, 183, 248, 99, 68, 202, 189, 33, 75, 123, 53, 111, 94, 25,
			64, 229, 165, 31, 192, 249, 191, 234, 174, 252, 197, 168, 250, 95,
			57, 15, 209, 111, 107, 115, 37, 48, 205, 149, 73, 84, 254, 156,
			233, 24, 202, 91, 43, 252, 121, 59, 134, 139, 126, 210, 174, 201,
			218, 91, 154, 181, 90, 94, 18, 124, 197, 77, 41, 79, 200, 243,
			228, 77, 23, 85, 97, 39, 108, 123, 223, 131, 247, 64, 131, 237,
			196, 130, 41, 31, 117, 183, 98, 134, 225, 5, 34, 59, 49, 206,
			176, 93, 170, 150, 31, 177, 219, 65, 40, 120, 194, 226, 76, 200,
			103, 254, 207, 50, 158, 169, 61, 86, 147, 135, 219, 32, 69, 151,
			23, 116, 45, 144, 73, 143, 49, 200, 143, 41, 83, 144, 7, 62,
			194, 100, 186, 58, 212, 47, 25, 46, 158, 235, 234, 151, 12, 79,
			222, 232, 238, 151, 12, 207, 145, 79, 117, 187, 196, 25, 181, 199,
			222, 45, 111, 244, 214, 36, 204, 131, 233, 181, 244, 57, 212, 245,
			24, 205, 139, 240, 224, 140, 209, 188, 8, 47, 187, 30, 195, 239,
			144, 105, 211, 244, 56, 99, 223, 44, 143, 51, 21, 83, 106, 155,
			22, 23, 73, 80, 75, 15, 181, 30, 206, 20, 75, 93, 173, 135,
			51, 35, 149, 174, 214, 195, 153, 249, 143, 200, 166, 110, 61, 56,
			101, 251, 220, 59, 229, 15, 95, 203, 186, 49, 75, 178, 232, 208,
			57, 100, 20, 29, 234, 37, 148, 139, 172, 171, 151, 80, 190, 176,
			210, 213, 75, 56, 55, 252, 54, 249, 15, 91, 53, 19, 166, 173,
			57, 84, 254, 55, 155, 245, 75, 172, 128, 249, 78, 189, 34, 146,
			165, 175, 36, 242, 66, 217, 42, 147, 182, 205, 203, 172, 160, 41,
			127, 174, 10, 171, 74, 91, 56, 133, 207, 213, 163, 83, 150, 198,
			131, 90, 19, 46, 236, 41, 145, 7, 81, 128, 0, 185, 112, 4,
			89, 37, 227, 9, 188, 3, 88, 26, 183, 56, 139, 101, 9, 173,
			115, 84, 22, 71, 178, 224, 43, 152, 240, 118, 121, 218, 193, 118,
			154, 191, 43, 214, 158, 62, 73, 217, 62, 127, 51, 225, 121, 75,
			195, 19, 170, 75, 198, 159, 123, 173, 118, 200, 89, 37, 130, 15,
			220, 23, 106, 243, 10, 108, 90, 73, 224, 123, 107, 126, 194, 240,
			23, 117, 249, 245, 172, 204, 194, 227, 26, 94, 39, 204, 75, 137,
			44, 150, 233, 87, 79, 167, 164, 46, 109, 35, 98, 217, 3, 76,
			98, 245, 250, 73, 91, 94, 34, 230, 59, 151, 45, 86, 231, 158,
			200, 100, 195, 199, 244, 106, 166, 221, 25, 221, 11, 177, 40, 158,
			177, 153, 105, 147, 20, 129, 234, 110, 161, 204, 232, 90, 169, 138,
			132, 153, 55, 166, 228, 13, 219, 2, 247, 93, 180, 175, 168, 247,
			131, 37, 69, 94, 28, 56, 75, 126, 31, 145, 34, 144, 32, 245,
			77, 167, 82, 22, 172, 231, 53, 142, 181, 184, 23, 165, 44, 138,
			149, 186, 172, 163, 110, 21, 222, 94, 178, 63, 145, 165, 58, 131,
			152, 60, 161, 82, 9, 184, 21, 134, 107, 73, 28, 177, 22, 175,
			53, 189, 40, 72, 91, 144, 216, 238, 172, 62, 170, 18, 114, 138,
			12, 40, 37, 16, 104, 113, 190, 67, 219, 20, 191, 201, 46, 144,
			111, 141, 150, 136, 226, 89, 231, 66, 57, 99, 189, 110, 147, 90,
			73, 137, 179, 125, 110, 16, 147, 123, 186, 22, 71, 17, 175, 153,
			162, 106, 119, 232, 123, 117, 161, 179, 115, 215, 96, 211, 131, 44,
			15, 242, 101, 165, 209, 168, 133, 164, 26, 147, 29, 218, 166, 120,
			118, 138, 117, 122, 85, 87, 237, 121, 211, 143, 42, 2, 53, 214,
			213, 171, 186, 122, 250, 114, 87, 175, 234, 234, 213, 57, 200, 218,
			14, 162, 184, 106, 93, 206, 203, 247, 85, 247, 162, 28, 182, 41,
			126, 203, 170, 230, 37, 206, 183, 220, 171, 114, 24, 83, 188, 104,
			45, 228, 133, 183, 69, 87, 9, 113, 40, 190, 102, 77, 230, 149,
			176, 107, 238, 89, 57, 92, 160, 248, 186, 117, 62, 47, 195, 92,
			119, 85, 121, 166, 72, 157, 27, 214, 50, 202, 203, 18, 55, 186,
			203, 18, 75, 246, 105, 83, 122, 40, 2, 53, 212, 85, 150, 88,
			58, 81, 234, 42, 75, 44, 141, 142, 201, 93, 6, 40, 254, 192,
			154, 207, 159, 207, 31, 184, 87, 96, 151, 130, 69, 157, 15, 173,
			187, 114, 151, 2, 48, 127, 88, 56, 9, 187, 20, 36, 226, 62,
			82, 239, 227, 130, 242, 252, 71, 246, 136, 161, 108, 138, 63, 82,
			15, 233, 130, 52, 232, 199, 246, 9, 61, 133, 36, 53, 96, 40,
			155, 226, 143, 201, 144, 94, 104, 83, 124, 83, 21, 56, 10, 42,
			159, 222, 180, 79, 25, 10, 230, 84, 37, 164, 32, 59, 116, 43,
			54, 213, 83, 96, 193, 21, 251, 164, 161, 108, 138, 87, 74, 35,
			122, 161, 67, 241, 106, 174, 21, 216, 116, 53, 87, 4, 124, 179,
			58, 92, 210, 11, 11, 20, 175, 229, 91, 131, 149, 215, 242, 173,
			11, 54, 197, 107, 249, 214, 69, 138, 111, 217, 103, 244, 20, 152,
			253, 86, 174, 72, 209, 166, 248, 214, 233, 113, 189, 112, 128, 226,
			219, 246, 132, 158, 2, 139, 222, 206, 229, 15, 216, 20, 223, 30,
			63, 163, 23, 186, 20, 223, 201, 101, 184, 8, 40, 115, 24, 215,
			166, 248, 78, 105, 68, 58, 2, 81, 231, 30, 220, 119, 192, 17,
			178, 156, 93, 56, 37, 5, 200, 14, 209, 134, 22, 160, 122, 66,
			27, 90, 0, 146, 142, 216, 208, 214, 144, 61, 161, 205, 124, 33,
			146, 148, 89, 8, 142, 216, 204, 23, 218, 20, 223, 215, 102, 83,
			125, 159, 251, 218, 108, 72, 58, 226, 190, 54, 155, 236, 251, 60,
			176, 75, 122, 10, 28, 241, 64, 33, 13, 40, 155, 226, 7, 167,
			134, 205, 147, 253, 127, 2, 0, 0, 255, 255, 55, 195, 200, 66,
			212, 45, 0, 0},
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
