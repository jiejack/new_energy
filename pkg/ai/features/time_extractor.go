package features

import (
	"time"
)

type SimpleTimeExtractor struct {
	holidays map[string]bool
}

func NewSimpleTimeExtractor() *SimpleTimeExtractor {
	return &SimpleTimeExtractor{
		holidays: make(map[string]bool),
	}
}

func (e *SimpleTimeExtractor) Extract(t time.Time) TimeFeatures {
	features := TimeFeatures{}
	
	features.Hour = t.Hour()
	features.Day = t.Day()
	features.Weekday = int(t.Weekday())
	if features.Weekday == 0 {
		features.Weekday = 7
	}
	features.Month = int(t.Month())
	features.Year = t.Year()
	features.DayOfYear = t.YearDay()
	
	_, week := t.ISOWeek()
	features.WeekOfYear = week
	
	features.IsWeekend = t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
	
	features.Season = getSeason(t.Month())
	
	features.Quarter = (int(t.Month())-1)/3 + 1
	
	dateKey := t.Format("2006-01-02")
	features.IsHoliday = e.holidays[dateKey]
	
	return features
}

func (e *SimpleTimeExtractor) AddHoliday(date string) {
	e.holidays[date] = true
}

func (e *SimpleTimeExtractor) RemoveHoliday(date string) {
	delete(e.holidays, date)
}

func getSeason(month time.Month) int {
	switch month {
	case time.December, time.January, time.February:
		return 1
	case time.March, time.April, time.May:
		return 2
	case time.June, time.July, time.August:
		return 3
	case time.September, time.October, time.November:
		return 4
	default:
		return 0
	}
}

type ConfigurableTimeExtractor struct {
	config   TimeFeatureConfig
	extractor *SimpleTimeExtractor
}

func NewConfigurableTimeExtractor(config TimeFeatureConfig) *ConfigurableTimeExtractor {
	return &ConfigurableTimeExtractor{
		config:    config,
		extractor: NewSimpleTimeExtractor(),
	}
}

func (e *ConfigurableTimeExtractor) Extract(t time.Time) TimeFeatures {
	fullFeatures := e.extractor.Extract(t)
	
	if !e.config.Enabled {
		return TimeFeatures{}
	}
	
	features := TimeFeatures{}
	
	if e.config.Hour {
		features.Hour = fullFeatures.Hour
	}
	if e.config.Day {
		features.Day = fullFeatures.Day
	}
	if e.config.Weekday {
		features.Weekday = fullFeatures.Weekday
	}
	if e.config.Month {
		features.Month = fullFeatures.Month
	}
	if e.config.Quarter {
		features.Quarter = fullFeatures.Quarter
	}
	if e.config.Year {
		features.Year = fullFeatures.Year
	}
	if e.config.DayOfYear {
		features.DayOfYear = fullFeatures.DayOfYear
	}
	if e.config.WeekOfYear {
		features.WeekOfYear = fullFeatures.WeekOfYear
	}
	if e.config.IsWeekend {
		features.IsWeekend = fullFeatures.IsWeekend
	}
	if e.config.Season {
		features.Season = fullFeatures.Season
	}
	
	features.IsHoliday = fullFeatures.IsHoliday
	
	return features
}

func (e *ConfigurableTimeExtractor) AddHoliday(date string) {
	e.extractor.AddHoliday(date)
}
