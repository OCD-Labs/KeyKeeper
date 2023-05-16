package types

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Interval time.Duration

var ErrInvalidIntervalFormat = errors.New("invalid interval format")
var intervalFormats = [6]string{"seconds", "minutes", "hours", "days", "months", "years"}

// UnmarshalJSON satisfies the json.UnmarshalJSON interface
func (i *Interval) UnmarshalJSON(json []byte) error {
	val, err := strconv.Unquote(string(json))
	if err != nil {
		return ErrInvalidIntervalFormat
	}

	parts := strings.Split(val, " ")
	if len(parts) != 2 || !containsItem(intervalFormats, parts[1]) {
		return ErrInvalidIntervalFormat
	}

	interval, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidIntervalFormat
	}

	switch parts[1] {
	case "seconds":
		fmt.Println(Interval(time.Duration(interval) * time.Second))
		*i = Interval(time.Duration(interval) * time.Second)
	case "minutes":
		*i = Interval(time.Duration(interval) * time.Minute)
	case "hours":
		*i = Interval(time.Duration(interval) * time.Hour)
	case "days":
		*i = Interval(time.Duration(interval) * time.Hour * 24)
	case "months":
		*i = Interval(time.Duration(interval) * time.Hour * 24 * 30)
	case "years":
		*i = Interval(time.Duration(interval) * time.Hour * 24 * 365)
	}

	return nil
}

func containsItem(array [6]string, item string) bool {
	for _, el := range array {
		if el == item {
			return true
		}
	}

	return false
}
