package main

import (
	"fmt"
	"math"
	"time"

	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
	"github.com/spf13/cobra"

	"github.com/lmr-hh/easybell-billing-info/easybell"
)

func init() {
	rootCommand.AddCommand(lastMonthCommand)
}

// lastMonthCommand implements reporting the usage of the past month.
var lastMonthCommand = &cobra.Command{
	Use:   "last-month",
	Short: "Report the previous month's usage.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		now := time.Now()
		year, month, _ := now.Date()
		start := time.Date(year, month-1, 1, 0, 0, 0, 0, time.Local)
		end := start.AddDate(0, 1, 0)

		reader := easybell.NewCallLogReader(client, start, end)
		reader.Direction = easybell.CallDirectionSuccessfulOutbound
		usage, err := reader.ReadUsage()
		if err != nil {
			return err
		}

		printPreviousUsageReport(start, usage)
		if !sendWebhook {
			return nil
		}
		return sendPreviousUsageReport(start, usage)
	},
}

// printPreviousUsageReport prints the usage of the past month to the command line.
func printPreviousUsageReport(when time.Time, usage easybell.Usage) {
	fmt.Printf("EasyBell Usage Report for %s %d\n\n", when.Month().String(), when.Year())
	printUsage(usage)
}

// sendPreviousUsageReport sends a teams message with the usage of the past month.
func sendPreviousUsageReport(when time.Time, usage easybell.Usage) error {
	otherCallsVisible := usage.Other > 0
	card := adaptivecard.Card{
		Type:         adaptivecard.TypeAdaptiveCard,
		Schema:       adaptivecard.AdaptiveCardSchema,
		Version:      "1.4",
		FallbackText: "",
		Body: adaptivecard.Elements{{
			Type: adaptivecard.TypeElementContainer,
			Items: adaptivecard.Elements{{
				Type:   adaptivecard.TypeElementTextBlock,
				Text:   "easyBell Monatsübersicht",
				Weight: adaptivecard.WeightBolder,
				Size:   adaptivecard.SizeExtraLarge,
			}, {
				Type:     adaptivecard.TypeElementTextBlock,
				Text:     fmt.Sprintf("%s %d", months[when.Month()], when.Year()),
				Spacing:  adaptivecard.SpacingNone,
				IsSubtle: true,
				Weight:   adaptivecard.WeightBolder,
			}},
		}, {
			Type:      adaptivecard.TypeElementContainer,
			Separator: true,
			Items: adaptivecard.Elements{{
				Type: adaptivecard.TypeElementColumnSet,
				Columns: adaptivecard.Columns{
					makeGaugeElement(fmt.Sprintf("Festnetz (%.0f)", NationalQuota.Minutes()), fmt.Sprintf("%.0f min.", math.Ceil(usage.National.Minutes())), adaptivecard.HorizontalAlignmentLeft, adaptivecard.WeightBolder, minutesColor(usage.National, NationalQuota, adaptivecard.ColorGood)),
					makeGaugeElement(fmt.Sprintf("Mobil (%.0f)", MobileQuota.Minutes()), fmt.Sprintf("%.0f min.", math.Ceil(usage.Mobile.Minutes())), adaptivecard.HorizontalAlignmentCenter, adaptivecard.WeightBolder, minutesColor(usage.Mobile, MobileQuota, adaptivecard.ColorGood)),
					makeGaugeElement("Andere", fmt.Sprintf("%.0f min.", math.Ceil(usage.Other.Minutes())), adaptivecard.HorizontalAlignmentRight, adaptivecard.WeightBolder, minutesColor(usage.Other, 0, adaptivecard.ColorGood)),
				},
			}, {
				Type: adaptivecard.TypeElementColumnSet,
				Columns: adaptivecard.Columns{{
					Type:  adaptivecard.TypeColumn,
					Width: adaptivecard.ColumnWidthStretch,
					Items: []*adaptivecard.Element{{
						Type:   adaptivecard.TypeElementTextBlock,
						Text:   "Zusätzliche Kosten",
						Weight: adaptivecard.WeightBolder,
					}},
				}, {
					Type:  adaptivecard.TypeColumn,
					Width: adaptivecard.ColumnWidthAuto,
					Items: []*adaptivecard.Element{{
						Type:   adaptivecard.TypeElementTextBlock,
						Text:   fmt.Sprintf("%.2f €", calculateCost(usage)),
						Weight: adaptivecard.WeightBolder,
					}},
				}},
			}},
		}, {
			Type:    adaptivecard.TypeElementTextBlock,
			Text:    "Es sind in diesem Zeitraum internationale Anrufe getätigt worden. In der Kostenschätzung sind diese nicht berücksichtigt.",
			Wrap:    true,
			Spacing: adaptivecard.SpacingNone,
			Color:   adaptivecard.ColorWarning,
			Visible: &otherCallsVisible,
		}, {
			Type:      adaptivecard.TypeElementContainer,
			Separator: true,
			Items: adaptivecard.Elements{{
				Type: adaptivecard.TypeElementTextBlock,
				Text: "Diese Angaben sind Schätzwerte auf Basis der Anrufliste. Diese Angaben sollten mit dem Einzelverbindungsnachweis, bzw. der Rechnung des Monats abgeglichen werden.",
				Wrap: true,
				Size: adaptivecard.SizeSmall,
			}},
		}},
	}
	if msg, err := adaptivecard.NewMessageFromCard(card); err != nil {
		return err
	} else {
		return teamsClient.Send(teamsWebhookURL, msg)
	}
}
