package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/netlify/gotrue/crypto"
	"github.com/netlify/gotrue/mailer"
	"github.com/netlify/gotrue/models"
	"github.com/netlify/gotrue/storage"
	"github.com/pkg/errors"
)

func sendConfirmation(tx *storage.Connection, u *models.User, mailer mailer.Mailer, maxFrequency time.Duration, referrerURL string) error {
	if u.ConfirmationSentAt != nil && !u.ConfirmationSentAt.Add(maxFrequency).Before(time.Now()) {
		return nil
	}

	oldToken := u.ConfirmationToken
	u.ConfirmationToken = crypto.SecureToken()
	now := time.Now()
	if err := mailer.ConfirmationMail(u, referrerURL); err != nil {
		u.ConfirmationToken = oldToken
		return errors.Wrap(err, "Error sending confirmation email")
	}
	u.ConfirmationSentAt = &now
	return errors.Wrap(tx.UpdateOnly(u, "confirmation_token", "confirmation_sent_at"), "Database error updating user for confirmation")
}

func sendInvite(tx *storage.Connection, u *models.User, mailer mailer.Mailer, referrerURL string) error {
	oldToken := u.ConfirmationToken
	u.ConfirmationToken = crypto.SecureToken()
	now := time.Now()
	if err := mailer.InviteMail(u, referrerURL); err != nil {
		u.ConfirmationToken = oldToken
		return errors.Wrap(err, "Error sending invite email")
	}
	u.InvitedAt = &now
	return errors.Wrap(tx.UpdateOnly(u, "confirmation_token", "invited_at"), "Database error updating user for invite")
}

func (a *API) sendPasswordRecovery(tx *storage.Connection, u *models.User, mailer mailer.Mailer, maxFrequency time.Duration, referrerURL string) error {
	if u.RecoverySentAt != nil && !u.RecoverySentAt.Add(maxFrequency).Before(time.Now()) {
		totalTime := int((maxFrequency / time.Minute).Minutes())
		remainingDuration := maxFrequency - time.Since(*u.RecoverySentAt)
		remainingMinutes := int(remainingDuration.Minutes())
		remainingSeconds := int(remainingDuration.Seconds() - float64(remainingMinutes*60))

		jsonErr := map[string]string{
			"en": fmt.Sprintf("You cannot request more than one password recovery link within %d minutes. You must wait %d minutes and %d seconds.", totalTime, remainingMinutes, remainingSeconds),
			"tr": fmt.Sprintf("%d dakika içinde birden fazla şifremi unuttum linki talep edemezsiniz. Beklemeniz gereken süre: %d dakika %d saniyedir.", totalTime, remainingMinutes, remainingSeconds),
		}
		bytesErr, err := json.Marshal(jsonErr)
		if err != nil {
			return errors.Wrap(err, "Error marshalling error message")
		}
		return errors.New(string(bytesErr[:]))
	}

	oldToken := u.RecoveryToken
	u.RecoveryToken = crypto.SecureToken()
	now := time.Now()
	if err := mailer.RecoveryMail(u, referrerURL); err != nil {
		u.RecoveryToken = oldToken
		return errors.Wrap(err, "Error sending recovery email")
	}
	u.RecoverySentAt = &now
	return errors.Wrap(tx.UpdateOnly(u, "recovery_token", "recovery_sent_at"), "Database error updating user for recovery")
}

func (a *API) sendEmailChange(tx *storage.Connection, u *models.User, mailer mailer.Mailer, email string, referrerURL string) error {
	oldToken := u.EmailChangeToken
	oldEmail := u.EmailChange
	u.EmailChangeToken = crypto.SecureToken()
	u.EmailChange = email
	now := time.Now()
	if err := mailer.EmailChangeMail(u, referrerURL); err != nil {
		u.EmailChangeToken = oldToken
		u.EmailChange = oldEmail
		return err
	}

	u.EmailChangeSentAt = &now
	return errors.Wrap(tx.UpdateOnly(u, "email_change_token", "email_change", "email_change_sent_at"), "Database error updating user for email change")
}

func (a *API) validateEmail(ctx context.Context, email string) error {
	if email == "" {
		return unprocessableEntityError("An email address is required")
	}
	mailer := a.Mailer(ctx)
	if err := mailer.ValidateEmail(email); err != nil {
		return unprocessableEntityError("Unable to validate email address: " + err.Error())
	}
	return nil
}
