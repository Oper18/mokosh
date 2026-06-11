package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// GetTeams returns all teams owned by or visible to the authenticated user.
//
//	@Summary	returns teams for the authenticated user
//	@Id			GetTeams
//	@Tags		Teams
//	@Produce	json
//	@Success	200		{array}		entity.Team
//	@Failure	401,403	{object}	i18n.Response
//	@Router		/api/v1/teams [get]
func GetTeams(router *gin.RouterGroup) {
	router.GET("/teams", func(c *gin.Context) {
		s := Auth(c, acl.ResourceTeams, acl.ActionView)

		if s.Abort(c) {
			return
		}

		var teams entity.Teams

		q := entity.Db().Order("team_name ASC")

		if !s.GetUser().HasRole(acl.RoleAdmin) {
			q = q.Where("user_uid = ?", s.UserUID)
		}

		if err := q.Find(&teams).Error; err != nil {
			log.Errorf("team: %s (list)", err)
			AbortUnexpectedError(c)
			return
		}

		c.JSON(http.StatusOK, teams)
	})
}

// GetTeam returns the team with the given UID as JSON.
//
//	@Summary	returns a team
//	@Id			GetTeam
//	@Tags		Teams
//	@Produce	json
//	@Success	200			{object}	entity.Team
//	@Failure	401,403,404	{object}	i18n.Response
//	@Param		uid			path		string	true	"Team UID"
//	@Router		/api/v1/teams/{uid} [get]
func GetTeam(router *gin.RouterGroup) {
	router.GET("/teams/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourceTeams, acl.ActionView)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))

		team := entity.FindTeam(uid)

		if team == nil {
			AbortEntityNotFound(c)
			return
		}

		if !s.GetUser().HasRole(acl.RoleAdmin) && team.UserUID != s.UserUID {
			AbortForbidden(c)
			return
		}

		c.JSON(http.StatusOK, team)
	})
}

// UpdateTeam updates the name of an existing team.
//
//	@Summary	updates a team
//	@Id			UpdateTeam
//	@Tags		Teams
//	@Accept		json
//	@Produce	json
//	@Success	200					{object}	entity.Team
//	@Failure	400,401,403,404,500	{object}	i18n.Response
//	@Param		uid					path		string		true	"Team UID"
//	@Param		team				body		form.Team	true	"updated team properties"
//	@Router		/api/v1/teams/{uid} [put]
func UpdateTeam(router *gin.RouterGroup) {
	router.PUT("/teams/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourceTeams, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))

		team := entity.FindTeam(uid)

		if team == nil {
			AbortEntityNotFound(c)
			return
		}

		if !s.GetUser().HasRole(acl.RoleAdmin) && team.UserUID != s.UserUID {
			AbortForbidden(c)
			return
		}

		var frm form.Team

		LimitRequestBodyBytes(c, MaxMutationRequestBytes)

		if err := c.BindJSON(&frm); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}
			AbortBadRequest(c, err)
			return
		}

		if frm.TeamName == "" {
			AbortBadRequest(c)
			return
		}

		team.TeamName = frm.TeamName

		if err := team.Save(); err != nil {
			log.Errorf("team: %s (update)", err)
			AbortSaveFailed(c)
			return
		}

		c.JSON(http.StatusOK, team)
	})
}

// CreateTeam creates a new team owned by the authenticated user.
//
//	@Summary	creates a new team
//	@Id			CreateTeam
//	@Tags		Teams
//	@Accept		json
//	@Produce	json
//	@Success	201					{object}	entity.Team
//	@Failure	400,401,403,429,500	{object}	i18n.Response
//	@Param		team				body		form.Team	true	"team name"
//	@Router		/api/v1/teams [post]
func CreateTeam(router *gin.RouterGroup) {
	router.POST("/teams", func(c *gin.Context) {
		s := Auth(c, acl.ResourceTeams, acl.ActionCreate)

		if s.Abort(c) {
			return
		}

		var frm form.Team

		LimitRequestBodyBytes(c, MaxMutationRequestBytes)

		if err := c.BindJSON(&frm); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}
			AbortBadRequest(c, err)
			return
		}

		if frm.TeamName == "" {
			AbortBadRequest(c)
			return
		}

		team := entity.NewTeam(frm.TeamName, s.UserUID)

		if err := team.Create(); err != nil {
			log.Errorf("team: %s (create)", err)
			AbortUnexpectedError(c)
			return
		}

		c.JSON(http.StatusCreated, team)
	})
}

