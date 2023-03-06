package easybell

import (
	"encoding/json"
	"time"
)

type Time time.Time

func (t *Time) UnmarshalJSON(b []byte) error {
	var value string
	err := json.Unmarshal(b, &value)
	if err != nil {
		return err
	}
	parsed, err := time.Parse("02.01.2006 15:04:05", value)
	if err != nil {
		return err
	}
	*t = Time(parsed)
	return nil
}

type CallLogEntry struct {
	ID             string        `json:"ID"`
	Deleted        string        `json:"DELETED"`
	Time           Time          `json:"DATUM"`
	Minutes        int           `json:"DAUER"`
	Number         string        `json:"RUFNUMMER"`
	Direction      CallDirection `json:"RICHTUNG"`
	Partner        string        `json:"PARTNER"`
	CallType       CallType      `json:"TYPE"`
	Status         string        `json:"STATUS"`
	Kind           CallKind      `json:"ART"`
	FaxStatus      string        `json:"FAXSTATUS"`
	FaxErrorReason string        `json:"FAXERRORREASON"`
}

type CallLogPage struct {
	LastPage int             `json:"last_page"`
	Calls    []*CallLogEntry `json:"data"`
}
