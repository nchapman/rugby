// Package time provides time and duration operations for Rugby programs.
// Rugby: import rugby/time
//
// Example:
//
//	now = time.now()
//	tomorrow = now.add(time.hours(24))
//	puts now.format("2006-01-02")
package time

import (
	"time"
)

// Time wraps Go's time.Time with a Ruby-like API.
type Time struct {
	t time.Time
}

// Duration wraps Go's time.Duration.
type Duration struct {
	d time.Duration
}

// Now returns the current local time.
// Ruby: time.now
func Now() Time {
	return Time{t: time.Now()}
}

// UTC returns the current UTC time.
// Ruby: time.utc
func UTC() Time {
	return Time{t: time.Now().UTC()}
}

// Unix creates a Time from Unix seconds.
// Ruby: time.unix(seconds)
func Unix(sec int64) Time {
	return Time{t: time.Unix(sec, 0)}
}

// UnixMilli creates a Time from Unix milliseconds.
// Ruby: time.unix_milli(ms)
func UnixMilli(ms int64) Time {
	return Time{t: time.UnixMilli(ms)}
}

// Parse parses a time string using the given layout.
// See Go's time package for layout format (e.g., "2006-01-02 15:04:05").
// Ruby: time.parse(layout, str)
func Parse(layout, value string) (Time, error) {
	t, err := time.Parse(layout, value)
	if err != nil {
		return Time{}, err
	}
	return Time{t: t}, nil
}

// ParseRFC3339 parses an RFC3339 formatted time string.
// Ruby: time.parse_rfc3339(str)
func ParseRFC3339(value string) (Time, error) {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return Time{}, err
	}
	return Time{t: t}, nil
}

// Format formats the time using the given layout.
// Ruby: t.format(layout)
func (t Time) Format(layout string) string {
	return t.t.Format(layout)
}

// RFC3339 formats the time as RFC3339.
// Ruby: t.rfc3339
func (t Time) RFC3339() string {
	return t.t.Format(time.RFC3339)
}

// Unix returns the Unix timestamp in seconds.
// Ruby: t.unix
func (t Time) Unix() int64 {
	return t.t.Unix()
}

// UnixMilli returns the Unix timestamp in milliseconds.
// Ruby: t.unix_milli
func (t Time) UnixMilli() int64 {
	return t.t.UnixMilli()
}

// Year returns the year.
// Ruby: t.year
func (t Time) Year() int {
	return t.t.Year()
}

// Month returns the month (1-12).
// Ruby: t.month
func (t Time) Month() int {
	return int(t.t.Month())
}

// Day returns the day of the month.
// Ruby: t.day
func (t Time) Day() int {
	return t.t.Day()
}

// Hour returns the hour (0-23).
// Ruby: t.hour
func (t Time) Hour() int {
	return t.t.Hour()
}

// Minute returns the minute (0-59).
// Ruby: t.minute
func (t Time) Minute() int {
	return t.t.Minute()
}

// Second returns the second (0-59).
// Ruby: t.second
func (t Time) Second() int {
	return t.t.Second()
}

// Weekday returns the day of the week (0=Sunday, 6=Saturday).
// Ruby: t.weekday
func (t Time) Weekday() int {
	return int(t.t.Weekday())
}

// Add returns the time plus the duration.
// Ruby: t.add(duration)
func (t Time) Add(d Duration) Time {
	return Time{t: t.t.Add(d.d)}
}

// Sub returns the duration between two times.
// Ruby: t.sub(other)
func (t Time) Sub(other Time) Duration {
	return Duration{d: t.t.Sub(other.t)}
}

// Before reports whether t is before other.
// Ruby: t.before?(other)
func (t Time) Before(other Time) bool {
	return t.t.Before(other.t)
}

// After reports whether t is after other.
// Ruby: t.after?(other)
func (t Time) After(other Time) bool {
	return t.t.After(other.t)
}

// Equal reports whether t equals other.
// Ruby: t.equal?(other)
func (t Time) Equal(other Time) bool {
	return t.t.Equal(other.t)
}

// IsZero reports whether t is the zero time.
// Ruby: t.zero?
func (t Time) IsZero() bool {
	return t.t.IsZero()
}

