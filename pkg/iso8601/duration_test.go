package iso8601

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestParseDuration(t *testing.T) {
	type args struct {
		duration string
	}

	tests := []struct {
		name    string
		args    args
		wantDur ISODuration
		wantErr bool
	}{
		{
			name: "full",
			args: args{duration: "P3DT2H50M57S"},
			wantDur: ISODuration{
				MDays:    3000,
				MHours:   2000,
				MMinutes: 50000,
				MSeconds: 57000,
			},
		},
		{
			name: "time only",
			args: args{duration: "PT0H50M57S"},
			wantDur: ISODuration{
				MHours:   0,
				MMinutes: 50000,
				MSeconds: 57000,
			},
		},
		{
			name: "date only",
			args: args{duration: "P5Y3M1D"},
			wantDur: ISODuration{
				MYears:  5000,
				MMonths: 3000,
				MDays:   1000,
			},
		},
		{
			name: "fractions milli",
			args: args{duration: "PT5,001S"},
			wantDur: ISODuration{
				MSeconds: 5001,
			},
		},
		{
			name: "fractions deci",
			args: args{duration: "PT5.12S"},
			wantDur: ISODuration{
				MSeconds: 5120,
			},
		},
		{
			name: "fractions centi",
			args: args{duration: "PT,9S"},
			wantDur: ISODuration{
				MSeconds: 900,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDur, err := ParseDuration(tt.args.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDur, tt.wantDur) {
				t.Errorf("ParseDuration() gotDur = %v, want %v", gotDur, tt.wantDur)
			}
		})
	}
}

func TestISODuration_Milliseconds(t *testing.T) {
	dur := ISODuration{
		MDays:    3000,
		MHours:   2500,
		MMinutes: 0,
		MSeconds: 700,
	}

	assert.Equal(t, uint64(268200700), dur.Milliseconds())
}

func TestDurationFromMilliseconds(t *testing.T) {
	dur := DurationFromMS(268200700)
	assert.Equal(t, ISODuration{
		MDays:    3000,
		MHours:   2000,
		MMinutes: 30000,
		MSeconds: 700,
	}, dur)
}

func TestISODuration_AsDuration(t *testing.T) {
	dur := DurationFromMS(192839182)
	d := dur.AsDuration()
	assert.Equal(t, dur.Milliseconds(), uint64(d.Milliseconds()))
}

func TestISODuration_String(t *testing.T) {
	dur := ISODuration{
		MYears:   1050,
		MMonths:  9500,
		MDays:    1000,
		MHours:   12500,
		MMinutes: 3333,
		MSeconds: 50,
	}
	assert.Equal(t, dur.String(), "P1.05Y9.5M1DT12.5H3.333M.05S")

	dur = ISODuration{
		MMinutes: 50000,
	}
	assert.Equal(t, dur.String(), "PT50M")
}
