// Copyright 2020-2026 Politecnico di Torino
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"
)

// durationWithDaysRegex is the format read by ParseDurationWithDays, the same as in pkg/instautoctrl.
var durationWithDaysRegex = regexp.MustCompile(`^(\d+)([mhd])$`)

var durationWithDaysUnits = map[string]time.Duration{
	"m": time.Minute,
	"h": time.Hour,
	"d": 24 * time.Hour,
}

// ParseDurationWithDays parses a duration made of a number and a unit among m, h and d
// (e.g., 30m, 6h, 2d), with the same rules as instautoctrl.ParseDurationWithDays.
// Unlike that function, it refuses zero and the values that do not fit in a time.Duration.
func ParseDurationWithDays(input string) (time.Duration, error) {
	matches := durationWithDaysRegex.FindStringSubmatch(input)
	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid duration %q: expected a number followed by m, h or d", input)
	}

	unit := durationWithDaysUnits[matches[2]]
	value, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil || value > math.MaxInt64/int64(unit) {
		return 0, fmt.Errorf("invalid duration %q: the value is too large", input)
	}
	if value == 0 {
		return 0, fmt.Errorf("invalid duration %q: the value must be positive", input)
	}
	return time.Duration(value) * unit, nil
}

// FormatDurationWithDays formats a duration in the format read by ParseDurationWithDays,
// with the largest unit that divides it exactly (e.g., 2d, 6h, 90m).
// A duration that is not a whole number of minutes is rounded up to the minute.
func FormatDurationWithDays(d time.Duration) string {
	minutes := int64((d + time.Minute - 1) / time.Minute)
	switch {
	case minutes > 0 && minutes%(24*60) == 0:
		return strconv.FormatInt(minutes/(24*60), 10) + "d"
	case minutes > 0 && minutes%60 == 0:
		return strconv.FormatInt(minutes/60, 10) + "h"
	default:
		return strconv.FormatInt(minutes, 10) + "m"
	}
}
