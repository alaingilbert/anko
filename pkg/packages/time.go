package packages

import (
	"time"
)

func init() {
	Packages["time"] = map[string]any{
		"ANSIC":           time.ANSIC,
		"After":           time.After,
		"AfterFunc":       time.AfterFunc,
		"April":           time.April,
		"August":          time.August,
		"Date":            time.Date,
		"December":        time.December,
		"February":        time.February,
		"FixedZone":       time.FixedZone,
		"Friday":          time.Friday,
		"Hour":            time.Hour,
		"January":         time.January,
		"July":            time.July,
		"June":            time.June,
		"Kitchen":         time.Kitchen,
		"LoadLocation":    time.LoadLocation,
		"March":           time.March,
		"May":             time.May,
		"Microsecond":     time.Microsecond,
		"Millisecond":     time.Millisecond,
		"Minute":          time.Minute,
		"Monday":          time.Monday,
		"Nanosecond":      time.Nanosecond,
		"NewTicker":       time.NewTicker,
		"NewTimer":        time.NewTimer,
		"November":        time.November,
		"Now":             time.Now,
		"October":         time.October,
		"Parse":           time.Parse,
		"ParseDuration":   time.ParseDuration,
		"ParseInLocation": time.ParseInLocation,
		"RFC1123":         time.RFC1123,
		"RFC1123Z":        time.RFC1123Z,
		"RFC3339":         time.RFC3339,
		"RFC3339Nano":     time.RFC3339Nano,
		"RFC822":          time.RFC822,
		"RFC822Z":         time.RFC822Z,
		"RFC850":          time.RFC850,
		"RubyDate":        time.RubyDate,
		"Saturday":        time.Saturday,
		"Second":          time.Second,
		"September":       time.September,
		"Since":           time.Since,
		"Sleep":           time.Sleep,
		"Stamp":           time.Stamp,
		"StampMicro":      time.StampMicro,
		"StampMilli":      time.StampMilli,
		"StampNano":       time.StampNano,
		"Sunday":          time.Sunday,
		"Thursday":        time.Thursday,
		"Tick":            time.Tick,
		"Tuesday":         time.Tuesday,
		"Unix":            time.Unix,
		"UnixDate":        time.UnixDate,
		"Wednesday":       time.Wednesday,
	}
	PackageTypes["time"] = map[string]any{
		"Duration": time.Duration(0),
		"Ticker":   time.Ticker{},
		"Time":     time.Time{},
	}
	timeGo18()
	timeGo110()
}
