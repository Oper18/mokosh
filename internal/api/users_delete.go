package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// DeleteUser permanently deletes a user account along with all photos and files it owns.
//
//	@Summary	delete a user account and all owned media
//	@Id			DeleteUser
//	@Tags		Users
//	@Produce	json
//	@Param		uid					path		string	true	"user uid"
//	@Success	200					{object}	i18n.Response
//	@Failure	400,401,403,404,429	{object}	i18n.Response
//	@Router		/api/v1/users/{uid} [delete]
func DeleteUser(router *gin.RouterGroup) {
	router.DELETE("/users/:uid", func(c *gin.Context) {
		conf := get.Config()

		// Account deletion requires authentication and enabled settings.
		if conf.Public() || conf.DisableSettings() {
			AbortForbidden(c)
			return
		}

		// Check if the session user is allowed to manage all accounts or delete his/her own account.
		s := AuthAny(c, acl.ResourceUsers, acl.Permissions{acl.ActionManage, acl.AccessOwn, acl.ActionDelete})

		if s.Abort(c) {
			return
		}

		clientIp := ClientIP(c)
		uid := clean.UID(c.Param("uid"))

		// Check if the current user has account management privileges for all accounts.
		isAdmin := acl.Rules.AllowAll(acl.ResourceUsers, s.GetUserRole(), acl.Permissions{acl.AccessAll, acl.ActionManage})

		var m *entity.User

		// Regular users may only delete their own account.
		if !isAdmin && s.GetUser().UserUID != uid {
			event.AuditErr([]string{clientIp, "session %s", "users", "delete", status.Denied}, s.RefID)
			AbortForbidden(c)
			return
		} else if s.GetUser().UserUID == uid {
			m = s.GetUser()
		} else if m = entity.FindUserByUID(uid); m == nil {
			Abort(c, http.StatusNotFound, i18n.ErrUserNotFound)
			return
		}

		// Protect the primary/system user from deletion.
		if m.ID <= 1 {
			event.AuditErr([]string{clientIp, "session %s", "users", m.UserName, "delete", status.Denied}, s.RefID)
			Abort(c, http.StatusForbidden, i18n.ErrForbidden)
			return
		}

		// Permanently delete all photos and files owned (created) by the user.
		if n, err := photoprism.DeleteUserPhotos(m.UserUID); err != nil {
			event.AuditErr([]string{clientIp, "session %s", "users", m.UserName, "delete media", status.Error(err)}, s.RefID)
		} else if n > 0 {
			event.AuditInfo([]string{clientIp, "session %s", "users", m.UserName, "delete media", status.Succeeded}, s.RefID)
		}

		// Permanently delete the user account, its sessions, and related records (also flushes the session cache).
		if err := m.DeletePermanently(); err != nil {
			event.AuditErr([]string{clientIp, "session %s", "users", m.UserName, "delete", status.Error(err)}, s.RefID)
			AbortDeleteFailed(c)
			return
		}

		// Log event.
		event.AuditInfo([]string{clientIp, "session %s", "users", m.UserName, "deleted", status.Succeeded}, s.RefID)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgPermanentlyDeleted))
	})
}
