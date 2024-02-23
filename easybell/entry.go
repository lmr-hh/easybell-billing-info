package easybell

import (
	"encoding/json"
	"time"
)

// These constants identify known direction filters of the easyBell API.
const (
	CallDirectionAny                = "*"
	CallDirectionAnyOutbound        = "1"
	CallDirectionAnyInbound         = "2"
	CallDirectionSuccessfulOutbound = "11"
	CallDirectionSuccessfulInbound  = "21"
	CallDirectionFailedOutbound     = "12"
	CallDirectionFailedInbound      = "22"
)

// These constants identify known call type values of the easyBell API.
const (
	CallTypeAny        = "*"
	CallTypeForward    = "forward"
	CallTypeRegular    = "call"
	CallTypeVoicebox   = "voicebox"
	CallTypeFax2Mail   = "fax2mail"
	CallTypeSMS        = "sms"
	CallTypeConference = "conference"
)

// These constants identify known call kind values of the easyBell API.
const (
	CallKindAny           = "*"
	CallKindMobile        = "mobile"
	CallKindNational      = "national"
	CallKindInternational = "international"
)

// A CallLogEntry represents a single call as returned from the easyBell API.
type CallLogEntry struct {
	ID             string
	Deleted        string
	Time           time.Time
	Duration       time.Duration
	Number         string
	Direction      string
	Partner        string
	CallType       string
	Status         string
	Kind           string
	FaxStatus      string
	FaxErrorReason string
}

func (e *CallLogEntry) UnmarshalJSON(data []byte) (err error) {
	aux := struct {
		ID             string `json:"ID"`
		Deleted        string `json:"DELETED"`
		Time           string `json:"DATUM"`
		Duration       int    `json:"DAUER"`
		Number         string `json:"RUFNUMMER"`
		Direction      string `json:"RICHTUNG"`
		Partner        string `json:"PARTNER"`
		CallType       string `json:"TYPE"`
		Status         string `json:"STATUS"`
		Kind           string `json:"ART"`
		FaxStatus      string `json:"FAXSTATUS"`
		FaxErrorReason string `json:"FAXERRORREASON"`
	}{}
	if err = json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*e = CallLogEntry{
		ID:             aux.ID,
		Deleted:        aux.Deleted,
		Duration:       time.Duration(aux.Duration) * time.Second,
		Number:         aux.Number,
		Direction:      aux.Direction,
		Partner:        aux.Partner,
		CallType:       aux.CallType,
		Status:         aux.Status,
		Kind:           aux.Kind,
		FaxStatus:      aux.FaxStatus,
		FaxErrorReason: aux.FaxErrorReason,
	}
	if e.Time, err = time.Parse("02.01.2006 15:04:05", string(aux.Time)); err != nil {
		return err
	}
	return nil
}