// String returns the time as a string in RFC3339 format.
// Ruby: t.to_s
func (t Time) String() string {
	return t.t.Format(time.RFC3339)
}

// UTC converts the time to UTC.
// Ruby: t.utc
func (t Time) UTC() Time {
	return Time{t: t.t.UTC()}
}

// Local converts the time to local timezone.
// Ruby: t.local
func (t Time) Local() Time {
	return Time{t: t.t.Local()}
}

// Since returns the duration since t.
// Ruby: time.since(t)
func Since(t Time) Duration {
	return Duration{d: time.Since(t.t)}
}

// Until returns the duration until t.
// Ruby: time.until(t)
func Until(t Time) Duration {
	return Duration{d: time.Until(t.t)}
}

// Nanoseconds creates a duration of n nanoseconds.
// Ruby: time.nanoseconds(n)
func Nanoseconds(n int64) Duration {
	return Duration{d: time.Duration(n) * time.Nanosecond}
}

// Microseconds creates a duration of n microseconds.
// Ruby: time.microseconds(n)
func Microseconds(n int64) Duration {
	return Duration{d: time.Duration(n) * time.Microsecond}
}

// Milliseconds creates a duration of n milliseconds.
// Ruby: time.milliseconds(n)
func Milliseconds(n int64) Duration {
	return Duration{d: time.Duration(n) * time.Millisecond}
}

// Seconds creates a duration of n seconds.
// Ruby: time.seconds(n)
func Seconds(n int64) Duration {
	return Duration{d: time.Duration(n) * time.Second}
}

// Minutes creates a duration of n minutes.
// Ruby: time.minutes(n)
func Minutes(n int64) Duration {
	return Duration{d: time.Duration(n) * time.Minute}
}

// Hours creates a duration of n hours.
// Ruby: time.hours(n)
func Hours(n int64) Duration {
	return Duration{d: time.Duration(n) * time.Hour}
}

// Days creates a duration of n days.
// Ruby: time.days(n)
func Days(n int64) Duration {
	return Duration{d: time.Duration(n) * 24 * time.Hour}
}

// ToSeconds returns the duration in seconds.
// Ruby: d.seconds
func (d Duration) ToSeconds() float64 {
	return d.d.Seconds()
}

// ToMilliseconds returns the duration in milliseconds.
// Ruby: d.milliseconds
func (d Duration) ToMilliseconds() int64 {
	return d.d.Milliseconds()
}

// ToMinutes returns the duration in minutes.
// Ruby: d.minutes
func (d Duration) ToMinutes() float64 {
	return d.d.Minutes()
}

// ToHours returns the duration in hours.
// Ruby: d.hours
func (d Duration) ToHours() float64 {
	return d.d.Hours()
}

// String returns the duration as a string.
// Ruby: d.to_s
func (d Duration) String() string {
	return d.d.String()
}

// Add returns d + other.
// Ruby: d.add(other)
func (d Duration) Add(other Duration) Duration {
	return Duration{d: d.d + other.d}
}

// Sub returns d - other.
// Ruby: d.sub(other)
func (d Duration) Sub(other Duration) Duration {
	return Duration{d: d.d - other.d}
}

// Mul returns d * n.
// Ruby: d.mul(n)
func (d Duration) Mul(n int64) Duration {
	return Duration{d: d.d * time.Duration(n)}
}

// Abs returns the absolute value of the duration.
// Ruby: d.abs
func (d Duration) Abs() Duration {
	if d.d < 0 {
		return Duration{d: -d.d}
	}
	return d
}

// Sleep pauses the current goroutine for the duration.
// Ruby: time.sleep(duration)
func Sleep(d Duration) {
	time.Sleep(d.d)
}

// SleepSeconds pauses for n seconds.
// Ruby: time.sleep_seconds(n)
func SleepSeconds(n float64) {
	time.Sleep(time.Duration(n * float64(time.Second)))
}

// Common layout constants for convenience.
const (
	RFC3339Layout  = time.RFC3339
	RFC822Layout   = time.RFC822
	DateLayout     = "2006-01-02"
	TimeLayout     = "15:04:05"
	DateTimeLayout = "2006-01-02 15:04:05"
)
