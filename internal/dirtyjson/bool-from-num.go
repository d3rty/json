package dirtyjson

import (
	"github.com/d3rty/json/internal/config"
	"github.com/d3rty/json/internal/option"
)

// boolFromNum is a parser func to parse a bool from a number (float64)
type boolFromNum = func(i float64) option.Bool

// parsersBoolFromNum stores all registered boolFromNumber parsers.
// We're OK without mutex for now
var parsersBoolFromNum map[config.BoolFromNumberParser]boolFromNum

func init() {
	parsersBoolFromNum = make(map[config.BoolFromNumberParser]boolFromNum)

	parsersBoolFromNum[config.BoolFromNumberBinary] = func(i float64) option.Bool {
		if i == 0 {
			return option.False()
		} else if i == 1 {
			return option.True()
		}

		return option.NoneBool()
	}

	parsersBoolFromNum[config.BoolFromNumberPositiveNegative] = func(i float64) option.Bool {
		if i <= 0 {
			return option.False()
		}

		return option.True()
	}

	parsersBoolFromNum[config.BoolFromNumberSignOfOne] = func(i float64) option.Bool {
		if i == -1 {
			return option.False()
		} else if i == 1 {
			return option.True()
		}

		return option.NoneBool()
	}
}
