package astral

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/sj14/astral/pkg/astral"
	"github.com/soerenschneider/aether/internal"
	"github.com/soerenschneider/aether/internal/templates"
	"github.com/soerenschneider/aether/pkg"
)

type Astral struct {
	location           *time.Location
	observer           astral.Observer
	defaultTemplate    *template.Template
	simpleTemplate     *template.Template
	excludeFromSummary bool
}

type Lat float64
type Lon float64

func New(lat Lat, lon Lon, templateData templates.TemplateData, location *time.Location) (*Astral, error) {
	if err := templateData.Validate(); err != nil {
		return nil, err
	}

	ds := &Astral{
		observer: astral.Observer{
			Latitude:  float64(lat),
			Longitude: float64(lon),
		},
		location: cmp.Or(location, time.UTC),
	}

	convertDegrees := func(deg float64) string {
		dir, emoji := pkg.TranslateDegreeToDirection(deg)
		return fmt.Sprintf("(%s %s)", dir, emoji)
	}

	funcMap := template.FuncMap{
		"degreesToCompass": convertDegrees,
	}

	var err error
	ds.defaultTemplate, err = template.New("astral-regular").Funcs(funcMap).Parse(string(templateData.DefaultTemplate))
	if err != nil {
		return nil, err
	}

	if len(templateData.SimpleTemplate) > 0 {
		ds.simpleTemplate, err = template.New("astral-simple").Funcs(funcMap).Parse(string(templateData.SimpleTemplate))
		if err != nil {
			return nil, err
		}
	}

	return ds, nil
}

func (b *Astral) GetData(_ context.Context) (*internal.Data, error) {
	data, err := b.get(time.Now().In(b.location))
	if err != nil {
		return nil, err
	}

	var renderedDefaultTemplate bytes.Buffer
	if err := b.defaultTemplate.Execute(&renderedDefaultTemplate, data); err != nil {
		return nil, err
	}

	var renderedSimpleTemplate bytes.Buffer
	if b.simpleTemplate != nil {
		if err := b.defaultTemplate.Execute(&renderedDefaultTemplate, data); err != nil {
			return nil, err
		}
	}

	var summary []string
	if !b.excludeFromSummary {
		summary = getSummary(*data, time.Now())
	}

	return &internal.Data{
		Summary:                    summary,
		RenderedDefaultTemplate:    renderedDefaultTemplate.Bytes(),
		RenderedSimplifiedTemplate: renderedSimpleTemplate.Bytes(),
	}, nil
}

func (b *Astral) get(date time.Time) (*AstralData, error) {
	ret := &AstralData{}

	var err error
	ret.Sunrise, err = astral.Sunrise(b.observer, date)
	if err != nil {
		return nil, err
	}

	ret.Sunset, err = astral.Sunset(b.observer, date)
	if err != nil {
		return nil, err
	}

	ret.AzimuthSunrise = astral.Azimuth(b.observer, ret.Sunrise)
	ret.AzimuthSunset = astral.Azimuth(b.observer, ret.Sunset)

	ret.BlueHourRising.Start, ret.BlueHourRising.End, err = astral.BlueHour(b.observer, date, astral.SunDirectionRising)
	if err != nil {
		return nil, err
	}

	ret.BlueHourSetting.Start, ret.BlueHourSetting.End, err = astral.BlueHour(b.observer, date, astral.SunDirectionSetting)
	if err != nil {
		return nil, err
	}
	ret.AzimuthBlueHourRising = astral.Azimuth(b.observer, ret.BlueHourRising.Start)
	ret.AzimuthBlueHourSetting = astral.Azimuth(b.observer, ret.BlueHourSetting.End)

	ret.GoldenHourRising.Start, ret.GoldenHourRising.End, err = astral.GoldenHour(b.observer, date, astral.SunDirectionRising)
	if err != nil {
		return nil, err
	}

	ret.GoldenHourSetting.Start, ret.GoldenHourSetting.End, err = astral.GoldenHour(b.observer, date, astral.SunDirectionSetting)
	if err != nil {
		return nil, err
	}

	ret.AzimuthGoldenHourRising = astral.Azimuth(b.observer, ret.GoldenHourRising.Start)
	ret.AzimuthGoldenHourSetting = astral.Azimuth(b.observer, ret.GoldenHourSetting.End)

	return ret, nil
}

func (b *Astral) Name() string {
	return "Astral"
}