// DeleteTeam permanently deletes a team and all its memberships and album attachments.
//
//	@Summary	deletes a team
//	@Id			DeleteTeam
//	@Tags		Teams
//	@Produce	json
//	@Success	200				{object}	entity.Team
//	@Failure	401,403,404,500	{object}	i18n.Response
//	@Param		uid				path		string	true	"Team UID"
//	@Router		/api/v1/teams/{uid} [delete]
func DeleteTeam(router *gin.RouterGroup) {
	router.DELETE("/teams/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourceTeams, acl.ActionDelete)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))

		team := entity.FindTeam(uid)

		if team == nil {
			AbortEntityNotFound(c)
			return
		}

		// Non-admin users can only delete their own teams.
		if !s.GetUser().HasRole(acl.RoleAdmin) && team.UserUID != s.UserUID {
			AbortForbidden(c)
			return
		}

		if err := team.Delete(); err != nil {
			log.Errorf("team: %s (delete)", err)
			AbortDeleteFailed(c)
			return
		}

		c.JSON(http.StatusOK, team)
	})
}

// AddUserToTeam adds a user to a team.
//
//	@Summary	adds a user to a team
//	@Id			AddUserToTeam
//	@Tags		Teams
//	@Accept		json
//	@Produce	json
//	@Success	200					{object}	entity.TeamUser
//	@Failure	400,401,403,404,500	{object}	i18n.Response
//	@Param		uid					path		string			true	"Team UID"
//	@Param		member				body		form.TeamUser	true	"user to add"
//	@Router		/api/v1/teams/{uid}/users [post]
func AddUserToTeam(router *gin.RouterGroup) {
	router.POST("/teams/:uid/users", func(c *gin.Context) {
		s := Auth(c, acl.ResourceTeams, acl.ActionManageOwn)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))

		team := entity.FindTeam(uid)

		if team == nil {
			AbortEntityNotFound(c)
			return
		}

		// Non-admin users can only manage their own teams.
		if !s.GetUser().HasRole(acl.RoleAdmin) && team.UserUID != s.UserUID {
			AbortForbidden(c)
			return
		}

		var frm form.TeamUser

		LimitRequestBodyBytes(c, MaxMutationRequestBytes)

		if err := c.BindJSON(&frm); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}
			AbortBadRequest(c, err)
			return
		}

		if !rnd.IsUID(frm.UserUID, entity.UserUID) {
			AbortBadRequest(c)
			return
		}

		// Verify the target user exists.
		if entity.FindUserByUID(frm.UserUID) == nil {
			AbortEntityNotFound(c)
			return
		}

		member := entity.NewTeamUser(team.TeamUID, frm.UserUID)

		if result := entity.FirstOrCreateTeamUser(member); result == nil {
			AbortUnexpectedError(c)
			return
		} else {
			c.JSON(http.StatusOK, result)
		}
	})
}

