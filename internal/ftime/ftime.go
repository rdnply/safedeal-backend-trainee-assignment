package ftime

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"time"
)

type FormatTime struct {
	time.Time
}

const Layout = "2006-01-02T15:04:05Z"

func New(t time.Time) *FormatTime {
	return &FormatTime{Time: t}
}

func (f FormatTime) MarshalJSON() ([]byte, error) {
	t := f.Time.UTC()
	s := t.Format(time.RFC3339)

	t, err := time.Parse(Layout, s)
	if err != nil {
		return nil, err
	}

	s = fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ", t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	return []byte(`"` + s + `"`), nil
}

func (f *FormatTime) UnmarshalJSON(b []byte) error {
	s := string(b)

	t, err := time.Parse(`"`+Layout+`"`, s)
	if err != nil {
		return err
	}

	loc, err := time.LoadLocation("Local")
	if err != nil {
		return err
	}

	f.Time = t.In(loc)

	return nil
}

func (f *FormatTime) Scan(value interface{}) error {
	var t sql.NullTime
	if err := t.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		return nil
	} else {
		t := t.Time.UTC()
		s := t.Format(time.RFC3339)
		t, err := time.Parse(Layout, s)
		if err != nil {
			return err
		}

		loc, err := time.LoadLocation("Local")
		if err != nil {
			return err
		}

		f.Time = t.In(loc)
	}

	return nil
}

func (f FormatTime) Value() (driver.Value, error) {
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return nil, err
	}

	return f.Time.In(loc), nil
}
