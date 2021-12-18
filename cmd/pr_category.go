package cmd

import (
	"fmt"

	"github.com/enescakir/emoji"
)

// According to https://stackoverflow.com/questions/14426366/what-is-an-idiomatic-way-of-representing-enums-in-go
// this is how you define enums in Go.
type PrCategory int

const (
	GENERAL PrCategory = iota
	FEATURES
	BUG_FIXES
	MAINTENANCE
	DOCUMENTATION
	LOGISTICS
	PERFORMANCE
	TESTS
)

func (c PrCategory) String() string {
	switch c {
	case GENERAL:
		return "GENERAL"
	case FEATURES:
		return "FEATURES"
	case BUG_FIXES:
		return "BUG_FIXES"
	case MAINTENANCE:
		return "MAINTENANCE"
	case DOCUMENTATION:
		return "DOCUMENTATION"
	case LOGISTICS:
		return "LOGISTICS"
	case PERFORMANCE:
		return "PERFORMANCE"
	case TESTS:
		return "TESTS"
	default:
		return ""
	}
}

func (c PrCategory) Pretty() string {
	switch c {
	case GENERAL:
		return fmt.Sprintf(`%s Other Changes`, emoji.PuzzlePiece)
	case FEATURES:
		return fmt.Sprintf(`%s Features`, emoji.Rocket)
	case BUG_FIXES:
		return fmt.Sprintf(`%s Bug Fixes`, emoji.Bug)
	case MAINTENANCE:
		return fmt.Sprintf(`%s Maintenance`, emoji.Toolbox)
	case DOCUMENTATION:
		return fmt.Sprintf(`%s Documentation`, emoji.Books)
	case LOGISTICS:
		return fmt.Sprintf(`%s Logistics`, emoji.Gear)
	case PERFORMANCE:
		return fmt.Sprintf(`%s Performance`, emoji.Snail)
	case TESTS:
		return fmt.Sprintf(`%s Tests`, emoji.TestTube)
	default:
		return ""
	}
}