// RemoveUserFromTeam removes a user from a team.
//
//	@Summary	removes a user from a team
//	@Id			RemoveUserFromTeam
//	@Tags		Teams
//	@Produce	json
//	@Success	200				{object}	gin.H
//	@Failure	401,403,404,500	{object}	i18n.Response
//	@Param		uid				path		string	true	"Team UID"
//	@Param		userUID			path		string	true	"User UID"
//	@Router		/api/v1/teams/{uid}/users/{userUID} [delete]
func RemoveUserFromTeam(router *gin.RouterGroup) {
	router.DELETE("/teams/:uid/users/:userUID", func(c *gin.Context) {
		s := Auth(c, acl.ResourceTeams, acl.ActionManageOwn)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))
		userUID := clean.UID(c.Param("userUID"))

		team := entity.FindTeam(uid)

		if team == nil {
			AbortEntityNotFound(c)
			return
		}

		// Non-admin users can only manage their own teams.
		if !s.GetUser().HasRole(acl.RoleAdmin) && team.UserUID != s.UserUID {
			AbortForbidden(c)
			return
		}

		m := entity.TeamUser{}

		if err := entity.Db().Where("team_uid = ? AND user_uid = ?", team.TeamUID, userUID).First(&m).Error; err != nil {
			AbortEntityNotFound(c)
			return
		}

		if err := m.Delete(); err != nil {
			log.Errorf("team: %s (remove user)", err)
			AbortDeleteFailed(c)
			return
		}

		c.JSON(http.StatusOK, gin.H{"team_uid": team.TeamUID, "user_uid": userUID})
	})
}

// AddTeamToAlbum grants a team access to an album.
//
//	@Summary	grants a team access to an album
//	@Id			AddTeamToAlbum
//	@Tags		Teams
//	@Accept		json
//	@Produce	json
//	@Success	200					{object}	entity.TeamAlbum
//	@Failure	400,401,403,404,500	{object}	i18n.Response
//	@Param		uid					path		string			true	"Team UID"
//	@Param		album				body		form.TeamAlbum	true	"album to attach"
//	@Router		/api/v1/teams/{uid}/albums [post]
func AddTeamToAlbum(router *gin.RouterGroup) {
	router.POST("/teams/:uid/albums", func(c *gin.Context) {
		s := Auth(c, acl.ResourceTeams, acl.ActionManageOwn)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))

		team := entity.FindTeam(uid)

		if team == nil {
			AbortEntityNotFound(c)
			return
		}

		// Non-admin users can only manage their own teams.
		if !s.GetUser().HasRole(acl.RoleAdmin) && team.UserUID != s.UserUID {
			AbortForbidden(c)
			return
		}

		var frm form.TeamAlbum

		LimitRequestBodyBytes(c, MaxMutationRequestBytes)

		if err := c.BindJSON(&frm); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}
			AbortBadRequest(c, err)
			return
		}

		if !rnd.IsUID(frm.AlbumUID, entity.AlbumUID) {
			AbortBadRequest(c)
			return
		}

		// Verify the album exists.
		if _, err := query.AlbumByUID(frm.AlbumUID); err != nil {
			AbortEntityNotFound(c)
			return
		}

		ta := entity.NewTeamAlbum(team.TeamUID, frm.AlbumUID)

		if result := entity.FirstOrCreateTeamAlbum(ta); result == nil {
			AbortUnexpectedError(c)
			return
		} else {
			c.JSON(http.StatusOK, result)
		}
	})
}

// RemoveTeamFromAlbum removes a team's access to an album.
//
//	@Summary	removes a team's access to an album
//	@Id			RemoveTeamFromAlbum
//	@Tags		Teams
//	@Produce	json
//	@Success	200				{object}	gin.H
//	@Failure	401,403,404,500	{object}	i18n.Response
//	@Param		uid				path		string	true	"Team UID"
//	@Param		albumUID		path		string	true	"Album UID"
//	@Router		/api/v1/teams/{uid}/albums/{albumUID} [delete]
func RemoveTeamFromAlbum(router *gin.RouterGroup) {
	router.DELETE("/teams/:uid/albums/:albumUID", func(c *gin.Context) {
		s := Auth(c, acl.ResourceTeams, acl.ActionManageOwn)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))
		albumUID := clean.UID(c.Param("albumUID"))

		team := entity.FindTeam(uid)

		if team == nil {
			AbortEntityNotFound(c)
			return
		}

		// Non-admin users can only manage their own teams.
		if !s.GetUser().HasRole(acl.RoleAdmin) && team.UserUID != s.UserUID {
			AbortForbidden(c)
			return
		}

		m := entity.TeamAlbum{}

		if err := entity.Db().Where("team_uid = ? AND album_uid = ?", team.TeamUID, albumUID).First(&m).Error; err != nil {
			AbortEntityNotFound(c)
			return
		}

		if err := m.Delete(); err != nil {
			log.Errorf("team: %s (remove album)", err)
			AbortDeleteFailed(c)
			return
		}

		c.JSON(http.StatusOK, gin.H{"team_uid": team.TeamUID, "album_uid": albumUID})
	})
}

