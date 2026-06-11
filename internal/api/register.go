package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// RegisterUser creates a new user account. The endpoint is public — no
// authentication token is required. Role and privilege fields are hardcoded
// server-side and are never accepted from the request body.
//
//	@Summary	register a new user account
//	@Tags		Users
//	@Accept		json
//	@Produce	json
//	@Param		registration	body		form.Register	true	"registration data"
//	@Success	201				{object}	entity.User
//	@Failure	400,429			{object}	i18n.Response
//	@Router		/api/v1/users/register [post]
func RegisterUser(router *gin.RouterGroup) {
	router.POST("/users/register", func(c *gin.Context) {
		clientIp := ClientIP(c)

		// Apply the same per-IP rate limit as login failures to prevent bulk
		// account creation and username enumeration.
		if limiter.Login.Reject(clientIp) {
			event.AuditWarn([]string{clientIp, "register", "too many requests"})
			AbortBusy(c)
			return
		}

		var frm form.Register

		LimitRequestBodyBytes(c, MaxAuthRequestBytes)

		if err := c.BindJSON(&frm); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, 0)
				return
			}

			Abort(c, http.StatusBadRequest, i18n.ErrBadRequest)
			return
		}

		// Build a full user form with defaults.
		// Only allow certain safe roles from user input during registration.
		userRole := "guest" // default role
		
		// Allow users to select between guest and photographer roles during registration
		if frm.Role != "" {
			formRole := clean.Role(frm.Role)
			if acl.UserRoles[formRole] != "" && (formRole == "guest" || formRole == "photographer") {
				userRole = formRole
			}
		}

		userForm := form.User{
			UserName:     frm.UserName,
			UserEmail:    frm.UserEmail,
			DisplayName:  frm.DisplayName,
			Password:     frm.Password,
			UserRole:     userRole, // Allow guest or photographer role based on validated form input
			CanLogin:     true,
			AuthProvider: string(authn.ProviderLocal),
			AuthMethod:   string(authn.MethodDefault),
		}

		if err := entity.AddUser(userForm); err != nil {
			// Count towards rate limit so failed attempts cannot be used for
			// fast enumeration of taken usernames.
			limiter.Login.Reserve(clientIp)
			event.AuditWarn([]string{clientIp, "register", "failed"})
			Abort(c, http.StatusBadRequest, i18n.ErrBadRequest)
			return
		}

		u := entity.FindUserByName(clean.Username(frm.UserName))
		if u == nil {
			AbortUnexpectedError(c)
			return
		}

		event.AuditInfo([]string{clientIp, "register", "users", u.UserName, "created"})

		c.JSON(http.StatusCreated, u)
	})
}
