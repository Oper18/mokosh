package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
)

// GetShares returns all UserShare records for resources owned by the authenticated user.
//
//	@Summary	returns user shares for resources owned by the authenticated user
//	@Id			GetShares
//	@Tags		Shares
//	@Produce	json
//	@Success	200			{array}		entity.UserShare
//	@Failure	401,403,500	{object}	i18n.Response
//	@Router		/api/v1/shares [get]
func GetShares(router *gin.RouterGroup) {
	router.GET("/shares", func(c *gin.Context) {
		s := Auth(c, acl.ResourceShares, acl.ActionView)

		if s.Abort(c) {
			return
		}

		shares, err := query.UserSharesByOwner(s.UserUID)

		if err != nil {
			log.Errorf("shares: %s (list)", err)
			AbortUnexpectedError(c)
			return
		}

		c.JSON(http.StatusOK, shares)
	})
}

// UpdateSharePerm updates the permission, expiry, and comment of a specific UserShare.
// The authenticated user must own the resource referenced by shareUID.
//
//	@Summary	updates permission for a user share
//	@Id			UpdateSharePerm
//	@Tags		Shares
//	@Accept		json
//	@Produce	json
//	@Success	200					{object}	entity.UserShare
//	@Failure	400,401,403,404,500	{object}	i18n.Response
//	@Param		shareUID			path		string			true	"resource UID (album or photo)"
//	@Param		userUID				path		string			true	"user UID of the share recipient"
//	@Param		perm				body		form.SharePerm	true	"updated permission"
//	@Router		/api/v1/shares/{shareUID}/users/{userUID} [put]
func UpdateSharePerm(router *gin.RouterGroup) {
	router.PUT("/shares/:shareUID/users/:userUID", func(c *gin.Context) {
		s := Auth(c, acl.ResourceShares, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		shareUID := clean.UID(c.Param("shareUID"))
		userUID := clean.UID(c.Param("userUID"))

		// Admins can update any share; others must own the resource.
		var share *entity.UserShare

		if s.GetUser().HasRole(acl.RoleAdmin) {
			share = entity.FindUserShare(entity.UserShare{UserUID: userUID, ShareUID: shareUID})
		} else {
			share = query.UserShareOwned(shareUID, userUID, s.UserUID)
		}

		if share == nil {
			AbortEntityNotFound(c)
			return
		}

		var frm form.SharePerm

		LimitRequestBodyBytes(c, MaxMutationRequestBytes)

		if err := c.BindJSON(&frm); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, 0)
				return
			}
			AbortBadRequest(c, err)
			return
		}

		values := entity.Values{
			"perm":       frm.Perm,
			"expires_at": frm.ExpiresAt,
			"comment":    frm.Comment,
		}

		if err := share.Updates(values); err != nil {
			log.Errorf("shares: %s (update perm)", err)
			AbortSaveFailed(c)
			return
		}

		share.Perm = frm.Perm
		share.ExpiresAt = frm.ExpiresAt
		share.Comment = frm.Comment

		c.JSON(http.StatusOK, share)
	})
}
