package main

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
	"github.com/spf13/cobra"

	"github.com/lmr-hh/easybell-billing-info/easybell"
)

var estimationPeriod time.Duration

func init() {
	currentMonthCommand.Flags().DurationVarP(&estimationPeriod, "estimate", "e", 35*24*time.Hour, "The number of days to include when estimating the usage until the end of the month.")
	rootCommand.AddCommand(currentMonthCommand)
}

var currentMonthCommand = &cobra.Command{
	Use:   "current-month",
	Short: "Report the current month's usage and an estimate to the end of the month.",
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if estimationPeriod <= 24*time.Hour {
			return errors.New("estimation period must be at least 1 day")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		now := time.Now()
		year, month, _ := now.Date()
		startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
		estimationStart := now.Add(-estimationPeriod)
		endOfMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.Local)
		fullMonth := endOfMonth.Sub(startOfMonth)

		reader := easybell.NewCallLogReader(client, startOfMonth, endOfMonth)
		reader.Direction = easybell.CallDirectionSuccessfulOutbound
		currentUsage, err := reader.ReadUsage()
		if err != nil {
			return err
		}

		reader.Reset(estimationStart, now)
		pastUsage, err := reader.ReadUsage()
		if err != nil {
			return err
		}

		factor := fullMonth.Hours() / estimationPeriod.Hours()
		estimateUsage := easybell.Usage{
			National: time.Duration(float64(pastUsage.National) * factor),
			Mobile:   time.Duration(float64(pastUsage.Mobile) * factor),
			Other:    time.Duration(float64(pastUsage.Other) * factor),
		}

		printCurrentUsageReport(now, currentUsage, estimateUsage)
		if !sendWebhook {
			return nil
		}
		return sendCurrentUsageReport(now, currentUsage, estimateUsage)
	},
}

func printCurrentUsageReport(now time.Time, currentUsage easybell.Usage, estimateUsage easybell.Usage) {
	fmt.Printf("EasyBell Usage Report for %s %d\n\n", now.Month().String(), now.Year())
	fmt.Printf("This Month:\n")
	printUsage(currentUsage)
	fmt.Printf("\nEstimated Usage at the End of the Month:\n")
	printUsage(estimateUsage)
	fmt.Printf("\nThe estimate is based on the average usage of the last %.1f days.\n", estimationPeriod.Hours()/24)
}

func sendCurrentUsageReport(now time.Time, currentUsage easybell.Usage, estimateUsage easybell.Usage) error {
	otherCallsVisible := currentUsage.Other > 0
	card := adaptivecard.Card{
		Type:         adaptivecard.TypeAdaptiveCard,
		Schema:       adaptivecard.AdaptiveCardSchema,
		Version:      "1.4",
		FallbackText: "",
		Body: adaptivecard.Elements{{
			Type: adaptivecard.TypeElementContainer,
			Items: adaptivecard.Elements{{
				Type:   adaptivecard.TypeElementTextBlock,
				Text:   "easyBell Telefonieverbrauch",
				Wrap:   true,
				Weight: adaptivecard.WeightBolder,
				Size:   adaptivecard.SizeExtraLarge,
			}, {
				Type:     adaptivecard.TypeElementTextBlock,
				Text:     fmt.Sprintf("%s %d", months[now.Month()], now.Year()),
				Wrap:     true,
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
					makeGaugeElement("Festnetz", formatDuration(currentUsage.National), adaptivecard.HorizontalAlignmentLeft, adaptivecard.WeightDefault, minutesColor(currentUsage.National, NationalQuota, adaptivecard.ColorDefault)),
					makeGaugeElement("Mobil", formatDuration(currentUsage.Mobile), adaptivecard.HorizontalAlignmentCenter, adaptivecard.WeightDefault, minutesColor(currentUsage.Mobile, MobileQuota, adaptivecard.ColorDefault)),
					makeGaugeElement("Andere", formatDuration(currentUsage.Other), adaptivecard.HorizontalAlignmentRight, adaptivecard.WeightDefault, minutesColor(currentUsage.Other, 0, adaptivecard.ColorDefault)),
				},
			}},
		}, {
			Type:      adaptivecard.TypeElementContainer,
			Separator: true,
			Items: adaptivecard.Elements{{
				Type:   adaptivecard.TypeElementTextBlock,
				Text:   "Prognose zum Monatsende",
				Size:   adaptivecard.SizeLarge,
				Weight: adaptivecard.WeightBolder,
			}, {
				Type:     adaptivecard.TypeElementTextBlock,
				Text:     "Diese Daten sind ein Schätzwert für den Telefonverbrauch am Monatsende. Sie beruhen auf den Daten der letzten fünf Wochen.",
				Wrap:     true,
				Spacing:  adaptivecard.SpacingNone,
				Size:     adaptivecard.SizeSmall,
				IsSubtle: true,
			}, {
				Type: adaptivecard.TypeElementColumnSet,
				Columns: adaptivecard.Columns{
					makeGaugeElement(fmt.Sprintf("Festnetz (%.0f)", NationalQuota.Minutes()), fmt.Sprintf("%02.0f min.", math.Ceil(estimateUsage.National.Minutes())), adaptivecard.HorizontalAlignmentLeft, adaptivecard.WeightBolder, minutesColor(estimateUsage.National, NationalQuota, adaptivecard.ColorGood)),
					makeGaugeElement(fmt.Sprintf("Mobil (%.0f)", MobileQuota.Minutes()), fmt.Sprintf("%02.0f min.", math.Ceil(estimateUsage.Mobile.Minutes())), adaptivecard.HorizontalAlignmentCenter, adaptivecard.WeightBolder, minutesColor(estimateUsage.Mobile, MobileQuota, adaptivecard.ColorGood)),
					makeGaugeElement("Andere", fmt.Sprintf("%02.0f min.", math.Ceil(estimateUsage.Other.Minutes())), adaptivecard.HorizontalAlignmentRight, adaptivecard.WeightBolder, minutesColor(estimateUsage.Other, 0, adaptivecard.ColorGood)),
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
						Text:   fmt.Sprintf("%.2f €", calculateCost(estimateUsage)),
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
		}},
	}
	if msg, err := adaptivecard.NewMessageFromCard(card); err != nil {
		return err
	} else {
		return teamsClient.Send(teamsWebhookURL, msg)
	}
}
