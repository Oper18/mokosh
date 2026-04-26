package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/react"
	"github.com/photoprism/photoprism/pkg/txt"
)

// authReact returns an authorized session for photo reactions.
// It accepts both role-based ActionReact and visitor sessions with PermReact on any share.
func authReact(c *gin.Context) *entity.Session {
	s := AuthAny(c, acl.ResourcePhotos, acl.Permissions{acl.ActionReact})
	if s.Valid() {
		return s
	}

	if vs := Session(ClientIP(c), AuthToken(c)); vs != nil && vs.IsVisitor() && vs.HasSharePerm(entity.PermReact) {
		return vs
	}

	return s
}

// LikePhoto flags a photo as favorite.
//
//	@Summary	flags a photo as favorite
//	@Id			LikePhoto
//	@Tags		Photos
//	@Accept		json
//	@Produce	json
//	@Success	200				{object}	gin.H
//	@Failure	401,403,404,500	{object}	i18n.Response
//	@Param		uid				path		string	true	"photo uid"
//	@Router		/api/v1/photos/{uid}/like [post]
func LikePhoto(router *gin.RouterGroup) {
	router.POST("/photos/:uid/like", func(c *gin.Context) {
		s := AuthAny(c, acl.ResourcePhotos, acl.Permissions{acl.ActionUpdate, acl.ActionReact})

		if s.Abort(c) {
			return
		}

		id := clean.UID(c.Param("uid"))
		m, err := query.PhotoByUID(id)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		if get.Config().Develop() && acl.Rules.Allow(acl.ResourcePhotos, s.GetUserRole(), acl.ActionReact) {
			logWarn("react", m.React(s.GetUser(), react.Find("love"), ""))
		}

		if acl.Rules.Allow(acl.ResourcePhotos, s.GetUserRole(), acl.ActionUpdate) {
			err = m.SetFavorite(true)

			if err != nil {
				log.Errorf("photo: %s", err.Error())
				AbortSaveFailed(c)
				return
			}

			SaveSidecarYaml(&m)
			PublishPhotoEvent(StatusUpdated, id, c)
		}

		c.JSON(http.StatusOK, gin.H{"photo": m})
	})
}

// DislikePhoto removes the favorite flags from a photo.
//
//	@Summary	removes the favorite flags from a photo
//	@Id			DislikePhoto
//	@Tags		Photos
//	@Accept		json
//	@Produce	json
//	@Success	200				{object}	gin.H
//	@Failure	401,403,404,500	{object}	i18n.Response
//	@Param		uid				path		string	true	"photo uid"
//	@Router		/api/v1/photos/{uid}/like [delete]
func DislikePhoto(router *gin.RouterGroup) {
	router.DELETE("/photos/:uid/like", func(c *gin.Context) {
		s := AuthAny(c, acl.ResourcePhotos, acl.Permissions{acl.ActionUpdate, acl.ActionReact})

		if s.Abort(c) {
			return
		}

		id := clean.UID(c.Param("uid"))
		m, err := query.PhotoByUID(id)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		if get.Config().Develop() && acl.Rules.Allow(acl.ResourcePhotos, s.GetUserRole(), acl.ActionReact) {
			logWarn("react", m.UnReact(s.GetUser()))
		}

		if acl.Rules.Allow(acl.ResourcePhotos, s.GetUserRole(), acl.ActionUpdate) {
			err = m.SetFavorite(false)

			if err != nil {
				log.Errorf("photo: %s", err.Error())
				AbortSaveFailed(c)
				return
			}

			SaveSidecarYaml(&m)
			PublishPhotoEvent(StatusUpdated, id, c)
		}

		c.JSON(http.StatusOK, gin.H{"photo": m})
	})
}