// AddTeamToPhoto grants a team access to a photo.
//
//	@Summary	grants a team access to a photo
//	@Id			AddTeamToPhoto
//	@Tags		Teams
//	@Accept		json
//	@Produce	json
//	@Success	200					{object}	entity.TeamPhoto
//	@Failure	400,401,403,404,500	{object}	i18n.Response
//	@Param		uid					path		string			true	"Team UID"
//	@Param		photo				body		form.TeamPhoto	true	"photo to attach"
//	@Router		/api/v1/teams/{uid}/photos [post]
func AddTeamToPhoto(router *gin.RouterGroup) {
	router.POST("/teams/:uid/photos", func(c *gin.Context) {
		s := Auth(c, acl.ResourceTeams, acl.ActionManageOwn)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))

		team := entity.FindTeam(uid)

		if team == nil {
			AbortEntityNotFound(c)
			return
		}

		if !s.GetUser().HasRole(acl.RoleAdmin) && team.UserUID != s.UserUID {
			AbortForbidden(c)
			return
		}

		var frm form.TeamPhoto

		LimitRequestBodyBytes(c, MaxMutationRequestBytes)

		if err := c.BindJSON(&frm); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}
			AbortBadRequest(c, err)
			return
		}

		if !rnd.IsUID(frm.PhotoUID, entity.PhotoUID) {
			AbortBadRequest(c)
			return
		}

		if _, err := query.PhotoByUID(frm.PhotoUID); err != nil {
			AbortEntityNotFound(c)
			return
		}

		tp := entity.NewTeamPhoto(team.TeamUID, frm.PhotoUID)

		if result := entity.FirstOrCreateTeamPhoto(tp); result == nil {
			AbortUnexpectedError(c)
			return
		} else {
			c.JSON(http.StatusOK, result)
		}
	})
}

// RemoveTeamFromPhoto removes a team's access to a photo.
//
//	@Summary	removes a team's access to a photo
//	@Id			RemoveTeamFromPhoto
//	@Tags		Teams
//	@Produce	json
//	@Success	200				{object}	gin.H
//	@Failure	401,403,404,500	{object}	i18n.Response
//	@Param		uid				path		string	true	"Team UID"
//	@Param		photoUID		path		string	true	"Photo UID"
//	@Router		/api/v1/teams/{uid}/photos/{photoUID} [delete]
func RemoveTeamFromPhoto(router *gin.RouterGroup) {
	router.DELETE("/teams/:uid/photos/:photoUID", func(c *gin.Context) {
		s := Auth(c, acl.ResourceTeams, acl.ActionManageOwn)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))
		photoUID := clean.UID(c.Param("photoUID"))

		team := entity.FindTeam(uid)

		if team == nil {
			AbortEntityNotFound(c)
			return
		}

		if !s.GetUser().HasRole(acl.RoleAdmin) && team.UserUID != s.UserUID {
			AbortForbidden(c)
			return
		}

		m := entity.TeamPhoto{}

		if err := entity.Db().Where("team_uid = ? AND photo_uid = ?", team.TeamUID, photoUID).First(&m).Error; err != nil {
			AbortEntityNotFound(c)
			return
		}

		if err := m.Delete(); err != nil {
			log.Errorf("team: %s (remove photo)", err)
			AbortDeleteFailed(c)
			return
		}

		c.JSON(http.StatusOK, gin.H{"team_uid": team.TeamUID, "photo_uid": photoUID})
	})
}
