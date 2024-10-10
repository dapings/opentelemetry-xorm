package now

import (
	"testing"
	"time"
)

var (
	format = "2006-01-02 15:04:05.999999999"
)

func assertT(t *testing.T) func(time.Time, string, string, bool) {
	return func(actual time.Time, expected string, msg string, formatted bool) {
		var actualStr string
		if formatted {
			actualStr = actual.Format(format)
		} else {
			actualStr = actual.String()
		}

		if actualStr != expected {
			t.Errorf("failed: %s: got: %v, expected: %v", msg, actualStr, expected)
		}
	}
}

func TestConfig(t *testing.T) {
	assert := assertT(t)

	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Errorf("load location for Asia/Shanghai, should return no error, but got %v", err)
	}

	conf := Config{
		WeekStartDay: time.Monday,
		TimeLocation: location,
		TimeFormats:  []string{"2006-01-02 15:04:05"},
	}

	result, err := conf.Parse("2024-09-24 15:17:06")
	msg := "parse in location"
	if err != nil {
		msg += err.Error()
	}
	assert(result, "2024-09-24 15:17:06 +0800 CST", msg, false)

	result = conf.MustParse("2024-09-24 15:17:06")
	assert(result, "2024-09-24 15:17:06 +0800 CST", "parse in location", false)
}

