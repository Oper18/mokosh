package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/txt"
)

// UpdateLink updates a share link and return it as JSON.
//
// PUT /api/v1/:entity/:uid/links/:link
func UpdateLink(c *gin.Context) {
	s := Auth(c, acl.ResourceShares, acl.ActionUpdate)

	if s.Invalid() {
		AbortForbidden(c)
		return
	}

	var frm form.Link

	// Assign and validate request form values.
	LimitRequestBodyBytes(c, MaxMutationRequestBytes)

	if err := c.BindJSON(&frm); err != nil {
		if IsRequestBodyTooLarge(err) {
			AbortRequestTooLarge(c, i18n.ErrBadRequest)
			return
		}

		AbortBadRequest(c, err)
		return
	}

	link := entity.FindLink(clean.Token(c.Param("link")))

	link.LinkName = clean.Name(frm.LinkName)
	link.SetSlug(frm.ShareSlug)
	link.MaxViews = frm.MaxViews
	link.LinkExpires = frm.LinkExpires
	link.Perm = frm.Perm

	if frm.LinkToken != "" {
		link.LinkToken = strings.TrimSpace(strings.ToLower(frm.LinkToken))
	}

	if frm.Password != "" {
		if err := link.SetPassword(frm.Password); err != nil {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}
	}

	if err := link.Save(); err != nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": txt.UpperFirst(err.Error())})
		return
	}

	syncLinkNameToVisitor(link)

	UpdateClientConfig()

	PublishAlbumEvent(StatusUpdated, link.ShareUID, c)

	c.JSON(http.StatusOK, link)
}

// syncLinkNameToVisitor keeps auth_users.display_name in sync with link.LinkName.
// If no visitor user exists for the link yet and a name is set, one is created.
func syncLinkNameToVisitor(link *entity.Link) {
	if link == nil {
		return
	}

	var share entity.UserShare
	err := entity.Db().Where("link_uid = ?", link.LinkUID).First(&share).Error

	if err != nil {
		// No UserShare for this link yet.
		if link.LinkName == "" {
			return
		}
		// Create a dedicated visitor user for this link.
		visitor := entity.NewUser()
		visitor.UserRole = acl.RoleVisitor.String()
		visitor.CanLogin = false
		visitor.DisplayName = link.LinkName
		if createErr := visitor.Create(); createErr != nil {
			log.Errorf("link: failed to create visitor user (%s)", createErr)
			return
		}
		newShare := entity.NewUserShare(visitor.UserUID, link.ShareUID, link.Perm, nil)
		newShare.LinkUID = link.LinkUID
		if saveErr := newShare.Save(); saveErr != nil {
			log.Errorf("link: failed to create user share for visitor (%s)", saveErr)
		}
		return
	}

	// Update the display name of the already-linked visitor user.
	u := entity.FindUserByUID(share.UserUID)
	if u == nil || !u.IsVisitor() {
		return
	}

	if err := entity.Db().Model(u).UpdateColumn("display_name", link.LinkName).Error; err != nil {
		log.Errorf("link: failed to sync display name to visitor user %s (%s)", u.UserUID, err)
	}
}

// DeleteLink deletes a share link.
//
// DELETE /api/v1/:entity/:uid/links/:link
func DeleteLink(c *gin.Context) {
	s := Auth(c, acl.ResourceShares, acl.ActionDelete)

	if s.Invalid() {
		AbortForbidden(c)
		return
	}

	link := entity.FindLink(clean.Token(c.Param("link")))

	if err := link.Delete(); err != nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": txt.UpperFirst(err.Error())})
		return
	}

	UpdateClientConfig()

	PublishAlbumEvent(StatusUpdated, link.ShareUID, c)

	c.JSON(http.StatusOK, link)
}

