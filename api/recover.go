package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/netlify/gotrue/models"
	"github.com/netlify/gotrue/storage"
)

// RecoverParams holds the parameters for a password recovery request
type RecoverParams struct {
	Email string `json:"email"`
}

// Recover sends a recovery email
func (a *API) Recover(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	config := a.getConfig(ctx)
	instanceID := getInstanceID(ctx)
	params := &RecoverParams{}
	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(params)
	if err != nil {
		return badRequestError("Could not read verification params: %v", err)
	}

	if params.Email == "" {
		return unprocessableEntityError("Password recovery requires an email")
	}

	aud := a.requestAud(ctx, r)
	user, err := models.FindUserByEmailAndAudience(a.db, instanceID, params.Email, aud)
	if err != nil {
		if models.IsNotFoundError(err) {
			// return notFoundError(err.Error())
			return sendJSON(w, http.StatusOK, &map[string]string{})
		}
		return internalServerError("Database error finding user").WithInternalError(err)
	}

	maxFrequency := config.SMTP.MaxFrequency
	if user.RecoverySentAt != nil && !user.RecoverySentAt.Add(maxFrequency).Before(time.Now()) {
		totalTime := int(maxFrequency.Minutes())
		remainingDuration := maxFrequency - time.Since(*user.RecoverySentAt)
		remainingMinutes := int(remainingDuration.Minutes())
		remainingSeconds := int(remainingDuration.Seconds() - float64(remainingMinutes*60))

		jsonErr := map[string]string{
			"en": fmt.Sprintf("You cannot request more than one password recovery link within %d minutes. You must wait %d minutes and %d seconds.", totalTime, remainingMinutes, remainingSeconds),
			"tr": fmt.Sprintf("%d dakika içinde birden fazla şifremi unuttum linki talep edemezsiniz. Beklemeniz gereken süre: %d dakika %d saniyedir.", totalTime, remainingMinutes, remainingSeconds),
		}
		bytesErr, err := json.Marshal(jsonErr)
		if err != nil {
			return unprocessableEntityError("Error marshalling error message")
		}
		return tooEarlyError(string(bytesErr[:]))
	}

	err = a.db.Transaction(func(tx *storage.Connection) error {
		if terr := models.NewAuditLogEntry(tx, instanceID, user, models.UserRecoveryRequestedAction, nil); terr != nil {
			return terr
		}

		mailer := a.Mailer(ctx)
		referrer := a.getReferrer(r)
		return a.sendPasswordRecovery(tx, user, mailer, config.SMTP.MaxFrequency, referrer)
	})
	if err != nil {
		return internalServerError("Error recovering user").WithInternalError(err)
	}

	return sendJSON(w, http.StatusOK, &map[string]string{})
}
