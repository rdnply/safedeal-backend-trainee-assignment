package types

import (
	"fmt"
	"time"
)

type FormatTime struct {
	time.Time
}

const layout = "2006-01-02T15:04:05Z"

func (f FormatTime) MarshalJSON() ([]byte, error) {
	t := f.Time.UTC()
	s := t.Format(time.RFC3339)

	t, err := time.Parse(layout, s)
	if err != nil {
		return nil, err
	}

	s = fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ", t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	return []byte(`"` + s + `"`), nil
}

func (f *FormatTime) UnmarshalJSON(b []byte) error {
	s := string(b)

	t, err := time.Parse(`"`+layout+`"`, s)
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