// CreateLink adds a new share link and returns it as JSON.
// Note: Internal helper used by resource-specific endpoints (e.g., albums, photos).
// Swagger annotations are defined on those public handlers to avoid generating
// undocumented generic paths like "/api/v1/{entity}/{uid}/links".
func CreateLink(c *gin.Context) {
	s := Auth(c, acl.ResourceShares, acl.ActionCreate)

	if s.Abort(c) {
		return
	}

	uid := clean.UID(c.Param("uid"))

	if uid == "" {
		AbortBadRequest(c)
		return
	}

	var frm form.Link

	LimitRequestBodyBytes(c, MaxMutationRequestBytes)

	if err := c.BindJSON(&frm); err != nil {
		if IsRequestBodyTooLarge(err) {
			AbortRequestTooLarge(c, i18n.ErrBadRequest)
			return
		}

		AbortBadRequest(c, err)
		return
	}

	link := entity.NewUserLink(uid, s.UserUID)

	link.LinkName = clean.Name(frm.LinkName)
	link.SetSlug(frm.ShareSlug)
	link.MaxViews = frm.MaxViews
	link.LinkExpires = frm.LinkExpires
	if frm.Perm != 0 {
		link.Perm = frm.Perm
	} else {
		link.Perm = entity.PermView
	}

	if frm.Password != "" {
		if err := link.SetPassword(frm.Password); err != nil {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}
	}

	if err := link.Save(); err != nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": txt.UpperFirst(err.Error())})
		return
	}

	UpdateClientConfig()

	PublishAlbumEvent(StatusUpdated, link.ShareUID, c)

	c.JSON(http.StatusOK, link)
}

// CreateAlbumLink adds a new album share link and return it as JSON.
//
//	@Summary	adds a new album share link and return it as JSON
//	@Id			CreateAlbumLink
//	@Tags		Links, Albums
//	@Accept		json
//	@Produce	json
//	@Success	200						{object}	entity.Link
//	@Failure	400,401,403,404,409,429	{object}	i18n.Response
//	@Param		uid						path		string		true	"album uid"
//	@Param		link					body		form.Link	true	"link properties (currently supported: slug, expires)"
//	@Router		/api/v1/albums/{uid}/links [post]
func CreateAlbumLink(router *gin.RouterGroup) {
	router.POST("/albums/:uid/links", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		if _, err := query.AlbumByUID(clean.UID(c.Param("uid"))); err != nil {
			AbortAlbumNotFound(c)
			return
		}

		CreateLink(c)
	})
}

// UpdateAlbumLink updates an album share link and return it as JSON.
//
//	@Summary	updates an album share link and return it as JSON
//	@Id			UpdateAlbumLink
//	@Tags		Links, Albums
//	@Accept		json
//	@Produce	json
//	@Success	200						{object}	entity.Link
//	@Failure	400,401,403,429,409,500	{object}	i18n.Response
//	@Param		uid						path		string		true	"album uid"
//	@Param		linkuid					path		string		true	"link uid"
//	@Param		link					body		form.Link	true	"properties to be updated (currently supported: slug, expires, token)"
//	@Router		/api/v1/albums/{uid}/links/{linkuid} [put]
func UpdateAlbumLink(router *gin.RouterGroup) {
	router.PUT("/albums/:uid/links/:link", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		UpdateLink(c)
	})
}

// DeleteAlbumLink deletes an album share link.
//
//	@Summary	deletes an album share link
//	@Id			DeleteAlbumLink
//	@Tags		Links, Albums
//	@Accept		json
//	@Produce	json
//	@Success	200				{object}	entity.Link
//	@Failure	401,403,429,409	{object}	i18n.Response
//	@Param		uid				path		string	true	"album"
//	@Param		linkuid			path		string	true	"link uid"
//	@Router		/api/v1/albums/{uid}/links/{linkuid} [delete]
func DeleteAlbumLink(router *gin.RouterGroup) {
	router.DELETE("/albums/:uid/links/:link", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		DeleteLink(c)
	})
}

// GetAlbumLinks returns all share links for the given UID as JSON.
//
//	@Summary	returns all share links for the given UID as JSON
//	@Id			GetAlbumLinks
//	@Tags		Links, Albums
//	@Produce	json
//	@Success	200				{object}	entity.Link
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		uid				path		string	true	"album uid"
//	@Router		/api/v1/albums/{uid}/links [get]
func GetAlbumLinks(router *gin.RouterGroup) {
	router.GET("/albums/:uid/links", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		m, err := query.AlbumByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		c.JSON(http.StatusOK, m.Links())
	})
}

