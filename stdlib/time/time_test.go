package time

import (
	"testing"
	gotime "time"
)

func TestNow(t *testing.T) {
	before := gotime.Now()
	now := Now()
	after := gotime.Now()

	if now.t.Before(before) || now.t.After(after) {
		t.Error("Now() should return current time")
	}
}

func TestUTC(t *testing.T) {
	utc := UTC()
	if utc.t.Location() != gotime.UTC {
		t.Error("UTC() should return UTC time")
	}
}

func TestUnix(t *testing.T) {
	ts := int64(1609459200) // 2021-01-01 00:00:00 UTC
	tm := Unix(ts)
	if tm.Unix() != ts {
		t.Errorf("Unix() = %d, want %d", tm.Unix(), ts)
	}
}

func TestUnixMilli(t *testing.T) {
	ms := int64(1609459200000)
	tm := UnixMilli(ms)
	if tm.UnixMilli() != ms {
		t.Errorf("UnixMilli() = %d, want %d", tm.UnixMilli(), ms)
	}
}

func TestParse(t *testing.T) {
	tm, err := Parse("2006-01-02", "2021-06-15")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if tm.Year() != 2021 || tm.Month() != 6 || tm.Day() != 15 {
		t.Errorf("Parse() = %d-%d-%d, want 2021-6-15", tm.Year(), tm.Month(), tm.Day())
	}
}

func TestParseError(t *testing.T) {
	_, err := Parse("2006-01-02", "invalid")
	if err == nil {
		t.Error("Parse() should return error for invalid input")
	}
}

func TestParseRFC3339(t *testing.T) {
	tm, err := ParseRFC3339("2021-06-15T10:30:00Z")
	if err != nil {
		t.Fatalf("ParseRFC3339() error = %v", err)
	}
	if tm.Year() != 2021 || tm.Hour() != 10 {
		t.Errorf("ParseRFC3339() year=%d hour=%d, want 2021, 10", tm.Year(), tm.Hour())
	}
}

func TestFormat(t *testing.T) {
	tm := Unix(1609459200).UTC()
	formatted := tm.Format("2006-01-02")
	if formatted != "2021-01-01" {
		t.Errorf("Format() = %q, want %q", formatted, "2021-01-01")
	}
}

func TestRFC3339(t *testing.T) {
	tm := Unix(1609459200).UTC()
	formatted := tm.RFC3339()
	if formatted != "2021-01-01T00:00:00Z" {
		t.Errorf("RFC3339() = %q, want %q", formatted, "2021-01-01T00:00:00Z")
	}
}

func TestTimeComponents(t *testing.T) {
	tm, _ := Parse("2006-01-02 15:04:05", "2021-06-15 10:30:45")

	if tm.Year() != 2021 {
		t.Errorf("Year() = %d, want 2021", tm.Year())
	}
	if tm.Month() != 6 {
		t.Errorf("Month() = %d, want 6", tm.Month())
	}
	if tm.Day() != 15 {
		t.Errorf("Day() = %d, want 15", tm.Day())
	}
	if tm.Hour() != 10 {
		t.Errorf("Hour() = %d, want 10", tm.Hour())
	}
	if tm.Minute() != 30 {
		t.Errorf("Minute() = %d, want 30", tm.Minute())
	}
	if tm.Second() != 45 {
		t.Errorf("Second() = %d, want 45", tm.Second())
	}
}

func TestWeekday(t *testing.T) {
	// 2021-06-15 is a Tuesday (weekday 2)
	tm, _ := Parse("2006-01-02", "2021-06-15")
	if tm.Weekday() != 2 {
		t.Errorf("Weekday() = %d, want 2 (Tuesday)", tm.Weekday())
	}
}

func TestAdd(t *testing.T) {
	tm := Unix(1609459200)
	added := tm.Add(Hours(24))
	if added.Unix() != 1609459200+86400 {
		t.Errorf("Add() = %d, want %d", added.Unix(), 1609459200+86400)
	}
}

func TestSub(t *testing.T) {
	t1 := Unix(1609459200)
	t2 := Unix(1609459200 + 3600)
	diff := t2.Sub(t1)
	if diff.ToSeconds() != 3600 {
		t.Errorf("Sub() = %f seconds, want 3600", diff.ToSeconds())
	}
}

func TestBeforeAfterEqual(t *testing.T) {
	t1 := Unix(1000)
	t2 := Unix(2000)
	t3 := Unix(1000)

	if !t1.Before(t2) {
		t.Error("t1.Before(t2) = false, want true")
	}
	if !t2.After(t1) {
		t.Error("t2.After(t1) = false, want true")
	}
	if !t1.Equal(t3) {
		t.Error("t1.Equal(t3) = false, want true")
	}
}

func TestIsZero(t *testing.T) {
	var zero Time
	if !zero.IsZero() {
		t.Error("zero.IsZero() = false, want true")
	}
	now := Now()
	if now.IsZero() {
		t.Error("now.IsZero() = true, want false")
	}
}

