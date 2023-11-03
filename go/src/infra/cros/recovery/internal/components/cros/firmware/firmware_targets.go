// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package firmware

// targetOverrideModels holds a map of models that need to override its firmware target.
var targetOverrideModels = map[string]string{
	// TODO(b/226402941): Read existing ec image name using futility.
	"dragonair": "dratini",
	// Models that use _signed version of firmware.
	"drallion360": "drallion_signed",
	"sarien":      "sarien_signed",
	"arcada":      "arcada_signed",
	"drallion":    "drallion_signed",
	// Octopus board.
	"foob360":    "foob",
	"blooglet":   "bloog",
	"garg360":    "garg",
	"laser14":    "phaser",
	"bluebird":   "casta",
	"vorticon":   "meep",
	"dorp":       "meep",
	"orbatrix":   "fleex",
	"blooguard":  "bloog",
	"grabbiter":  "fleex",
	"apel":       "ampton",
	"nospike":    "ampton",
	"phaser360":  "phaser",
	"blorb":      "bobba",
	"droid":      "bobba",
	"garfour":    "garg",
	"vortininja": "meep",
	"sparky":     "bobba",
	"sparky360":  "bobba",
	"bobba360":   "bobba",
	"mimrock":    "meep",
	// Grunt board.
	"barla":     "careena",
	"kasumi":    "aleena",
	"kasumi360": "aleena",
	// Brya board.
	"zavala":   "volmar",
	"crota360": "crota",
	// Trogdor board.
	"pazquel360": "pazquel",
	"limozeen":   "lazor",
	// Volteer board.
	"lillipup": "lindar",
	// Screebo board
	"screebo4es": "screebo4es",
}

// targetOverrideHwidSkus holds a map of hwid that need to override its firmware target.
// Some latest models may uses different firmware based on there hwid, so decide firmware
// based on models are not sufficient for them.
var targetOverridebyHwid = map[string]string{
	// Nissa board.
	"JOXER-ZELG B3B-C3A-Q6B-I2C-C9Y-A9N":         "joxer_ufs",
	"NIRWEN-ZZCR B2B-B2A-B2A-W7H":                "nivviks_ufs",
	"NEREID-ZZCR A4B-B2C-E3E-F2A-A2B-48T":        "nereid_hdmi",
	"PUJJOTEEN-JQII E5B-D2J-E5D-Q2Q-K4E-I5I-A6I": "pujjo_5g",
	"YAVIKS-BVSW B3B-A2C-B2A-Q6W-I3T":            "yaviks_ufs",
	"YAVIKS-BVSW C4B-C2C-C3B-H4H-66N":            "yaviks_ufs",
	"YAVIKSO-OFRL C4C-C2F-D3B-W3P-N4E":           "yaviks_ufs",
	"YAVIKSO-OFRL C4C-B2D-D3B-Q3E-X66":           "yaviks_ufs",
	"YAVILLY-ITWJ B3B-G2H-C8K-79E-W4S":           "yavilla_ufs",
	"AVIJO-PORH B3C-G3K-O4K-79T-I3X":             "yavilla_ufs",
}

// ecExemptedModels holds a map of models that doesn't have EC firmware.
var ecExemptedModels = map[string]bool{
	"drallion360": true,
	"sarien":      true,
	"arcada":      true,
	"drallion":    true,
}