// CreatePhotoLink adds a new photo share link and returns it as JSON.
//
//	@Summary	adds a new photo share link and returns it as JSON
//	@Id			CreatePhotoLink
//	@Tags		Links, Photos
//	@Accept		json
//	@Produce	json
//	@Success	200						{object}	entity.Link
//	@Failure	400,401,403,404,409,429	{object}	i18n.Response
//	@Param		uid						path		string		true	"photo uid"
//	@Param		link					body		form.Link	true	"link properties"
//	@Router		/api/v1/photos/{uid}/links [post]
func CreatePhotoLink(router *gin.RouterGroup) {
	router.POST("/photos/:uid/links", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		if _, err := query.PhotoByUID(clean.UID(c.Param("uid"))); err != nil {
			AbortEntityNotFound(c)
			return
		}

		CreateLink(c)
	})
}

// UpdatePhotoLink updates an existing photo share link and returns it as JSON.
//
//	@Summary	updates an existing photo share link and returns it as JSON
//	@Id			UpdatePhotoLink
//	@Tags		Links, Photos
//	@Accept		json
//	@Produce	json
//	@Success	200						{object}	entity.Link
//	@Failure	400,401,403,409,429,500	{object}	i18n.Response
//	@Param		uid						path		string		true	"photo uid"
//	@Param		linkuid					path		string		true	"link uid"
//	@Param		link					body		form.Link	true	"properties to update"
//	@Router		/api/v1/photos/{uid}/links/{linkuid} [put]
func UpdatePhotoLink(router *gin.RouterGroup) {
	router.PUT("/photos/:uid/links/:link", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		UpdateLink(c)
	})
}

// DeletePhotoLink deletes a photo share link.
//
//	@Summary	deletes a photo share link
//	@Id			DeletePhotoLink
//	@Tags		Links, Photos
//	@Produce	json
//	@Success	200				{object}	entity.Link
//	@Failure	401,403,409,429	{object}	i18n.Response
//	@Param		uid				path		string	true	"photo uid"
//	@Param		linkuid			path		string	true	"link uid"
//	@Router		/api/v1/photos/{uid}/links/{linkuid} [delete]
func DeletePhotoLink(router *gin.RouterGroup) {
	router.DELETE("/photos/:uid/links/:link", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		DeleteLink(c)
	})
}

// GetPhotoLinks returns all share links for the given photo UID as JSON.
//
//	@Summary	returns all share links for the given photo UID as JSON
//	@Id			GetPhotoLinks
//	@Tags		Links, Photos
//	@Produce	json
//	@Success	200				{object}	entity.Link
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		uid				path		string	true	"photo uid"
//	@Router		/api/v1/photos/{uid}/links [get]
func GetPhotoLinks(router *gin.RouterGroup) {
	router.GET("/photos/:uid/links", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		m, err := query.PhotoByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, m.Links())
	})
}

/*

// CreateLabelLink adds a new label share link and return it as JSON.
//
//	@Tags 		Links, Labels
//	@Router		/api/v1/labels/{uid}/links [post]
func CreateLabelLink(router *gin.RouterGroup) {
	router.POST("/labels/:uid/links", func(c *gin.Context) {
		s := Auth(c, acl.ResourceLabels, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		if _, err := query.LabelByUID(clean.UID(c.Param("uid"))); err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrLabelNotFound)
			return
		}

		CreateLink(c)
	})
}

// UpdateLabelLink updates a label share link and return it as JSON.
//
// PUT /api/v1/labels/:uid/links/:link
func UpdateLabelLink(router *gin.RouterGroup) {
	router.PUT("/labels/:uid/links/:link", func(c *gin.Context) {
		s := Auth(c, acl.ResourceLabels, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		UpdateLink(c)
	})
}

// DeleteLabelLink deletes a label share link.
//
// DELETE /api/v1/labels/:uid/links/:link
func DeleteLabelLink(router *gin.RouterGroup) {
	router.DELETE("/labels/:uid/links/:link", func(c *gin.Context) {
		s := Auth(c, acl.ResourceLabels, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		DeleteLink(c)
	})
}

// GetLabelLinks returns all share links for the given UID as JSON.
//
// GET /api/v1/labels/:uid/links
func GetLabelLinks(router *gin.RouterGroup) {
	router.GET("/labels/:uid/links", func(c *gin.Context) {
		s := Auth(c, acl.ResourceLabels, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		m, err := query.LabelByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		c.JSON(http.StatusOK, m.Links())
	})
}
*/
