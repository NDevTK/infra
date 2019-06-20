// Code generated by cproto. DO NOT EDIT.

package api

import discovery "go.chromium.org/luci/grpc/discovery"

import "github.com/golang/protobuf/protoc-gen-go/descriptor"

func init() {
	discovery.RegisterDescriptorSetCompressed(
		[]string{
			"drone_queen.Drone", "drone_queen.InventoryProvider",
		},
		[]byte{31, 139,
			8, 0, 0, 0, 0, 0, 0, 255, 164, 90, 61, 112, 27, 73,
			118, 198, 252, 16, 164, 90, 255, 173, 63, 46, 184, 171, 125, 34,
			181, 18, 184, 75, 14, 41, 82, 210, 89, 148, 118, 207, 32, 48,
			164, 102, 23, 4, 112, 3, 64, 92, 237, 149, 75, 59, 196, 52,
			136, 89, 13, 102, 112, 51, 13, 96, 233, 114, 249, 39, 119, 57,
			112, 228, 192, 129, 203, 129, 131, 187, 216, 85, 174, 114, 116, 185,
			35, 7, 46, 71, 142, 92, 229, 212, 161, 157, 185, 94, 119, 15,
			0, 82, 226, 233, 238, 140, 34, 74, 120, 211, 175, 191, 247, 211,
			175, 251, 189, 126, 35, 242, 239, 15, 200, 167, 199, 113, 124, 28,
			178, 141, 65, 18, 243, 248, 104, 216, 221, 224, 65, 159, 165, 220,
			235, 15, 44, 241, 136, 94, 149, 12, 86, 198, 176, 252, 156, 92,
			104, 101, 60, 116, 145, 204, 167, 172, 19, 71, 126, 186, 168, 129,
			86, 52, 220, 140, 164, 55, 201, 92, 228, 69, 113, 186, 168, 131,
			86, 156, 115, 37, 177, 251, 167, 228, 70, 39, 238, 91, 103, 48,
			119, 175, 76, 16, 27, 248, 168, 161, 125, 247, 197, 113, 192, 123,
			195, 35, 171, 19, 247, 55, 142, 227, 208, 139, 142, 167, 42, 14,
			248, 201, 128, 165, 83, 77, 255, 71, 211, 254, 94, 55, 246, 27,
			187, 191, 212, 239, 238, 75, 228, 134, 226, 181, 14, 89, 24, 126,
			19, 197, 227, 168, 133, 115, 190, 254, 143, 251, 36, 79, 205, 187,
			185, 183, 26, 249, 151, 75, 68, 187, 68, 141, 187, 57, 186, 245,
			235, 75, 32, 38, 116, 226, 16, 118, 135, 221, 46, 75, 82, 88,
			7, 9, 245, 48, 5, 223, 227, 30, 4, 17, 103, 73, 167, 231,
			69, 199, 12, 186, 113, 210, 247, 56, 129, 114, 60, 56, 73, 130,
			227, 30, 135, 173, 205, 205, 63, 80, 19, 192, 137, 58, 22, 64,
			41, 12, 65, 140, 165, 144, 176, 148, 37, 35, 230, 91, 4, 122,
			156, 15, 210, 157, 141, 13, 159, 141, 88, 24, 15, 88, 146, 102,
			190, 64, 67, 7, 74, 137, 245, 35, 169, 196, 6, 33, 224, 50,
			63, 72, 121, 18, 28, 13, 121, 16, 71, 224, 69, 62, 12, 83,
			6, 65, 4, 105, 60, 76, 58, 76, 60, 57, 10, 34, 47, 57,
			17, 122, 165, 107, 48, 14, 120, 15, 226, 68, 252, 27, 15, 57,
			129, 126, 236, 7, 221, 160, 227, 33, 194, 26, 120, 9, 131, 1,
			75, 250, 1, 231, 204, 135, 65, 18, 143, 2, 159, 249, 192, 123,
			30, 7, 222, 67, 235, 194, 48, 30, 7, 209, 49, 224, 74, 6,
			56, 41, 197, 73, 4, 250, 140, 239, 16, 2, 248, 249, 252, 140,
			98, 41, 196, 221, 76, 163, 78, 236, 51, 232, 15, 83, 14, 9,
			227, 94, 16, 9, 84, 239, 40, 30, 225, 144, 242, 24, 129, 40,
			230, 65, 135, 173, 1, 239, 5, 41, 132, 65, 202, 17, 97, 86,
			98, 228, 159, 81, 199, 15, 210, 78, 232, 5, 125, 150, 88, 231,
			41, 17, 68, 179, 190, 200, 148, 24, 36, 177, 63, 236, 176, 169,
			30, 100, 170, 200, 255, 75, 15, 2, 202, 58, 63, 238, 12, 251,
			44, 226, 94, 182, 72, 27, 113, 2, 49, 239, 177, 4, 250, 30,
			103, 73, 224, 133, 233, 212, 213, 98, 129, 120, 143, 17, 152, 213,
			126, 98, 84, 141, 5, 98, 38, 2, 71, 94, 159, 161, 66, 179,
			177, 21, 197, 211, 49, 225, 247, 128, 167, 104, 81, 36, 161, 226,
			36, 133, 190, 119, 2, 71, 12, 35, 197, 7, 30, 3, 139, 252,
			56, 73, 25, 6, 197, 32, 137, 251, 49, 103, 32, 125, 194, 83,
			240, 89, 18, 140, 152, 15, 221, 36, 238, 19, 233, 133, 52, 238,
			242, 49, 134, 137, 138, 32, 72, 7, 172, 131, 17, 4, 131, 36,
			192, 192, 74, 48, 118, 34, 25, 69, 105, 42, 116, 39, 208, 122,
			233, 52, 161, 89, 223, 107, 29, 150, 92, 27, 156, 38, 52, 220,
			250, 43, 167, 98, 87, 96, 247, 53, 180, 94, 218, 80, 174, 55,
			94, 187, 206, 254, 203, 22, 188, 172, 87, 43, 182, 219, 132, 82,
			173, 2, 229, 122, 173, 229, 58, 187, 237, 86, 221, 109, 18, 88,
			46, 53, 193, 105, 46, 139, 145, 82, 237, 53, 216, 223, 54, 92,
			187, 217, 132, 186, 11, 206, 65, 163, 234, 216, 21, 56, 44, 185,
			110, 169, 214, 114, 236, 230, 26, 56, 181, 114, 181, 93, 113, 106,
			251, 107, 176, 219, 110, 65, 173, 222, 34, 80, 117, 14, 156, 150,
			93, 129, 86, 125, 77, 136, 125, 119, 30, 212, 247, 224, 192, 118,
			203, 47, 75, 181, 86, 105, 215, 169, 58, 173, 215, 66, 224, 158,
			211, 170, 161, 176, 189, 186, 75, 160, 4, 141, 146, 219, 114, 202,
			237, 106, 201, 133, 70, 219, 109, 212, 155, 54, 160, 101, 21, 167,
			89, 174, 150, 156, 3, 187, 98, 129, 83, 131, 90, 29, 236, 87,
			118, 173, 5, 205, 151, 165, 106, 245, 180, 161, 4, 234, 135, 53,
			219, 69, 237, 103, 205, 132, 93, 27, 170, 78, 105, 183, 106, 163,
			40, 97, 103, 197, 113, 237, 114, 11, 13, 154, 254, 42, 59, 21,
			187, 214, 42, 85, 215, 8, 52, 27, 118, 217, 41, 85, 215, 192,
			254, 214, 62, 104, 84, 75, 238, 235, 53, 5, 218, 180, 127, 214,
			182, 107, 45, 167, 84, 133, 74, 233, 160, 180, 111, 55, 161, 248,
			33, 175, 52, 220, 122, 185, 237, 218, 7, 168, 117, 125, 15, 154,
			237, 221, 102, 203, 105, 181, 91, 54, 236, 215, 235, 21, 225, 236,
			166, 237, 190, 114, 202, 118, 243, 57, 84, 235, 77, 225, 176, 118,
			211, 94, 35, 80, 41, 181, 74, 66, 116, 195, 173, 239, 57, 173,
			230, 115, 252, 189, 219, 110, 58, 194, 113, 78, 173, 101, 187, 110,
			187, 209, 114, 234, 181, 85, 120, 89, 63, 180, 95, 217, 46, 148,
			75, 237, 166, 93, 17, 30, 174, 215, 208, 90, 140, 21, 187, 238,
			190, 70, 88, 244, 131, 88, 129, 53, 56, 124, 105, 183, 94, 218,
			46, 58, 85, 120, 171, 132, 110, 104, 182, 92, 167, 220, 154, 101,
			171, 187, 208, 170, 187, 45, 50, 99, 39, 212, 236, 253, 170, 179,
			111, 215, 202, 54, 14, 215, 17, 230, 208, 105, 218, 171, 80, 114,
			157, 38, 50, 56, 66, 48, 28, 150, 94, 67, 189, 45, 172, 198,
			133, 106, 55, 109, 34, 127, 207, 132, 238, 154, 88, 79, 112, 246,
			160, 84, 121, 229, 160, 230, 138, 187, 81, 111, 54, 29, 21, 46,
			194, 109, 229, 151, 202, 231, 22, 33, 11, 68, 211, 169, 1, 11,
			119, 240, 215, 2, 53, 150, 115, 207, 201, 69, 98, 46, 252, 215,
			124, 78, 18, 151, 200, 28, 18, 58, 53, 150, 231, 239, 144, 203,
			36, 47, 168, 156, 36, 175, 144, 121, 73, 106, 146, 86, 204, 243,
			212, 88, 46, 236, 40, 196, 149, 220, 167, 10, 81, 147, 132, 100,
			66, 177, 43, 19, 68, 13, 17, 87, 38, 136, 154, 64, 92, 153,
			32, 106, 6, 53, 86, 10, 119, 21, 226, 253, 220, 174, 66, 212,
			37, 33, 153, 116, 164, 230, 175, 43, 68, 29, 17, 145, 148, 136,
			186, 64, 68, 90, 49, 207, 83, 227, 254, 205, 146, 66, 252, 44,
			183, 166, 16, 13, 73, 72, 38, 67, 167, 198, 103, 243, 55, 20,
			162, 129, 136, 72, 74, 68, 67, 32, 34, 173, 152, 231, 169, 241,
			217, 237, 47, 20, 226, 131, 220, 134, 66, 52, 37, 33, 153, 76,
			157, 26, 15, 230, 151, 20, 162, 137, 136, 72, 74, 68, 83, 32,
			34, 173, 152, 231, 169, 241, 224, 174, 165, 16, 31, 230, 150, 21,
			226, 156, 36, 36, 211, 156, 78, 141, 135, 243, 5, 133, 56, 135,
			136, 72, 74, 196, 57, 129, 136, 180, 98, 54, 168, 241, 240, 147,
			123, 10, 177, 152, 187, 167, 16, 243, 146, 144, 76, 121, 157, 26,
			197, 249, 69, 133, 152, 71, 68, 36, 37, 98, 94, 32, 34, 173,
			152, 231, 169, 81, 92, 2, 242, 79, 87, 137, 110, 230, 168, 249,
			38, 247, 86, 43, 252, 234, 42, 148, 96, 82, 27, 137, 76, 198,
			82, 22, 241, 20, 60, 24, 196, 65, 196, 69, 254, 9, 250, 88,
			15, 248, 108, 192, 34, 159, 69, 34, 127, 121, 209, 137, 124, 254,
			199, 113, 196, 8, 158, 251, 29, 47, 100, 145, 239, 37, 107, 83,
			20, 230, 131, 151, 130, 42, 216, 68, 158, 235, 38, 94, 103, 154,
			205, 179, 1, 76, 214, 88, 189, 9, 26, 171, 153, 56, 148, 197,
			72, 16, 65, 187, 85, 6, 123, 16, 119, 122, 66, 156, 5, 14,
			135, 32, 5, 22, 97, 13, 128, 149, 10, 230, 75, 145, 233, 26,
			73, 28, 178, 1, 15, 58, 176, 159, 176, 227, 56, 9, 188, 8,
			202, 74, 39, 24, 247, 130, 78, 15, 216, 143, 156, 161, 64, 204,
			109, 83, 166, 76, 113, 2, 71, 94, 231, 237, 216, 75, 144, 35,
			134, 19, 230, 37, 16, 71, 239, 136, 244, 210, 116, 216, 71, 169,
			94, 24, 66, 63, 136, 134, 156, 137, 234, 5, 158, 110, 146, 137,
			73, 97, 28, 29, 175, 65, 96, 49, 11, 66, 230, 13, 166, 166,
			38, 12, 150, 211, 62, 243, 18, 230, 47, 67, 26, 203, 162, 40,
			138, 103, 185, 8, 112, 239, 40, 100, 40, 51, 98, 12, 69, 118,
			227, 68, 150, 135, 3, 172, 119, 68, 42, 7, 87, 20, 138, 65,
			170, 210, 234, 230, 230, 230, 163, 117, 241, 215, 218, 220, 220, 17,
			127, 223, 161, 21, 207, 158, 61, 123, 182, 254, 104, 107, 125, 251,
			81, 107, 107, 123, 231, 201, 179, 157, 39, 207, 172, 103, 217, 231,
			59, 139, 192, 238, 9, 58, 156, 39, 65, 135, 11, 87, 42, 149,
			18, 132, 95, 131, 49, 3, 22, 165, 195, 132, 201, 167, 99, 6,
			29, 244, 88, 28, 141, 88, 194, 129, 199, 68, 173, 106, 220, 7,
			112, 247, 202, 176, 189, 189, 253, 12, 203, 89, 6, 8, 25, 29,
			167, 22, 129, 38, 99, 240, 243, 172, 46, 29, 143, 199, 86, 192,
			120, 215, 138, 147, 227, 141, 164, 219, 193, 47, 78, 178, 248, 143,
			252, 143, 138, 191, 13, 215, 42, 150, 2, 246, 143, 94, 127, 16,
			50, 120, 180, 3, 229, 184, 63, 24, 114, 54, 19, 197, 66, 157,
			70, 189, 233, 124, 11, 223, 99, 208, 20, 87, 191, 183, 84, 85,
			57, 101, 154, 20, 247, 207, 229, 200, 244, 90, 146, 50, 254, 70,
			173, 87, 81, 76, 175, 181, 171, 213, 213, 213, 247, 242, 137, 176,
			45, 110, 174, 62, 159, 209, 105, 235, 67, 58, 29, 51, 142, 40,
			113, 215, 247, 78, 102, 116, 75, 121, 50, 236, 112, 33, 96, 228,
			133, 192, 71, 74, 226, 41, 246, 7, 124, 180, 6, 66, 161, 231,
			191, 175, 73, 35, 139, 143, 144, 250, 77, 22, 73, 166, 97, 202,
			58, 240, 57, 60, 218, 220, 60, 109, 225, 246, 185, 22, 30, 6,
			209, 246, 22, 124, 191, 207, 120, 243, 36, 229, 172, 143, 195, 165,
			116, 47, 8, 89, 235, 244, 66, 236, 57, 85, 187, 229, 28, 216,
			208, 229, 74, 141, 243, 230, 60, 232, 242, 76, 211, 182, 83, 107,
			61, 125, 12, 60, 232, 188, 77, 225, 75, 40, 22, 139, 242, 201,
			106, 151, 91, 254, 248, 101, 112, 220, 171, 120, 92, 204, 90, 133,
			23, 47, 96, 123, 107, 21, 254, 4, 196, 88, 53, 30, 103, 67,
			153, 223, 54, 54, 160, 132, 250, 250, 241, 56, 21, 144, 184, 153,
			30, 109, 110, 206, 28, 69, 169, 53, 97, 96, 226, 8, 122, 244,
			244, 221, 93, 54, 65, 195, 233, 143, 158, 62, 126, 252, 248, 39,
			219, 79, 55, 55, 39, 91, 254, 136, 117, 227, 132, 65, 59, 10,
			126, 204, 80, 158, 253, 100, 243, 44, 138, 245, 251, 45, 102, 81,
			218, 15, 197, 162, 116, 202, 134, 88, 44, 252, 172, 194, 250, 172,
			58, 31, 136, 96, 196, 65, 119, 101, 56, 159, 205, 224, 136, 0,
			88, 61, 21, 0, 143, 207, 13, 128, 175, 189, 145, 7, 223, 203,
			133, 180, 58, 195, 36, 97, 17, 71, 150, 131, 32, 12, 131, 116,
			38, 0, 240, 132, 132, 190, 120, 10, 95, 194, 249, 19, 126, 67,
			152, 195, 151, 211, 167, 86, 196, 198, 187, 195, 32, 244, 89, 82,
			92, 69, 195, 154, 202, 67, 74, 132, 116, 204, 170, 196, 194, 15,
			242, 212, 164, 237, 65, 196, 209, 114, 197, 41, 77, 87, 102, 11,
			15, 172, 90, 71, 136, 44, 116, 153, 250, 224, 201, 185, 62, 80,
			86, 100, 121, 19, 26, 39, 188, 39, 111, 48, 167, 220, 63, 171,
			126, 113, 245, 236, 218, 236, 51, 94, 158, 122, 163, 184, 74, 196,
			199, 48, 49, 169, 191, 89, 184, 78, 254, 86, 35, 166, 41, 202,
			59, 95, 191, 89, 248, 43, 13, 220, 105, 238, 206, 66, 47, 238,
			138, 244, 41, 244, 72, 131, 168, 51, 27, 133, 228, 253, 97, 8,
			7, 120, 163, 61, 98, 210, 146, 115, 178, 10, 121, 95, 90, 249,
			14, 130, 168, 19, 14, 211, 96, 196, 44, 66, 46, 147, 57, 212,
			206, 164, 166, 175, 191, 17, 133, 23, 146, 115, 168, 237, 124, 70,
			105, 212, 240, 23, 174, 102, 148, 65, 13, 159, 222, 32, 255, 41,
			237, 210, 168, 241, 131, 78, 11, 255, 166, 65, 45, 142, 214, 35,
			118, 236, 241, 96, 196, 78, 215, 15, 158, 50, 20, 48, 133, 190,
			175, 126, 176, 160, 166, 38, 102, 153, 25, 70, 94, 56, 100, 169,
			188, 32, 79, 193, 196, 53, 62, 229, 65, 24, 66, 207, 27, 49,
			136, 102, 101, 10, 104, 53, 145, 200, 60, 216, 137, 135, 17, 199,
			180, 140, 213, 66, 86, 34, 157, 245, 157, 74, 191, 107, 234, 75,
			222, 227, 31, 205, 164, 230, 15, 186, 127, 83, 249, 64, 155, 67,
			171, 51, 255, 104, 232, 131, 133, 203, 25, 101, 80, 227, 135, 107,
			215, 143, 242, 162, 135, 179, 77, 254, 23, 200, 122, 16, 117, 19,
			111, 195, 27, 12, 88, 116, 28, 68, 108, 195, 79, 226, 136, 173,
			255, 98, 200, 88, 180, 225, 13, 130, 141, 148, 37, 163, 160, 163,
			218, 96, 244, 162, 24, 126, 35, 134, 11, 31, 106, 203, 45, 255,
			90, 35, 212, 101, 131, 56, 225, 21, 156, 230, 178, 95, 12, 89,
			202, 233, 39, 132, 72, 152, 225, 48, 240, 69, 75, 238, 130, 123,
			65, 60, 105, 15, 3, 159, 30, 146, 171, 97, 236, 249, 111, 130,
			200, 15, 58, 30, 143, 19, 217, 158, 187, 184, 101, 89, 51, 210,
			173, 119, 129, 173, 106, 236, 249, 206, 100, 150, 123, 37, 60, 69,
			23, 182, 201, 149, 211, 28, 244, 30, 185, 228, 15, 249, 155, 142,
			55, 240, 58, 1, 63, 17, 186, 92, 118, 47, 250, 67, 94, 86,
			143, 150, 255, 89, 39, 55, 78, 137, 74, 7, 113, 148, 50, 250,
			83, 146, 79, 185, 199, 135, 178, 167, 120, 101, 235, 225, 249, 202,
			201, 25, 86, 83, 176, 187, 106, 218, 25, 47, 232, 103, 189, 80,
			38, 87, 217, 143, 131, 32, 17, 165, 218, 27, 244, 236, 162, 33,
			188, 80, 56, 219, 152, 180, 38, 39, 128, 123, 101, 58, 5, 31,
			210, 21, 114, 217, 75, 211, 224, 56, 98, 254, 27, 127, 200, 211,
			69, 19, 140, 226, 5, 247, 82, 246, 176, 50, 228, 41, 50, 249,
			137, 23, 68, 65, 116, 44, 153, 230, 36, 83, 246, 16, 153, 150,
			159, 144, 188, 212, 159, 94, 39, 151, 219, 181, 111, 106, 245, 195,
			218, 27, 219, 117, 235, 238, 181, 28, 205, 19, 189, 254, 205, 53,
			141, 94, 35, 151, 178, 161, 118, 219, 169, 92, 211, 151, 247, 49,
			0, 66, 230, 165, 12, 81, 126, 203, 0, 160, 196, 20, 122, 232,
			66, 15, 241, 123, 249, 22, 174, 194, 12, 144, 244, 233, 114, 145,
			208, 10, 235, 132, 94, 114, 10, 63, 3, 208, 78, 3, 156, 226,
			148, 0, 91, 191, 210, 200, 156, 88, 38, 218, 32, 23, 103, 86,
			141, 126, 250, 129, 96, 43, 192, 135, 22, 92, 34, 78, 116, 126,
			7, 241, 172, 91, 222, 65, 124, 199, 220, 45, 70, 174, 59, 209,
			136, 69, 60, 78, 78, 26, 178, 81, 151, 160, 152, 25, 203, 206,
			136, 121, 215, 59, 103, 196, 188, 199, 41, 187, 115, 223, 25, 222,
			32, 248, 250, 95, 23, 73, 158, 154, 102, 206, 189, 75, 254, 81,
			19, 125, 104, 51, 71, 183, 126, 169, 157, 106, 41, 63, 122, 6,
			173, 30, 131, 106, 187, 236, 64, 105, 200, 123, 113, 146, 90, 231,
			244, 149, 219, 169, 232, 18, 170, 238, 221, 180, 11, 27, 164, 112,
			28, 143, 88, 18, 225, 101, 44, 242, 85, 83, 177, 52, 240, 58,
			8, 28, 116, 88, 148, 178, 53, 120, 197, 146, 20, 239, 115, 91,
			214, 102, 118, 130, 122, 145, 56, 41, 227, 97, 228, 103, 61, 206,
			170, 83, 182, 107, 77, 27, 186, 65, 136, 71, 228, 5, 162, 27,
			57, 106, 228, 231, 139, 170, 247, 177, 176, 112, 67, 221, 135, 73,
			174, 48, 237, 125, 32, 49, 237, 125, 144, 73, 95, 65, 36, 71,
			50, 233, 43, 200, 132, 67, 38, 125, 133, 220, 60, 53, 200, 205,
			143, 8, 33, 122, 62, 71, 205, 75, 185, 219, 26, 38, 216, 60,
			114, 93, 90, 184, 76, 254, 65, 35, 102, 94, 96, 80, 189, 82,
			248, 27, 145, 96, 179, 48, 65, 179, 59, 94, 24, 50, 31, 142,
			78, 64, 44, 137, 184, 51, 38, 130, 5, 194, 96, 196, 34, 150,
			202, 43, 239, 49, 227, 80, 105, 183, 8, 200, 141, 219, 199, 12,
			141, 181, 0, 94, 139, 208, 106, 215, 46, 85, 176, 14, 142, 19,
			240, 25, 247, 130, 48, 133, 88, 250, 67, 52, 92, 61, 188, 18,
			100, 157, 115, 33, 73, 100, 43, 162, 218, 197, 22, 65, 115, 242,
			210, 56, 154, 191, 158, 81, 58, 53, 40, 189, 159, 81, 6, 53,
			232, 198, 46, 169, 10, 139, 52, 106, 220, 210, 43, 133, 159, 194,
			76, 148, 158, 111, 144, 96, 129, 120, 28, 177, 36, 237, 5, 3,
			12, 130, 74, 187, 149, 78, 228, 98, 150, 186, 53, 145, 139, 203,
			116, 107, 34, 23, 115, 214, 173, 141, 93, 225, 98, 141, 154, 139,
			185, 143, 165, 139, 113, 206, 226, 194, 71, 228, 136, 152, 121, 209,
			79, 90, 210, 43, 133, 54, 204, 132, 51, 112, 22, 134, 242, 162,
			174, 146, 24, 120, 71, 241, 144, 139, 75, 183, 136, 67, 38, 212,
			0, 111, 228, 5, 161, 184, 46, 139, 203, 168, 112, 49, 42, 46,
			77, 80, 90, 202, 38, 213, 146, 210, 82, 19, 222, 89, 82, 90,
			106, 194, 59, 75, 82, 75, 51, 71, 205, 187, 185, 85, 109, 82,
			105, 221, 93, 40, 144, 191, 156, 84, 90, 247, 244, 197, 194, 159,
			193, 244, 248, 67, 175, 161, 38, 120, 96, 66, 118, 46, 203, 43,
			180, 90, 43, 11, 160, 198, 198, 153, 67, 211, 94, 60, 12, 125,
			130, 151, 253, 17, 147, 123, 137, 245, 7, 252, 228, 57, 120, 16,
			177, 177, 196, 25, 99, 21, 114, 196, 206, 193, 155, 173, 172, 238,
			233, 119, 63, 158, 169, 172, 238, 233, 11, 51, 149, 213, 189, 11,
			55, 102, 42, 171, 123, 183, 239, 144, 231, 89, 97, 181, 162, 127,
			86, 176, 224, 76, 158, 22, 173, 9, 209, 206, 199, 72, 196, 65,
			56, 242, 66, 47, 234, 4, 209, 241, 169, 122, 101, 69, 191, 183,
			152, 213, 36, 121, 4, 187, 54, 83, 175, 172, 92, 135, 153, 122,
			101, 101, 5, 125, 108, 154, 57, 35, 71, 205, 251, 122, 209, 144,
			99, 134, 104, 240, 145, 69, 146, 146, 60, 82, 162, 19, 102, 126,
			92, 240, 97, 54, 161, 163, 115, 61, 72, 3, 81, 108, 11, 125,
			38, 202, 202, 13, 161, 72, 150, 66, 47, 30, 67, 223, 139, 78,
			8, 240, 152, 123, 161, 140, 140, 233, 126, 193, 179, 38, 29, 14,
			112, 107, 90, 132, 92, 37, 243, 82, 168, 73, 205, 135, 230, 125,
			209, 128, 147, 15, 230, 80, 13, 50, 165, 53, 106, 60, 188, 120,
			103, 74, 27, 212, 120, 88, 88, 18, 113, 162, 81, 243, 243, 220,
			174, 140, 19, 180, 251, 243, 133, 37, 226, 17, 211, 20, 209, 188,
			174, 223, 44, 180, 64, 150, 10, 234, 72, 80, 161, 44, 31, 169,
			248, 197, 253, 102, 129, 234, 46, 5, 125, 100, 243, 240, 214, 16,
			67, 167, 199, 58, 111, 213, 171, 18, 92, 13, 150, 36, 120, 52,
			203, 85, 208, 132, 230, 235, 250, 231, 159, 8, 111, 106, 122, 46,
			143, 34, 23, 50, 74, 163, 198, 250, 133, 171, 25, 101, 80, 99,
			157, 222, 16, 171, 160, 97, 116, 91, 250, 83, 185, 10, 154, 136,
			111, 107, 254, 50, 249, 115, 157, 228, 145, 68, 213, 183, 205, 219,
			133, 255, 214, 224, 84, 145, 160, 194, 22, 162, 152, 79, 94, 249,
			68, 113, 210, 247, 194, 240, 100, 162, 191, 240, 54, 235, 122, 195,
			144, 19, 89, 43, 67, 208, 157, 53, 58, 72, 65, 188, 202, 137,
			142, 33, 78, 96, 24, 189, 141, 226, 113, 100, 1, 236, 161, 125,
			242, 74, 181, 166, 166, 144, 201, 158, 31, 166, 44, 85, 123, 131,
			69, 195, 190, 2, 158, 28, 135, 157, 48, 192, 123, 150, 31, 179,
			84, 104, 135, 152, 68, 29, 20, 39, 140, 175, 205, 50, 137, 109,
			53, 76, 217, 172, 166, 18, 207, 146, 75, 174, 169, 141, 179, 109,
			94, 159, 210, 58, 53, 182, 111, 222, 194, 132, 34, 104, 141, 26,
			143, 205, 139, 147, 97, 77, 208, 249, 41, 173, 83, 227, 241, 5,
			50, 97, 215, 169, 241, 196, 188, 53, 25, 198, 233, 79, 204, 107,
			83, 26, 199, 111, 220, 36, 127, 167, 137, 200, 209, 168, 177, 163,
			47, 22, 254, 90, 251, 93, 79, 24, 167, 59, 59, 99, 236, 165,
			232, 64, 158, 101, 213, 68, 150, 14, 234, 253, 99, 55, 96, 161,
			47, 157, 161, 94, 2, 246, 228, 89, 196, 32, 245, 250, 76, 121,
			56, 78, 8, 46, 117, 44, 95, 225, 78, 2, 15, 183, 255, 206,
			36, 128, 196, 117, 101, 103, 18, 120, 232, 140, 29, 117, 232, 104,
			98, 251, 239, 220, 190, 67, 246, 132, 105, 58, 53, 94, 232, 155,
			133, 103, 112, 166, 44, 70, 243, 198, 61, 166, 210, 157, 186, 54,
			79, 243, 164, 100, 103, 211, 200, 215, 77, 106, 190, 208, 119, 22,
			149, 16, 61, 143, 184, 75, 25, 165, 81, 227, 197, 199, 95, 100,
			148, 65, 141, 23, 214, 6, 249, 67, 161, 128, 65, 141, 175, 244,
			251, 133, 109, 56, 85, 82, 139, 51, 111, 154, 74, 206, 57, 112,
			37, 158, 97, 34, 196, 132, 154, 163, 198, 87, 23, 175, 103, 148,
			70, 141, 175, 232, 167, 25, 133, 194, 150, 87, 72, 34, 36, 155,
			212, 40, 233, 247, 11, 8, 55, 83, 167, 159, 150, 124, 38, 191,
			171, 253, 38, 38, 88, 0, 45, 92, 183, 32, 37, 224, 133, 99,
			239, 68, 28, 138, 195, 35, 92, 96, 188, 2, 207, 154, 51, 209,
			213, 20, 66, 39, 212, 28, 53, 74, 19, 93, 77, 141, 26, 165,
			137, 174, 166, 65, 141, 210, 242, 138, 56, 211, 116, 106, 86, 114,
			95, 203, 51, 13, 125, 89, 89, 40, 144, 23, 196, 52, 197, 251,
			153, 61, 125, 177, 176, 241, 187, 5, 166, 92, 52, 93, 28, 87,
			123, 122, 69, 166, 42, 93, 28, 179, 123, 42, 106, 228, 187, 158,
			61, 21, 53, 186, 56, 174, 246, 110, 223, 33, 63, 23, 98, 53,
			106, 56, 250, 82, 161, 6, 194, 99, 211, 106, 100, 114, 232, 224,
			158, 247, 34, 121, 60, 226, 217, 225, 161, 59, 39, 3, 83, 165,
			200, 59, 235, 169, 99, 44, 27, 142, 62, 161, 230, 168, 225, 40,
			31, 233, 34, 148, 29, 122, 59, 163, 12, 106, 56, 31, 21, 176,
			30, 69, 119, 125, 147, 187, 43, 92, 132, 139, 254, 205, 130, 76,
			7, 38, 53, 15, 114, 63, 147, 174, 67, 7, 31, 44, 20, 200,
			95, 224, 174, 22, 239, 141, 26, 250, 82, 129, 75, 35, 48, 167,
			157, 169, 96, 120, 140, 91, 173, 239, 249, 236, 84, 49, 147, 85,
			48, 32, 184, 72, 182, 23, 101, 111, 97, 250, 95, 9, 178, 106,
			65, 4, 11, 243, 69, 197, 233, 179, 144, 201, 109, 139, 6, 152,
			184, 0, 70, 67, 159, 80, 115, 212, 104, 40, 83, 229, 123, 172,
			134, 50, 213, 20, 254, 111, 40, 83, 231, 168, 225, 42, 83, 231,
			52, 106, 184, 11, 75, 89, 239, 225, 255, 2, 0, 0, 255, 255,
			50, 213, 145, 114, 215, 35, 0, 0},
	)
}

// FileDescriptorSet returns a descriptor set for this proto package, which
// includes all defined services, and all transitive dependencies.
//
// Will not return nil.
//
// Do NOT modify the returned descriptor.
func FileDescriptorSet() *descriptor.FileDescriptorSet {
	// We just need ONE of the service names to look up the FileDescriptorSet.
	ret, err := discovery.GetDescriptorSet("drone_queen.Drone")
	if err != nil {
		panic(err)
	}
	return ret
}