// GetPhotoReactions returns aggregated emoji counts, paginated reaction details, and the current user's reactions for a photo.
//
//	@Summary	returns reactions for a photo
//	@Id			GetPhotoReactions
//	@Tags		Photos
//	@Produce	json
//	@Success	200			{object}	gin.H
//	@Failure	401,403,404	{object}	i18n.Response
//	@Param		uid			path		string	true	"photo uid"
//	@Router		/api/v1/photos/{uid}/react [get]
func GetPhotoReactions(router *gin.RouterGroup) {
	router.GET("/photos/:uid/react", func(c *gin.Context) {
		s := authReact(c)

		if s.Abort(c) {
			return
		}

		id := clean.UID(c.Param("uid"))
		if id == "" {
			AbortEntityNotFound(c)
			return
		}

		const pageLimit = 10
		offset := txt.Int(c.Query("offset"))
		if offset < 0 {
			offset = 0
		}

		counts, err := query.PhotoReactionCounts(id)

		if err != nil {
			log.Errorf("reaction: failed to get counts for %s: %s", id, err)
			AbortUnexpectedError(c)
			return
		}

		details, err := query.PhotoReactionDetails(id, offset, pageLimit)

		if err != nil {
			log.Errorf("reaction: failed to get details for %s: %s", id, err)
			AbortUnexpectedError(c)
			return
		}

		total, err := query.PhotoReactionTotal(id)

		if err != nil {
			log.Errorf("reaction: failed to get total for %s: %s", id, err)
			AbortUnexpectedError(c)
			return
		}

		mine := entity.FindUserReactions(id, s.GetUser().GetUID())

		log.Debugf("reaction: found %d/%d reaction(s) for photo %s (offset %d)", len(details), total, id, offset)

		c.JSON(http.StatusOK, gin.H{
			"reactions": counts,
			"details":   details,
			"mine":      mine,
			"total":     total,
			"offset":    offset,
			"limit":     pageLimit,
		})
	})
}

// ReactPhoto adds a new emoji reaction and/or comment on a photo.
//
//	@Summary	adds a reaction to a photo
//	@Id			ReactPhoto
//	@Tags		Photos
//	@Accept		json
//	@Produce	json
//	@Success	200				{object}	gin.H
//	@Failure	400,401,403,404	{object}	i18n.Response
//	@Param		uid				path		string		true	"photo uid"
//	@Param		react			body		form.React	true	"optional emoji and/or comment (at least one required)"
//	@Router		/api/v1/photos/{uid}/react [post]
func ReactPhoto(router *gin.RouterGroup) {
	router.POST("/photos/:uid/react", func(c *gin.Context) {
		s := authReact(c)

		if s.Abort(c) {
			return
		}

		var frm form.React

		LimitRequestBodyBytes(c, MaxMutationRequestBytes)

		if err := c.BindJSON(&frm); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}

			AbortBadRequest(c, err)
			return
		}

		if frm.Emoji == "" && frm.Comment == "" {
			AbortBadRequest(c, fmt.Errorf("emoji or comment required"))
			return
		}

		id := clean.UID(c.Param("uid"))
		m, err := query.PhotoByUID(id)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		r := entity.NewReaction(m.PhotoUID, s.GetUser().GetUID()).
			WithEmoji(frm.Emoji).
			WithComment(frm.Comment)

		if err = r.Create(); err != nil {
			log.Errorf("reaction: failed to create for photo %s user %s: %s", m.PhotoUID, s.GetUser().GetUID(), err)
			AbortSaveFailed(c)
			return
		}

		log.Debugf("reaction: created id %d for photo %s user %s", r.ID, m.PhotoUID, s.GetUser().GetUID())

		c.JSON(http.StatusOK, gin.H{"reaction": r})
	})
}

// DeletePhotoReaction removes a specific reaction by ID, verifying ownership.
//
//	@Summary	removes a specific reaction from a photo
//	@Id			DeletePhotoReaction
//	@Tags		Photos
//	@Produce	json
//	@Success	200			{object}	gin.H
//	@Failure	401,403,404	{object}	i18n.Response
//	@Param		uid			path		string	true	"photo uid"
//	@Param		id			path		int		true	"reaction id"
//	@Router		/api/v1/photos/{uid}/react/{id} [delete]
func DeletePhotoReaction(router *gin.RouterGroup) {
	router.DELETE("/photos/:uid/react/:id", func(c *gin.Context) {
		s := authReact(c)

		if s.Abort(c) {
			return
		}

		rid, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || rid == 0 {
			AbortEntityNotFound(c)
			return
		}

		r := entity.FindReactionByID(uint(rid))
		if r == nil {
			AbortEntityNotFound(c)
			return
		}

		if r.UserUID != s.GetUser().GetUID() {
			AbortForbidden(c)
			return
		}

		if err = r.Delete(); err != nil {
			log.Errorf("reaction: failed to delete id %d: %s", rid, err)
			AbortDeleteFailed(c)
			return
		}

		log.Debugf("reaction: deleted id %d", rid)

		c.JSON(http.StatusOK, gin.H{"id": rid})
	})
}
