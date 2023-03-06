package MonthlyUsage

import (
	"bytes"
	"encoding/json"
	"github.com/lmr-hh/functions/pkg/easybell"
	"net/http"
	"time"
)

type Handler struct {
	location *time.Location
	username string
	password string
	url      string

	NationalMinutes int
	MobileMinutes   int
}

func NewHandler(username string, password string, location *time.Location, url string) *Handler {
	handler := &Handler{}
	handler.location = location
	if handler.location == nil {
		handler.location = time.UTC
	}

	handler.username = username
	handler.password = password
	handler.url = url
	return handler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.sendMonthReport()
	if err != nil {
		_ = h.reportError(err)
		http.Error(w, "Error", http.StatusInternalServerError)
	} else {
		http.Error(w, "OK", http.StatusOK)
	}
}

func (h *Handler) sendMonthReport() error {
	now := time.Now().In(h.location)
	year, month, _ := now.Date()
	start := time.Date(year, month-1, 1, 0, 0, 0, 0, h.location)
	end := start.AddDate(0, 1, 0)

	eb := &easybell.Client{}
	err := eb.Login(h.username, h.password)
	if err != nil {
		return err
	}
	defer func() {
		logoutErr := eb.Logout()
		if err == nil {
			err = logoutErr
		}
	}()

	national, mobile, other, err := eb.CollectCalls(start, end, &easybell.CallLogFilter{DirectionFilter: easybell.SuccessfulOutbound})
	if err != nil {
		return err
	}

	card := h.lastMonthUsageCard(start, national, mobile, other)
	err = card.Prepare()
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"type": "message",
		"attachments": []interface{}{
			map[string]interface{}{
				"contentType": "application/vnd.microsoft.card.adaptive",
				"contentUrl":  nil,
				"content":     card,
			},
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = http.Post(h.url, "application/json", bytes.NewReader(b))
	return err
}

func (h *Handler) reportError(err error) error {
	data, err := json.Marshal(map[string]interface{}{
		"text": "Es ist ein Fehler beim Erstellen der easyBell Monatsübersicht aufgetreten. Bitte melde dies an den Administrator. Die Fehlermeldung ist '" + err.Error() + "'",
	})
	if err != nil {
		return err
	}
	_, err = http.Post(h.url, "application/json", bytes.NewReader(data))
	return err
}
