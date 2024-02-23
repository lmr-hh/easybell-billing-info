package main

import (
	"fmt"
	"math"
	"time"

	"github.com/atc0005/go-teams-notify/v2/adaptivecard"

	"github.com/lmr-hh/easybell-billing-info/easybell"
)

// calculateCost calculates the expected cost for u being over the quota.
func calculateCost(u easybell.Usage) float64 {
	return math.Ceil(max(u.National-NationalQuota, 0).Minutes())*NationalMinutePrice +
		math.Ceil(max(u.Mobile-MobileQuota, 0).Minutes())*MobileMinutePrice
}

// printUsage formats and prints u to stdout.
func printUsage(u easybell.Usage) {
	fmt.Printf("  National:      %06s / %04.0f:00 (%.2f %%)\n", formatDuration(u.National), NationalQuota.Minutes(), float64(u.National)/float64(NationalQuota)*100)
	fmt.Printf("  Mobile:         %5s /  %03.0f:00 (%.2f %%)\n", formatDuration(u.Mobile), MobileQuota.Minutes(), float64(u.Mobile)/float64(MobileQuota)*100)
	fmt.Printf("  International:  %5s /   00:00\n", formatDuration(u.Other))
}

// formatDuration formats d in a user-friendly way as mm:ss.
func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%02.0f:%02.0f", d.Truncate(time.Minute).Minutes(), (d - d.Truncate(time.Minute)).Seconds())
}

// makeGaugeElement returns a column with the specified text and value.
func makeGaugeElement(text, value, alignment, weight, color string) adaptivecard.Column {
	return adaptivecard.Column{
		Type:  adaptivecard.TypeColumn,
		Width: adaptivecard.ColumnWidthStretch,
		Items: []*adaptivecard.Element{{
			Type:                adaptivecard.TypeElementTextBlock,
			Text:                text,
			IsSubtle:            true,
			HorizontalAlignment: alignment,
		}, {
			Type:                adaptivecard.TypeElementTextBlock,
			Text:                value,
			Spacing:             adaptivecard.SpacingNone,
			Size:                adaptivecard.SizeExtraLarge,
			Weight:              weight,
			Color:               color,
			HorizontalAlignment: alignment,
		}},
	}
}

// minutesColor chooses a color for formatting d depending on how near d is to its quota.
func minutesColor(d, quota time.Duration, goodColor string) string {
	if d == 0 || d <= 0.9*60*quota {
		return goodColor
	}
	if d <= quota || d < 1.1*60*quota {
		return adaptivecard.ColorWarning
	}
	return adaptivecard.ColorAttention
}

// months contains the German month names.
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
