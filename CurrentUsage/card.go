package CurrentUsage

import (
	"fmt"
	"github.com/lmr-hh/functions/pkg/easybell"
	"math"
	"time"

	cards "github.com/DanielTitkov/go-adaptive-cards"
)

func (h *Handler) makeCard(now time.Time, currentNational, currentMobile, currentOther, estimateNational, estimateMobile, estimateOther int) *cards.Card {
	national := fmt.Sprintf("%02d:%02d", currentNational/60, currentNational%60)
	mobile := fmt.Sprintf("%02d:%02d", currentMobile/60, currentMobile%60)
	other := fmt.Sprintf("%02d:%02d", currentOther/60, currentOther%60)
	estNational := fmt.Sprintf("%02d min.", (estimateNational+59)/60)
	estMobile := fmt.Sprintf("%02d min.", (estimateMobile+59)/60)
	estOther := fmt.Sprintf("%02d min.", (estimateOther+59)/60)

	extraCost := math.Max(float64((estimateNational+59)/60-h.NationalMinutes), 0)*easybell.NationalMinutePrice +
		math.Max(float64((estimateMobile+59)/60-h.MobileMinutes), 0)*easybell.MobileMinutePrice

	otherCallsVisible := cards.FalsePtr()
	if currentOther > 0 {
		otherCallsVisible = cards.TruePtr()
	}

	return &cards.Card{
		Type:    cards.AdaptiveCardType,
		Version: "1.4",
		Schema:  cards.DefaultSchema,
		Body: []cards.Node{
			&cards.Container{
				Items: []cards.Node{
					&cards.TextBlock{
						Text:   "easyBell Telefonieverbrauch",
						Wrap:   cards.TruePtr(),
						Weight: "bolder",
						Size:   "extraLarge",
					},
					&cards.TextBlock{
						Text:     now.Format("01/2006 (vorläufig)"),
						Wrap:     cards.TruePtr(),
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
						Spacing: "medium",
						Columns: []*cards.Column{
							makeGaugeElement("Festnetz", national, "left", "bolder"),
							makeGaugeElement("Mobil", mobile, "center", "bolder"),
							makeGaugeElement("Andere", other, "right", "bolder"),
						},
					},
				},
			},
			&cards.Container{
				Separator: cards.TruePtr(),
				Items: []cards.Node{
					&cards.TextBlock{
						Text:   "Prognose zum Monatsende",
						Size:   "large",
						Weight: "bolder",
					},
					&cards.TextBlock{
						Text:     "Diese Daten sind ein Schätzwert für den Telefonverbrauch am Monatsende. Sie beruhen auf den Daten der letzten vier Wochen.",
						Wrap:     cards.TruePtr(),
						Spacing:  "none",
						Size:     "small",
						IsSubtle: cards.TruePtr(),
					},
					&cards.ColumnSet{
						Columns: []*cards.Column{
							makeGaugeElement(fmt.Sprintf("Festnetz (%d)", h.NationalMinutes), estNational, "left", minutesColor(estimateNational, h.NationalMinutes)),
							makeGaugeElement(fmt.Sprintf("Mobil (%d)", h.MobileMinutes), estMobile, "center", minutesColor(estimateMobile, h.MobileMinutes)),
							makeGaugeElement("Andere", estOther, "right", minutesColor(estimateOther, 0)),
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
										Text: fmt.Sprintf("%.2f €", extraCost),
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
						IsVisible: otherCallsVisible,
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
				Color:               color,
				HorizontalAlignment: alignment,
			},
		},
	}
}

func minutesColor(seconds, quota int) string {
	if seconds == 0 || seconds <= 0.9*60*quota {
		return "good"
	}
	if seconds <= quota || seconds < 1.1*60*quota {
		return "warning"
	}
	return "attention"
}
