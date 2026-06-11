package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/workers"
	"github.com/photoprism/photoprism/pkg/clean"
)

// CreateAlbumInvite creates a visitor user, a share link, and a user share for the given album.
//
//	@Summary	creates a visitor user with a share link for the given album
//	@Id			CreateAlbumInvite
//	@Tags		Invites, Albums
//	@Accept		json
//	@Produce	json
//	@Success	201					{object}	gin.H
//	@Failure	400,401,403,404,500	{object}	i18n.Response
//	@Param		uid					path		string		true	"album uid"
//	@Param		invite				body		form.Invite	true	"invite properties"
//	@Router		/api/v1/albums/{uid}/invites [post]
func CreateAlbumInvite(router *gin.RouterGroup) {
	router.POST("/albums/:uid/invites", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionShare)

		if s.Abort(c) {
			return
		}

		albumUID := clean.UID(c.Param("uid"))

		if _, err := query.AlbumByUID(albumUID); err != nil {
			AbortAlbumNotFound(c)
			return
		}

		var frm form.Invite

		LimitRequestBodyBytes(c, MaxMutationRequestBytes)

		if err := c.BindJSON(&frm); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, 0)
				return
			}

			AbortBadRequest(c, err)
			return
		}

		// Create a visitor user with no login capability.
		visitorName := clean.Name(frm.Name)
		visitor := entity.NewUser()
		visitor.UserRole = acl.RoleVisitor.String()
		visitor.CanLogin = false
		visitor.InviteToken = entity.GenerateToken()
		visitor.DisplayName = visitorName

		if err := visitor.Create(); err != nil {
			log.Errorf("invite: failed to create visitor user (%s)", err)
			AbortUnexpectedError(c)
			return
		}

		// Create a share link for the album owned by the requesting user.
		link := entity.NewUserLink(albumUID, s.UserUID)
		link.LinkName = visitorName
		link.Perm = frm.Perm
		link.Comment = frm.Comment

		if err := link.Save(); err != nil {
			log.Errorf("invite: failed to create link (%s)", err)
			AbortUnexpectedError(c)
			return
		}

		// Create a user share connecting the visitor to the album via the link.
		share := entity.NewUserShare(visitor.UserUID, albumUID, frm.Perm, frm.ExpiresAt)
		share.LinkUID = link.LinkUID
		share.Comment = frm.Comment

		if err := share.Save(); err != nil {
			log.Errorf("invite: failed to create user share (%s)", err)
			AbortUnexpectedError(c)
			return
		}

		// If the permission level is PermComment or below, generate degraded mini
		// copies of all media in the album asynchronously.
		if frm.Perm <= entity.PermComment {
			ownerName := s.GetUser().UserName
			go func() {
				m := workers.NewMini(get.Config())
				if err := m.ProcessShare(albumUID, ownerName); err != nil {
					log.Errorf("invite: mini worker failed (%s)", err)
				}
			}()
		}

		c.JSON(http.StatusCreated, gin.H{
			"LinkToken":   link.LinkToken,
			"InviteToken": visitor.InviteToken,
			"UserUID":     visitor.UserUID,
			"ShareUID":    albumUID,
			"Perm":        frm.Perm,
			"ExpiresAt":   frm.ExpiresAt,
		})
	})
}
