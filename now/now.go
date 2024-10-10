package now

import (
	"errors"
	"regexp"
	"time"
)

// WeekStartDay set week start day, default is sunday.
var WeekStartDay = time.Sunday

// TimeFormats default time formats will be parsed.
var TimeFormats = []string{
	"2006", "2006-1", "2006-1-2", "2006-1-2 15", "2006-1-2 15:4", "2006-1-2 15:4:5", "1-2",
	"15:4:5", "15:4", "15",
	"15:4:5 Jan 2, 2006 MST", "2006-01-02 15:04:05.999999999 -0700 MST", "2006-01-02T15:04:05Z0700", "2006-01-02T15:04:05Z07",
	"2006.1.2", "2006.1.2 15:04:05", "2006.01.02", "2006.01.02 15:04:05", "2006.01.02 15:04:05.999999999",
	"1/2/2006", "1/2/2006 15:4:5", "2006/01/02", "20060102", "2006/01/02 15:04:05",
	time.ANSIC, time.UnixDate, time.RubyDate, time.RFC822, time.RFC822Z, time.RFC850,
	time.RFC1123, time.RFC1123Z, time.RFC3339, time.RFC3339Nano,
	time.Kitchen, time.Stamp, time.StampMilli, time.StampMicro, time.StampNano,
}

var hasTimeRegexp = regexp.MustCompile(`(\s+|^\s*|T)\d{1,2}((:\d{1,2})*|((:\d{1,2}){2}\.(\d{3}|\d{6}|\d{9})))(\s*$|[Z+-])`) // match 15:04:05, 15:04:05.000, 15:04:05.000000 15, 2024-01-01 15:04, 2024-09-25T00:59:10Z, 2024-09-25T00:59:10+08:00, 2024-09-25T00:00:10-07:00 etc
var onlyTimeRegexp = regexp.MustCompile(`^\s*\d{1,2}((:\d{1,2})*|((:\d{1,2}){2}\.(\d{3}|\d{6}|\d{9})))\s*$`)                // match 15:04:05, 15, 15:04:05.000, 15:04:05.000000, etc

func formatTimeToList(t time.Time) []int {
	hour, min, sec := t.Clock()
	year, month, day := t.Date()
	return []int{t.Nanosecond(), sec, min, hour, day, int(month), year}
}

// Config configuration for now package.
type Config struct {
	WeekStartDay time.Weekday
	TimeLocation *time.Location
	TimeFormats  []string
}

// DefaultConfig default config.
var DefaultConfig *Config

// Parse parses string to time based on configuration.
func (conf *Config) Parse(strs ...string) (time.Time, error) {
	if conf.TimeLocation == nil {
		return conf.With(time.Now()).Parse(strs...)
	}

	return conf.With(time.Now().In(conf.TimeLocation)).Parse(strs...)
}

// MustParse must parse string to time, or will panic.
func (conf *Config) MustParse(strs ...string) time.Time {
	if conf.TimeLocation == nil {
		return conf.With(time.Now()).MustParse(strs...)
	}

	return conf.With(time.Now().In(conf.TimeLocation)).MustParse(strs...)
}

// With initialize Now based on configuration.
func (conf *Config) With(t time.Time) *Now {
	return &Now{Time: t, Config: conf}
}

// Now a now struct.
type Now struct {
	time.Time
	*Config
}

func (now *Now) parseWithFormat(str string, location *time.Location) (t time.Time, err error) {
	for _, format := range TimeFormats {
		t, err = time.ParseInLocation(format, str, location)
		if err == nil {
			return
		}
	}
	err = errors.New("can't parse string as time: " + str)
	return
}

// Parse parses string to time.
func (now *Now) Parse(strs ...string) (t time.Time, err error) {
	var (
		setCurrentTime  bool
		parseTime       []int
		currentLocation = now.Location()
		onlyTimeInStr   = true
		currentTime     = formatTimeToList(now.Time)
	)

	for _, str := range strs {
		hasTimeInStr := hasTimeRegexp.MatchString(str) // match 15:04:05, 15
		onlyTimeInStr = hasTimeInStr && onlyTimeInStr && onlyTimeRegexp.MatchString(str)
		if t, err = now.parseWithFormat(str, currentLocation); err == nil {
			location := t.Location()
			parseTime = formatTimeToList(t)

			for i, val := range parseTime {
				// don't reset hour,min,sec if current time str including time.
				if hasTimeInStr && i <= 3 {
					continue
				}

				// if value is zero, replace it with current time.
				if val == 0 {
					if setCurrentTime {
						parseTime[i] = currentTime[i]
					}
				} else {
					setCurrentTime = true
				}

				// if current time only includes time, should change day,month to current time.
				if onlyTimeInStr {
					if i == 4 || i == 5 {
						parseTime[i] = currentTime[i]
						continue
					}
				}
			}

			t = time.Date(parseTime[6], time.Month(parseTime[5]), parseTime[4], parseTime[3], parseTime[2], parseTime[1], parseTime[0], location)
			currentTime = formatTimeToList(t)
		}
	}

	return
}

// MustParse must parse string to time, or it will panic.
func (now *Now) MustParse(strs ...string) time.Time {
	t, err := now.Parse(strs...)
	if err != nil {
		panic(err)
	}
	return t
}

// With initialize Now with time.
func With(t time.Time) *Now {
	conf := DefaultConfig
	if conf == nil {
		conf = &Config{
			WeekStartDay: WeekStartDay,
			TimeFormats:  TimeFormats,
		}
	}

	return conf.With(t)
}

// MustParse must parse string to time, or will panic.
func MustParse(strs ...string) time.Time {
	return With(time.Now()).MustParse(strs...)
}
