# easybell-billing-info

Send Reports of [easyBell](http://easybell.de) usage to Teams channels.

## Usage

Use `easybell-billing-info` directly or via the `ghcr.io/lmr-hh/easybell-billing-info` Docker image.

See `easybell-billing-info --help` for details.

The following configuration parameters are supported globally:

| Environment Variable        | Command Line Flag          | Description                                                  |
| --------------------------- | -------------------------- | ------------------------------------------------------------ |
| `EASYBELL_USERNAME`         | None                       | Indicates the username of the easyBell user used to retrieve data from easyBell. |
| `EASYBELL_PASSWORD`         | None                       | Indicates the password corresponding to `EASYBELL_USERNAME`. |
| `EASYBELL_NATIONAL_MINUTES` | `-n`, `--national-minutes` | The quota of included national minutes, e.g. `1000m`.        |
| `EASYBELL_MOBILE_MINUTES`   | `-m`, `--mobile-minutes`   | The quota of included mobile minutes, e.g. `200m`.           |
| None                        | `--teams-webhook`          | Enable or disable sending messages via Teams. Default is `true`. |
| `EASYBELL_TEAMS_WEBHOOK`    | `--webhook-url`            | The URL of the teams webhook. Required if `--teams-webhook` is `true`. |
| None                        | `--national-price`         | Per-minute price for national phone calls over the quota.    |
| None                        | `--mobile-price`           | Per-minute price for mobile phone calls over the quota.      |