func TestParse(t *testing.T) {
	assert := assertT(t)

	n := time.Date(2024, 9, 24, 17, 51, 49, 123456789, time.UTC)

	assert(With(n).MustParse("2002"), "2002-01-01 00:00:00", "parse 2002", true)

	assert(With(n).MustParse("2002-10"), "2002-10-01 00:00:00", "Parse 2002-10", true)

	assert(With(n).MustParse("2002-10-12"), "2002-10-12 00:00:00", "Parse 2002-10-12", true)

	assert(With(n).MustParse("2002-10-12 22"), "2002-10-12 22:00:00", "Parse 2002-10-12 22", true)

	assert(With(n).MustParse("2002-10-12 22:14"), "2002-10-12 22:14:00", "Parse 2002-10-12 22:14", true)

	assert(With(n).MustParse("2002-10-12 2:4"), "2002-10-12 02:04:00", "Parse 2002-10-12 2:4", true)

	assert(With(n).MustParse("2002-10-12 02:04"), "2002-10-12 02:04:00", "Parse 2002-10-12 02:04", true)

	assert(With(n).MustParse("2002-10-12 22:14:56"), "2002-10-12 22:14:56", "Parse 2002-10-12 22:14:56", true)

	assert(With(n).MustParse("2002-10-12 00:14:56"), "2002-10-12 00:14:56", "Parse 2002-10-12 00:14:56", true)

	assert(With(n).MustParse("2013-12-19 23:28:09.999999999 +0800 CST"), "2013-12-19 23:28:09.999999999", "Parse two strings 2013-12-19 23:28:09.999999999 +0800 CST", true)

	assert(With(n).MustParse("10-12"), "2024-10-12 00:00:00", "Parse 10-12", true)

	assert(With(n).MustParse("18"), "2024-09-24 18:00:00", "Parse 18 as hour", true)

	assert(With(n).MustParse("18:20"), "2024-09-24 18:20:00", "Parse 18:20", true)

	assert(With(n).MustParse("00:01"), "2024-09-24 00:01:00", "Parse 00:01", true)

	assert(With(n).MustParse("00:00:00"), "2024-09-24 00:00:00", "Parse 00:00:00", true)

	assert(With(n).MustParse("18:20:39"), "2024-09-24 18:20:39", "Parse 18:20:39", true)

	assert(With(n).MustParse("18:20:39", "2011-01-01"), "2011-01-01 18:20:39", "Parse two strings 18:20:39, 2011-01-01", true)

	assert(With(n).MustParse("2011-1-1", "18:20:39"), "2011-01-01 18:20:39", "Parse two strings 2011-01-01, 18:20:39", true)

	assert(With(n).MustParse("2011-01-01", "18"), "2011-01-01 18:00:00", "Parse two strings 2011-01-01, 18", true)

	assert(With(n).MustParse("2002-10-12T00:14:56Z"), "2002-10-12 00:14:56", "Parse 2002-10-12T00:14:56Z", true)

	assert(With(n).MustParse("2002-10-12T00:00:56Z"), "2002-10-12 00:00:56", "Parse 2002-10-12T00:00:56Z", true)

	assert(With(n).MustParse("2002-10-12T00:00:00.999Z"), "2002-10-12 00:00:00.999", "Parse 2002-10-12T00:00:00.999Z", true)

	assert(With(n).MustParse("2002-10-12T00:14:56.999999Z"), "2002-10-12 00:14:56.999999", "Parse 2002-10-12T00:14:56.999999Z", true)

	assert(With(n).MustParse("2002-10-12T00:00:56.999999999Z"), "2002-10-12 00:00:56.999999999", "Parse 2002-10-12T00:00:56.999999999Z", true)

	assert(With(n).MustParse("2002-10-12T00:14:56+08:00"), "2002-10-12 00:14:56", "Parse 2002-10-12T00:14:56+08:00", true)

	_, off := With(n).MustParse("2002-10-12T00:14:56+08:00").Zone()
	if off != 28800 {
		t.Errorf("Parse 2002-10-12T00:14:56+08:00 shouldn't lose time zone offset")
	}

	assert(With(n).MustParse("2002-10-12T00:00:56-07:00"), "2002-10-12 00:00:56", "Parse 2002-10-12T00:00:56-07:00", true)

	_, off2 := With(n).MustParse("2002-10-12T00:00:56-07:00").Zone()
	if off2 != -25200 {
		t.Errorf("Parse 2002-10-12T00:00:56-07:00 shouldn't lose time zone offset")
	}

	assert(With(n).MustParse("2002-10-12T00:01:12.333+0200"), "2002-10-12 00:01:12.333", "Parse 2002-10-12T00:01:12.333+0200", true)

	_, off3 := With(n).MustParse("2002-10-12T00:01:12.333+0200").Zone()
	if off3 != 7200 {
		t.Errorf("Parse 2002-10-12T00:01:12.333+0200 shouldn't lose time zone offset")
	}

	assert(With(n).MustParse("2002-10-12T00:00:56.999999999+08:00"), "2002-10-12 00:00:56.999999999", "Parse 2002-10-12T00:00:56.999999999+08:00", true)

	_, off4 := With(n).MustParse("2002-10-12T00:14:56.999999999+08:00").Zone()
	if off4 != 28800 {
		t.Errorf("Parse 2002-10-12T00:14:56.999999999+08:00 shouldn't lose time zone offset")
	}

	assert(With(n).MustParse("2002-10-12T00:00:56.666666-07:00"), "2002-10-12 00:00:56.666666", "Parse 2002-10-12T00:00:56.666666-07:00", true)

	_, off5 := With(n).MustParse("2002-10-12T00:00:56.666666-07:00").Zone()
	if off5 != -25200 {
		t.Errorf("Parse 2002-10-12T00:00:56.666666-07:00 shouldn't lose time zone offset")
	}

	assert(With(n).MustParse("2002-10-12T00:01:12.999999999-06"), "2002-10-12 00:01:12.999999999", "Parse 2002-10-12T00:01:12.999999999-06", true)

	_, off6 := With(n).MustParse("2002-10-12T00:01:12.999999999-06").Zone()
	if off6 != -21600 {
		t.Errorf("Parse 2002-10-12T00:01:12.999999999-06 shouldn't lose time zone offset")
	}

	TimeFormats = append(TimeFormats, "02 Jan 15:04")
	assert(With(n).MustParse("04 Feb 12:09"), "2024-02-04 12:09:00", "Parse 04 Feb 12:09 with specified format", true)

	assert(With(n).MustParse("23:28:9 Dec 19, 2013 PST"), "2013-12-19 23:28:09", "Parse 23:28:9 Dec 19, 2013 PST", true)

	if With(n).MustParse("23:28:9 Dec 19, 2013 PST").Location().String() != "PST" {
		t.Errorf("Parse 23:28:9 Dec 19, 2013 PST shouldn't lose time zone")
	}

	n2 := With(n).MustParse("23:28:9 Dec 19, 2013 PST")
	if With(n2).MustParse("10:20").Location().String() != "PST" {
		t.Errorf("Parse 10:20 shouldn't change time zone")
	}

	TimeFormats = append(TimeFormats, "2006-01-02T15:04:05.0")
	TimeFormats = append(TimeFormats, "2006-01-02 15:04:05.000")
	assert(With(n).MustParse("2018-04-20 21:22:23.473"), "2018-04-20 21:22:23.473", "Parse 2018/04/20 21:22:23.473", true)

	TimeFormats = append(TimeFormats, "15:04:05.000")
	assert(With(n).MustParse("13:00:01.365"), "2024-09-24 13:00:01.365", "Parse 13:00:01.365", true)

	TimeFormats = append(TimeFormats, "2006-01-02 15:04:05.000000")
	assert(With(n).MustParse("2010-01-01 07:24:23.131384"), "2010-01-01 07:24:23.131384", "Parse 2010-01-01 07:24:23.131384", true)
	assert(With(n).MustParse("00:00:00.182736"), "2024-09-24 00:00:00.182736", "Parse 00:00:00.182736", true)

	n3 := MustParse("2017-12-11T10:25:49Z")
	if n3.Location() != time.UTC {
		t.Errorf("time location should be UTC, but got %v", n3.Location())
	}
}
