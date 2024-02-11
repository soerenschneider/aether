package astral

import (
	"bytes"
	"context"
	"html/template"
	"sync"
	"time"

	"github.com/sj14/astral/pkg/astral"
)

type Astral struct {
	location *time.Location
	observer astral.Observer
	once     sync.Once
	template *template.Template
}

func New(lat, lon float64) (*Astral, error) {
	a := &Astral{
		observer: astral.Observer{
			Latitude:  lat,
			Longitude: lon,
		},
		location: time.Now().Location(),
	}

	return a, nil
}

func (b *Astral) GetHtml(_ context.Context) (string, error) {
	b.once.Do(func() {
		if b.template == nil {
			b.template = template.Must(template.New("astral").Parse(defaultTemplate))
		}
	})

	data, err := b.get(time.Now().In(b.location))
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = b.template.Execute(&tpl, data)
	return tpl.String(), err
}

type astralData struct {
	BlueHourRising    TimeDuration
	BlueHourSetting   TimeDuration
	GoldenHourRising  TimeDuration
	GoldenHourSetting TimeDuration
}

type TimeDuration struct {
	Start time.Time
	End   time.Time
}

func (b *Astral) get(date time.Time) (*astralData, error) {
	ret := &astralData{}

	var err error
	ret.BlueHourRising.Start, ret.BlueHourRising.End, err = astral.BlueHour(b.observer, date, astral.SunDirectionRising)
	if err != nil {
		return nil, err
	}

	ret.BlueHourSetting.Start, ret.BlueHourSetting.End, err = astral.BlueHour(b.observer, date, astral.SunDirectionSetting)
	if err != nil {
		return nil, err
	}

	ret.GoldenHourRising.Start, ret.GoldenHourRising.End, err = astral.GoldenHour(b.observer, date, astral.SunDirectionRising)
	if err != nil {
		return nil, err
	}

	ret.GoldenHourSetting.Start, ret.GoldenHourSetting.End, err = astral.GoldenHour(b.observer, date, astral.SunDirectionSetting)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (b *Astral) Name() string {
	return "Astral"
}
