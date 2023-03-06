package MonthlyUsage

import (
	"fmt"
	"github.com/lmr-hh/functions/pkg/easybell"
	"math"
	"time"

	cards "github.com/DanielTitkov/go-adaptive-cards"
)

var months = map[time.Month]string{
	1:  "Januar",
	2:  "Februar",
	3:  "März",
	4:  "April",
	5:  "Mai",
	6:  "Juni",
	7:  "Juli",
	8:  "August",
	9:  "September",
	10: "Oktober",
	11: "November",
	12: "Dezember",
}

func (h *Handler) lastMonthUsageCard(date time.Time, national, mobile, other int) *cards.Card {
	nationalStr := fmt.Sprintf("%02d min.", (national+59)/60)
	mobileStr := fmt.Sprintf("%02d min.", (mobile+59)/60)
	otherStr := fmt.Sprintf("%02d min.", (other+59)/60)

	extraCost := math.Max(float64((national+59)/60-h.NationalMinutes), 0)*easybell.NationalMinutePrice +
		math.Max(float64((mobile+59)/60-h.MobileMinutes), 0)*easybell.MobileMinutePrice

	otherVisible := cards.FalsePtr()
	if other > 0 {
		otherVisible = cards.TruePtr()
	}
	return &cards.Card{
		Schema:  cards.DefaultSchema,
		Type:    cards.AdaptiveCardType,
		Version: "1.4",
		Body: []cards.Node{
			&cards.Container{
				Items: []cards.Node{
					&cards.TextBlock{
						Text:   "easyBell Monatsübersicht",
						Weight: "bolder",
						Size:   "extraLarge",
					},
					&cards.TextBlock{
						Text:     fmt.Sprintf("%s %d", months[date.Month()], date.Year()),
						Spacing:  "none",
						IsSubtle: cards.TruePtr(),
						Weight:   "bolder",
					},
				},
			},
			&cards.Container{
				Separator: cards.TruePtr(),
				Items: []cards.Node{
					&cards.ColumnSet{
						Columns: []*cards.Column{
							makeGaugeElement(fmt.Sprintf("Festnetz (%d)", h.NationalMinutes), nationalStr, "left", minutesColor(national, h.NationalMinutes)),
							makeGaugeElement(fmt.Sprintf("Mobil (%d)", h.MobileMinutes), mobileStr, "center", minutesColor(mobile, h.MobileMinutes)),
							makeGaugeElement("Andere", otherStr, "right", minutesColor(other, 0)),
						},
					},
					&cards.ColumnSet{
						Columns: []*cards.Column{
							{
								Width: "stretch",
								Items: []cards.Node{
									&cards.TextBlock{
										Text:   "Zusätzliche Kosten",
										Weight: "bolder",
									},
								},
							},
							{
								Width: "auto",
								Items: []cards.Node{
									&cards.TextBlock{
										Text:   fmt.Sprintf("%.2f €", extraCost),
										Weight: "bolder",
									},
								},
							},
						},
					},
				},
			},
			&cards.TextBlock{
				Text:      "Es sind in dieser Zeitspanne internationale Anrufe getätigt worden. In der Kostenschätzung sind diese nicht berücksichtigt.",
				Wrap:      cards.TruePtr(),
				Spacing:   "none",
				Color:     "warning",
				IsVisible: otherVisible,
			},
			&cards.Container{
				Separator: cards.TruePtr(),
				Items: []cards.Node{
					&cards.TextBlock{
						Text: "Diese Angaben sind Schätzwerte auf Basis der Anrufliste. Diese Angaben sollten mit dem Einzelverbindungsnachweis, bzw. der Rechnung des Monats abgeglichen werden.",
						Wrap: cards.TruePtr(),
						Size: "small",
					},
				},
			},
		},
	}
}

func makeGaugeElement(text string, value string, alignment string, color string) *cards.Column {
	return &cards.Column{
		Width: "stretch",
		Items: []cards.Node{
			&cards.TextBlock{
				Text:                text,
				IsSubtle:            cards.TruePtr(),
				HorizontalAlignment: alignment,
			},
			&cards.TextBlock{
				Text:                value,
				Spacing:             "none",
				Size:                "extraLarge",
				Weight:              "bolder",
				Color:               color,
				HorizontalAlignment: alignment,
			},
		},
	}
}

func minutesColor(seconds, quota int) string {
	if seconds == 0 || seconds <= quota*60 {
		return "good"
	}
	if seconds < 1.1*60*quota {
		return "warning"
	}
	return "attention"
}
