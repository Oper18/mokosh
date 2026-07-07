package acl

import "sync"

// RulesMutex guards concurrent updates to the ACL rules.
var RulesMutex = &sync.Mutex{}

// Rules specifies granted permissions by Resource and Role.
var Rules = ACL{
	ResourceFiles: GrantDefaults,
	ResourceFolders: Roles{
		RoleAdmin:        GrantFullAccess,
		RolePhotographer: GrantGuestOwn,
		RoleGuest:        GrantGuestOwn,
		RoleVisitor:      GrantSearchShared,
		RoleClient:       GrantFullAccess,
	},
	ResourceShares: Roles{
		RoleAdmin:        GrantFullAccess,
		RolePhotographer: GrantFullAccess, // Identical to guest.
		RoleGuest:        GrantFullAccess,
		RoleClient:       GrantFullAccess,
	},
	ResourcePhotos: Roles{
		RoleAdmin:        GrantFullAccess,
		RolePhotographer: GrantGuestPhotos, // Align with guest: search, download, and own updates only — no AccessLibrary (see GrantGuestPhotos doc).
		RoleGuest:        GrantGuestPhotos, // Restrict global photo updates and whole-library visibility; only own + shared photos.
		RoleVisitor:      GrantSearchShared,
		RoleClient:       GrantFullAccess,
	},
	ResourceVideos: GrantDefaults,
	ResourceFavorites: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantFullAccess,
	},
	ResourceAlbums: GrantDefaults,
	ResourceMoments: Roles{
		RoleAdmin:        GrantFullAccess,
		RolePhotographer: GrantGuestOwn,
		RoleGuest:        GrantGuestOwn,
		RoleVisitor:      GrantSearchShared,
		RoleClient:       GrantFullAccess,
	},
	ResourceCalendar: Roles{
		RoleAdmin:        GrantFullAccess,
		RolePhotographer: GrantGuestOwn,
		RoleGuest:        GrantGuestOwn,
		RoleVisitor:      GrantSearchShared,
		RoleClient:       GrantFullAccess,
	},
	ResourcePeople: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantFullAccess,
	},
	ResourcePlaces: Roles{
		RoleAdmin:        GrantFullAccess,
		RolePhotographer: GrantGuestOwn,
		RoleGuest:        GrantGuestOwn,
		RoleVisitor:      GrantViewShared,
		RoleInstance:     GrantUseOwn,
		RoleService:      GrantUseOwn,
		RolePortal:       GrantUseOwn,
		RoleClient:       GrantFullAccess,
	},
	ResourceLabels: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantFullAccess,
	},
	ResourceConfig: Roles{
		RoleAdmin:   GrantFullAccess,
		RolePortal:  GrantFullAccess,
		RoleClient:  GrantViewOwn,
		RoleDefault: GrantViewOwn,
	},
	ResourceSettings: Roles{
		RoleAdmin:        GrantFullAccess,
		RolePhotographer: GrantViewUpdateOwn,
		RoleGuest:        GrantViewUpdateOwn,
		RoleVisitor:      GrantViewOwn,
		RolePortal:       GrantFullAccess,
		RoleClient:       GrantViewUpdateOwn,
	},
	ResourceServices: Roles{
		RoleAdmin:  GrantFullAccess,
		RolePortal: GrantFullAccess,
	},
	ResourcePasscode: Roles{
		RoleAdmin:        GrantFullAccess,
		RolePhotographer: GrantConfigureOwn,
		RolePortal:       GrantFullAccess,
		RoleGuest:        GrantConfigureOwn,
	},
	ResourcePassword: Roles{
		RoleAdmin:        GrantFullAccess,
		RolePhotographer: GrantUpdateOwn,
		RolePortal:       GrantFullAccess,
		RoleGuest:        GrantUpdateOwn,
	},
	ResourceUsers: Roles{
		RoleAdmin:        GrantManageOwn,
		RolePhotographer: GrantViewUpdateOwn,
		RoleGuest:        GrantViewUpdateOwn,
		RoleInstance:     GrantViewOwn,
		RoleService:      GrantViewOwn,
		RolePortal:       GrantFullAccess,
		RoleClient:       GrantViewOwn,
	},
	ResourceSessions: Roles{
		RoleAdmin:   GrantManageOwn,
		RolePortal:  GrantFullAccess,
		RoleVisitor: GrantViewOwn,
		RoleDefault: GrantOwn, // Guest and photographer fall through here, keeping them identical.
	},
	ResourceLogs: Roles{
		RoleAdmin:  GrantFullAccess,
		RolePortal: GrantFullAccess,
		RoleClient: GrantFullAccess,
	},
	ResourceApi: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantPublishOwn,
	},
	ResourceWebDAV: Roles{
		RoleAdmin:  GrantFullAccess,
		RolePortal: GrantFullAccess,
		RoleClient: GrantFullAccess,
	},
	ResourceWebhooks: Roles{
		RoleAdmin:  GrantFullAccess,
		RoleClient: GrantPublishOwn,
	},
	ResourceMetrics: Roles{
		RoleAdmin:   GrantFullAccess,
		RoleService: GrantViewAll,
		RolePortal:  GrantViewAll,
		RoleClient:  GrantViewAll,
	},
	ResourceVision: Roles{
		RoleAdmin:    GrantFullAccess,
		RoleInstance: GrantUseOwn,
		RoleService:  GrantUseOwn,
		RolePortal:   GrantUseOwn,
		RoleClient:   GrantUseOwn,
	},
	ResourceCluster: Roles{
		RoleAdmin:    GrantFullAccess,
		RoleInstance: GrantSearchDownloadUpdateOwn,
		RoleService:  GrantSearchDownloadUpdateOwn,
		RolePortal:   GrantFullAccess,
		RoleClient:   GrantSearchDownloadUpdateOwn,
	},
	ResourceTeams: Roles{
		RoleAdmin:        GrantFullAccess,
		RolePhotographer: GrantManageOwn,
		RoleGuest:        GrantManageOwn,
	},
	ResourceFeedback: Roles{
		RoleAdmin: GrantFullAccess,
	},
	ResourceDefault: Roles{
		RoleAdmin:    GrantFullAccess,
		RoleInstance: GrantNone,
		RoleService:  GrantNone,
		RolePortal:   GrantNone,
		RoleClient:   GrantNone,
	},
}
