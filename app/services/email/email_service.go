package email

import (
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var (
	sendgridClient = sendgrid.NewSendClient("SENDGRID-API-KEY")
	from           = mail.NewEmail("Ben Tranter", "bentranter@hey.com")
)

// Send sends an email via SendGrid.
func Send(toEmail, toName, subject, content string) error {
	message := mail.NewSingleEmail(from, subject,
		mail.NewEmail(toName, toEmail), content, content)

	// SendGrid has insane defaults, so turn all that stuff off.
	message.TrackingSettings = &mail.TrackingSettings{
		SubscriptionTracking: &mail.SubscriptionTrackingSetting{
			Enable: boolptr(false),
		},
		ClickTracking: &mail.ClickTrackingSetting{
			Enable: boolptr(false),
		},
		OpenTracking: &mail.OpenTrackingSetting{
			Enable: boolptr(false),
		},
		BypassListManagement: &mail.Setting{
			Enable: boolptr(true),
		},
	}

	message.MailSettings = &mail.MailSettings{
		BypassListManagement: &mail.Setting{
			Enable: boolptr(true),
		},
	}

	fmt.Printf("\nsending message: %#v\n", message.Content)
	// response, err := sendgridClient.Send(message)
	// if err != nil {
	// 	return err
	// }

	// log.Info().
	// 	Int("status", response.StatusCode).
	// 	Interface("headers", response.Headers).
	// 	Str("body", response.Body).Msg("delivered sendgrid email")
	return nil
}

// boolptr is a helper for setting boolean values when configuring SendGrid
// emails.
func boolptr(b bool) *bool {
	return &b
}
