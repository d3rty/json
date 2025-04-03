package dirtyjson

import (
	"github.com/d3rty/json/internal/config"
	"github.com/d3rty/json/internal/option"
)

// boolFromNum is a parser func to parse a bool from a number (float64).
type boolFromNum = func(i float64) option.Bool

// We're OK without mutex for now.
//
//nolint:gochecknoglobals // we're ok with it as well
var (
	parsersBoolFromNum = map[config.BoolFromNumberAlg]boolFromNum{

		config.BoolFromNumberBinary: func(i float64) option.Bool {
			if i == 0 {
				return option.False()
			} else if i == 1 {
				return option.True()
			}

			return option.NoneBool()
		},

		config.BoolFromNumberPositiveNegative: func(i float64) option.Bool {
			if i <= 0 {
				return option.False()
			}

			return option.True()
		},

		config.BoolFromNumberSignOfOne: func(i float64) option.Bool {
			if i == -1 {
				return option.False()
			} else if i == 1 {
				return option.True()
			}

			return option.NoneBool()
		},

		config.BoolFromNumberUndefined: nil,
	}
)
