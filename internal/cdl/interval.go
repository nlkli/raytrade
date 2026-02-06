package cdl

import (
	"fmt"
	"strings"
)

type Interval uint16

const (
	M1  Interval = 1
	M3  Interval = 3
	M5  Interval = 5
	M15 Interval = 15
	M30 Interval = 30
	H1  Interval = 60
	H2  Interval = 120
	H4  Interval = 240
	H6  Interval = 360
	H12 Interval = 720
	D1  Interval = 1440
	D7  Interval = 10080
	D30 Interval = 43200
)

func (i Interval) AsSeconds() int {
	return int(i) * 60
}

func (i Interval) AsMilli() int {
	return i.AsSeconds() * 1000
}

func (i Interval) AsString() string {
	switch i {
	case M1:
		return "M1"
	case M3:
		return "M3"
	case M5:
		return "M5"
	case M15:
		return "M15"
	case M30:
		return "M30"
	case H1:
		return "H1"
	case H2:
		return "H2"
	case H4:
		return "H4"
	case H6:
		return "H6"
	case H12:
		return "H12"
	case D1:
		return "D1"
	case D7:
		return "D7"
	case D30:
		return "D30"
	default:
		return ""
	}
}

func IntervalFromString(s string) (Interval, error) {
	s = strings.ToUpper(strings.TrimSpace(s))

	switch s {
	// Minute intervals
	case "M1", "1":
		return M1, nil
	case "M3", "3":
		return M3, nil
	case "M5", "5":
		return M5, nil
	case "M15", "15":
		return M15, nil
	case "M30", "30":
		return M30, nil
	// Hour intervals
	case "H1", "60":
		return H1, nil
	case "H2", "120":
		return H2, nil
	case "H4", "240":
		return H4, nil
	case "H6", "360":
		return H6, nil
	case "H12", "720":
		return H12, nil
	// Day intervals
	case "D1", "1440", "D":
		return D1, nil
	case "D7", "10080", "W":
		return D7, nil
	case "D30", "43200", "M":
		return D30, nil
	default:
		return 0, fmt.Errorf("invalid interval: %s", s)
	}
}
