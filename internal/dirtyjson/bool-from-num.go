package dirtyjson

import (
	"github.com/amberpixels/abu/maybe"
	"github.com/d3rty/json/internal/config"
)

// boolFromNum is a parser func to parse a bool from a number (float64).
type boolFromNum = func(i float64) maybe.Bool

// We're OK without mutex for now.
//
//nolint:gochecknoglobals // we're ok with it as well
var (
	parsersBoolFromNum = map[config.BoolFromNumberAlg]boolFromNum{

		config.BoolFromNumberBinary: func(i float64) maybe.Bool {
			switch i {
			case 0:
				return maybe.False()
			case 1:
				return maybe.True()
			default:
				return maybe.NoneBool()
			}
		},

		config.BoolFromNumberPositiveNegative: func(i float64) maybe.Bool {
			if i <= 0 {
				return maybe.False()
			}

			return maybe.True()
		},

		config.BoolFromNumberSignOfOne: func(i float64) maybe.Bool {
			switch i {
			case -1:
				return maybe.False()
			case 1:
				return maybe.True()
			default:
				return maybe.NoneBool()
			}
		},

		config.BoolFromNumberUndefined: nil,
	}
)
