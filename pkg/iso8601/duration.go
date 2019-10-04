package iso8601

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	mMinute = 60
	mHour   = 60 * mMinute
	mDay    = 24 * mHour
	mMonth  = 30 * mDay
	mYear   = 365 * mDay

	second = 1000
	minute = 1000 * mMinute
	hour   = 1000 * mHour
	day    = 1000 * mDay
)

// ISODuration represents a parsed ISO 6801 duration.
type ISODuration struct {
	MYears  uint
	MMonths uint

	MDays    uint
	MHours   uint
	MMinutes uint
	MSeconds uint
}

func fmtFrac(v uint, prec int) (uint, []byte) {
	buf := make([]byte, prec+1)
	wr := len(buf)

	include := false
	for i := 0; i < prec; i++ {
		digit := v % 10

		// once we've had a non-zero value, force print!
		include = include || digit != 0

		if include {
			wr--
			buf[wr] = '0' + byte(digit)
		}

		v /= 10
	}

	if include {
		wr--
		buf[wr] = '.'
	}

	// only use written buffer
	return v, buf[wr:]
}

func writeUnit(sb *strings.Builder, v uint, prec int) {
	v, buf := fmtFrac(v, prec)
	if v != 0 {
		sb.WriteString(strconv.Itoa(int(v)))
	}

	sb.Write(buf)
}

func (d ISODuration) String() string {
	var sb strings.Builder
	sb.WriteByte('P')

	if d.MYears != 0 {
		writeUnit(&sb, d.MYears, 3)
		sb.WriteByte('Y')
	}
	if d.MMonths != 0 {
		writeUnit(&sb, d.MMonths, 3)
		sb.WriteByte('M')
	}
	if d.MDays != 0 {
		writeUnit(&sb, d.MDays, 3)
		sb.WriteByte('D')
	}

	if d.MHours != 0 || d.MMinutes != 0 || d.MSeconds != 0 {
		sb.WriteByte('T')
	}

	if d.MHours != 0 {
		writeUnit(&sb, d.MHours, 3)
		sb.WriteByte('H')
	}
	if d.MMinutes != 0 {
		writeUnit(&sb, d.MMinutes, 3)
		sb.WriteByte('M')
	}
	if d.MSeconds != 0 {
		writeUnit(&sb, d.MSeconds, 3)
		sb.WriteByte('S')
	}

	return sb.String()
}

// IsExact checks whether the duration is exact meaning it doesn't contain
// months or years as those aren't well-defined time spans.
func (d ISODuration) IsExact() bool {
	return d.MYears == 0 && d.MMonths == 0
}

// Normalize normalises the duration by using the most appropriate units of time.
// Normalisations thereby eliminates decimal fractions.
// Normalised durations DON'T use month or year units!
func (d ISODuration) Normalize() ISODuration {
	return DurationFromMS(d.Milliseconds())
}

// Milliseconds returns the amount of milliseconds of the time span.
// Note that it assumes that years have 365 and months 30 days.
func (d ISODuration) Milliseconds() uint64 {
	return uint64(d.MYears)*mYear +
		uint64(d.MMonths)*mMonth +
		uint64(d.MDays)*mDay +
		uint64(d.MHours)*mHour +
		uint64(d.MMinutes)*mMinute +
		uint64(d.MSeconds)
}

// AsDuration returns the iso duration as a time.Duration.
// Years and months are treated as in Milliseconds().
func (d ISODuration) AsDuration() time.Duration {
	return time.Duration(d.Milliseconds()) * time.Millisecond
}

// TotalSeconds returns the total amount of seconds in the duration.
// Years and months are treated as in Milliseconds().
func (d ISODuration) TotalSeconds() float64 {
	ms := d.Milliseconds()
	s := ms / 1000
	ms = ms % 1000
	return float64(s) + float64(ms)/1000
}

// DurationFromMS creates a duration from the amount of milliseconds.
// Keeps the units within range, but doesn't use months or years.
func DurationFromMS(ms uint64) ISODuration {
	var mm, mh, md uint64

	md, ms = second*(ms/day), ms%day
	mh, ms = second*(ms/hour), ms%hour
	mm, ms = second*(ms/minute), ms%minute

	return ISODuration{
		MDays:    uint(md),
		MHours:   uint(mh),
		MMinutes: uint(mm),
		MSeconds: uint(ms),
	}
}

var (
	// ErrInvalidFormat is returned when an ISO 6801 duration isn't formatted
	// correctly.
	ErrInvalidFormat = errors.New("iso8601: invalid format")
)

func parsePart(p string) (uint, error) {
	// decimal fraction can be specified using either comma or dot!
	dec := strings.IndexByte(p, '.')
	if dec == -1 {
		dec = strings.IndexByte(p, ',')
	}

	scale := 3

	if dec >= 0 {
		decimals := len(p) - dec - 1
		if decimals > 0 {
			frac := p[dec+1:]
			// limit decimal fractions to 3 digits
			if len(frac) > scale {
				frac = frac[:scale]
			}

			scale = scale - len(frac)

			// remove the decimal point
			p = p[:dec] + frac
		} else {
			// if there's nothing but a trailing comma/dot, just remove it
			p = p[:len(p)-1]
		}
	}

	// no decimal fraction, handle normally
	n, err := strconv.Atoi(p)
	if err != nil {
		return 0, err
	}

	return uint(math.Pow10(scale)) * uint(n), nil
}

func parseDate(dur *ISODuration, d string) error {
	if i := strings.IndexByte(d, 'Y'); i >= 0 {
		n, err := parsePart(d[:i])
		if err != nil {
			return err
		}

		dur.MYears = n

		d = d[i+1:]
	}

	if i := strings.IndexByte(d, 'M'); i >= 0 {
		n, err := parsePart(d[:i])
		if err != nil {
			return err
		}

		dur.MMonths = n

		d = d[i+1:]
	}

	if i := strings.IndexByte(d, 'D'); i >= 0 {
		n, err := parsePart(d[:i])
		if err != nil {
			return err
		}

		dur.MDays = n

		d = d[i+1:]
	}

	// If there's still something left after D, something's wrong with the format.
	if len(d) > 0 {
		return ErrInvalidFormat
	}

	return nil
}

func parseTime(dur *ISODuration, t string) error {
	if i := strings.IndexByte(t, 'H'); i >= 0 {
		n, err := parsePart(t[:i])
		if err != nil {
			return err
		}

		dur.MHours = n

		t = t[i+1:]
	}

	if i := strings.IndexByte(t, 'M'); i >= 0 {
		n, err := parsePart(t[:i])
		if err != nil {
			return err
		}

		dur.MMinutes = n

		t = t[i+1:]
	}

	if i := strings.IndexByte(t, 'S'); i >= 0 {
		n, err := parsePart(t[:i])
		if err != nil {
			return err
		}

		dur.MSeconds = n

		t = t[i+1:]
	}

	if len(t) > 0 {
		return ErrInvalidFormat
	}

	return nil
}

// ParseDuration parses an ISO 6801 duration.
// It doesn't parse durations using weeks or date-time notation.
func ParseDuration(duration string) (dur ISODuration, err error) {
	if duration == "" || duration[0] != 'P' {
		return dur, ErrInvalidFormat
	}

	duration = duration[1:]

	i := strings.IndexByte(duration, 'T')
	if i >= 0 {
		if err = parseTime(&dur, duration[i+1:]); err != nil {
			return
		}
	} else {
		i = len(duration)
	}

	if err = parseDate(&dur, duration[:i]); err != nil {
		return
	}

	return
}