func TestUTCLocal(t *testing.T) {
	now := Now()
	utc := now.UTC()
	local := utc.Local()

	// Times should be equal even if in different timezones
	if !now.Equal(local) && !now.Equal(utc) {
		// This might fail depending on timezone, so just check they're close
		diff := now.Sub(local).Abs()
		if diff.ToSeconds() > 1 {
			t.Error("UTC/Local conversion changed time significantly")
		}
	}
}

func TestSinceUntil(t *testing.T) {
	past := Now().Add(Seconds(-10))
	sincePast := Since(past)
	if sincePast.ToSeconds() < 9 || sincePast.ToSeconds() > 12 {
		t.Errorf("Since() = %f, want ~10 seconds", sincePast.ToSeconds())
	}

	future := Now().Add(Seconds(10))
	untilFuture := Until(future)
	if untilFuture.ToSeconds() < 9 || untilFuture.ToSeconds() > 11 {
		t.Errorf("Until() = %f, want ~10 seconds", untilFuture.ToSeconds())
	}
}

func TestDurationCreators(t *testing.T) {
	if Nanoseconds(1000).d != 1000*gotime.Nanosecond {
		t.Error("Nanoseconds() incorrect")
	}
	if Microseconds(1000).d != 1000*gotime.Microsecond {
		t.Error("Microseconds() incorrect")
	}
	if Milliseconds(1000).d != 1000*gotime.Millisecond {
		t.Error("Milliseconds() incorrect")
	}
	if Seconds(60).d != 60*gotime.Second {
		t.Error("Seconds() incorrect")
	}
	if Minutes(60).d != 60*gotime.Minute {
		t.Error("Minutes() incorrect")
	}
	if Hours(24).d != 24*gotime.Hour {
		t.Error("Hours() incorrect")
	}
	if Days(7).d != 7*24*gotime.Hour {
		t.Error("Days() incorrect")
	}
}

func TestDurationConversions(t *testing.T) {
	d := Hours(2)

	if d.ToHours() != 2 {
		t.Errorf("ToHours() = %f, want 2", d.ToHours())
	}
	if d.ToMinutes() != 120 {
		t.Errorf("ToMinutes() = %f, want 120", d.ToMinutes())
	}
	if d.ToSeconds() != 7200 {
		t.Errorf("ToSeconds() = %f, want 7200", d.ToSeconds())
	}
	if d.ToMilliseconds() != 7200000 {
		t.Errorf("ToMilliseconds() = %d, want 7200000", d.ToMilliseconds())
	}
}

func TestDurationString(t *testing.T) {
	d := Hours(1).Add(Minutes(30))
	s := d.String()
	if s != "1h30m0s" {
		t.Errorf("String() = %q, want %q", s, "1h30m0s")
	}
}

func TestDurationArithmetic(t *testing.T) {
	d1 := Hours(1)
	d2 := Minutes(30)

	sum := d1.Add(d2)
	if sum.ToMinutes() != 90 {
		t.Errorf("Add() = %f minutes, want 90", sum.ToMinutes())
	}

	diff := d1.Sub(d2)
	if diff.ToMinutes() != 30 {
		t.Errorf("Sub() = %f minutes, want 30", diff.ToMinutes())
	}

	mul := d2.Mul(4)
	if mul.ToHours() != 2 {
		t.Errorf("Mul() = %f hours, want 2", mul.ToHours())
	}
}

func TestDurationAbs(t *testing.T) {
	d := Seconds(-10)
	abs := d.Abs()
	if abs.ToSeconds() != 10 {
		t.Errorf("Abs() = %f, want 10", abs.ToSeconds())
	}
}

func TestLayoutConstants(t *testing.T) {
	tm := Unix(1609459200).UTC()

	if tm.Format(DateLayout) != "2021-01-01" {
		t.Errorf("DateLayout format = %q, want %q", tm.Format(DateLayout), "2021-01-01")
	}
	if tm.Format(TimeLayout) != "00:00:00" {
		t.Errorf("TimeLayout format = %q, want %q", tm.Format(TimeLayout), "00:00:00")
	}
	if tm.Format(DateTimeLayout) != "2021-01-01 00:00:00" {
		t.Errorf("DateTimeLayout format = %q, want %q", tm.Format(DateTimeLayout), "2021-01-01 00:00:00")
	}
}

func TestSleep(t *testing.T) {
	start := Now()
	Sleep(Milliseconds(50))
	elapsed := Since(start)
	if elapsed.ToMilliseconds() < 40 {
		t.Errorf("Sleep(50ms) elapsed = %dms, want >= 40ms", elapsed.ToMilliseconds())
	}
}

func TestTimeString(t *testing.T) {
	tm := Unix(1609459200).UTC()
	s := tm.String()
	if s != "2021-01-01T00:00:00Z" {
		t.Errorf("String() = %q, want %q", s, "2021-01-01T00:00:00Z")
	}
}
