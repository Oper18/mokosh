package api

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// watermarkTextTTL is how long the watermark text decision is cached per session.
const watermarkTextTTL = 5 * time.Minute

// GetThumb returns a thumbnail image matching the file hash, crop area, and type.
//
//	@Summary		returns a thumbnail image with the requested size
//	@Description	Fore more information see:
//	@Description	- https://docs.photoprism.app/developer-guide/api/thumbnails/#image-endpoint-uri
//	@Id				GetThumb
//	@Produce		image/jpeg, image/svg+xml
//	@Tags			Images, Files
//	@Failure		403		{file}	image/svg+xml
//	@Success		200		{file}	image/svg+xml
//	@Success		200		{file}	image/jpg
//	@Param			thumb	path	string	true	"SHA1 file hash, optionally with a crop area suffixed, e.g. '-016014058037'"
//	@Param			token	path	string	true	"user-specific security token provided with session or 'public' when running Mokosh in public mode"
//	@Param			size	path	string	true	"thumbnail size"	Enums(tile_50, tile_100, left_224, right_224, tile_224, tile_500, fit_720, tile_1080, fit_1280, fit_1600, fit_1920, fit_2048, fit_2560, fit_3840, fit_4096, fit_7680)
//	@Router			/api/v1/t/{thumb}/{token}/{size} [get]
func GetThumb(router *gin.RouterGroup) {
	router.GET("/t/:thumb/:token/:size", func(c *gin.Context) {
		if InvalidPreviewToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", brokenIconSvg)
			return
		}

		// Resolve watermark label for sessions that have no download permission.
		watermarkText := thumbWatermarkText(c)

		logPrefix := "thumb"

		start := time.Now()
		conf := get.Config()
		attachment := c.Query("download") != ""
		fileHash, cropArea := crop.ParseThumb(clean.Token(c.Param("thumb")))

		// Is cropped thumbnail?
		if cropArea != "" {
			cropName := crop.Name(clean.Token(c.Param("size")))

			cropSize, ok := crop.Sizes[cropName]

			if !ok {
				log.Errorf("%s: invalid size %s", logPrefix, clean.Log(string(cropName)))
				c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
				return
			}

			fileName, err := crop.FromRequest(fileHash, cropArea, cropSize, conf.ThumbCachePath())

			if err != nil {
				log.Warnf("%s: %s", logPrefix, err)
				c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
				return
			} else if fileName == "" {
				log.Errorf("%s: empty file name - you may have found a bug", logPrefix)
				c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
				return
			}

			// Add HTTP cache header.
			AddImmutableCacheHeader(c)

			if watermarkText != "" {
				wmKey := CacheKey("wm", watermarkText+":"+fileHash, string(cropName))
				wmCache := get.ThumbCache()
				if wmData, ok := wmCache.Get(wmKey); ok {
					c.Data(http.StatusOK, "image/jpeg", wmData.(ByteCache).Data)
				} else if wmBytes, wmErr := thumb.WatermarkFile(fileName, watermarkText); wmErr == nil {
					wmCache.SetDefault(wmKey, ByteCache{Data: wmBytes})
					c.Data(http.StatusOK, "image/jpeg", wmBytes)
				} else {
					log.Warnf("%s: watermark failed: %s", logPrefix, wmErr)
					if attachment {
						c.FileAttachment(fileName, cropName.Jpeg())
					} else {
						c.File(fileName)
					}
				}
				return
			}

			if attachment {
				c.FileAttachment(fileName, cropName.Jpeg())
			} else {
				c.File(fileName)
			}

			return
		}

		sizeName := thumb.Name(clean.Token(c.Param("size")))

		size, ok := thumb.Sizes[sizeName]

		if !ok {
			log.Errorf("%s: invalid size %s", logPrefix, clean.Log(sizeName.String()))
			c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
			return
		}

		if size.Uncached() && !conf.ThumbUncached() {
			sizeName, size = thumb.Find(conf.ThumbSizePrecached())

			if sizeName == "" {
				log.Errorf("%s: invalid size %d", logPrefix, conf.ThumbSizePrecached())
				c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
				return
			}
		}

		cache := get.ThumbCache()
		cacheKey := CacheKey("thumbs", fileHash, string(sizeName))
		wmCacheKey := CacheKey("wm", watermarkText+":"+fileHash, string(sizeName))

		if cacheData, ok := cache.Get(cacheKey); ok {
			log.Tracef("api: cache hit for %s [%s]", cacheKey, time.Since(start))

			cached := cacheData.(ThumbCache)

			if !fs.FileExists(cached.FileName) {
				log.Errorf("%s: %s not found", logPrefix, fileHash)
				c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
				return
			}

			// Add HTTP cache header.
			AddImmutableCacheHeader(c)

			if watermarkText != "" {
				if wmData, ok := cache.Get(wmCacheKey); ok {
					c.Data(http.StatusOK, "image/jpeg", wmData.(ByteCache).Data)
				} else if wmBytes, wmErr := thumb.WatermarkFile(cached.FileName, watermarkText); wmErr == nil {
					cache.SetDefault(wmCacheKey, ByteCache{Data: wmBytes})
					c.Data(http.StatusOK, "image/jpeg", wmBytes)
				} else {
					log.Warnf("%s: watermark failed: %s", logPrefix, wmErr)
					if attachment {
						c.FileAttachment(cached.FileName, cached.ShareName)
					} else {
						c.File(cached.FileName)
					}
				}
				return
			}

			if attachment {
				c.FileAttachment(cached.FileName, cached.ShareName)
			} else {
				c.File(cached.FileName)
			}

			return
		}

		// Return existing thumbs straight away; skip this fast path for watermarked responses
		// so that we can apply the watermark before serving.
		if watermarkText == "" && !attachment {
			if fileName, err := size.ResolvedName(fileHash, conf.ThumbCachePath()); err == nil {
				// Add HTTP cache header.
				AddImmutableCacheHeader(c)

				// Return requested content.
				c.File(fileName)
				return
			}
		} else if watermarkText != "" && !attachment {
			// Check watermark cache before doing a full DB lookup.
			if wmData, ok := cache.Get(wmCacheKey); ok {
				AddImmutableCacheHeader(c)
				c.Data(http.StatusOK, "image/jpeg", wmData.(ByteCache).Data)
				return
			}

			// Serve watermarked thumbnail from existing disk cache if available.
			if fileName, err := size.ResolvedName(fileHash, conf.ThumbCachePath()); err == nil {
				AddImmutableCacheHeader(c)
				if wmBytes, wmErr := thumb.WatermarkFile(fileName, watermarkText); wmErr == nil {
					cache.SetDefault(wmCacheKey, ByteCache{Data: wmBytes})
					c.Data(http.StatusOK, "image/jpeg", wmBytes)
				} else {
					log.Warnf("%s: watermark failed: %s", logPrefix, wmErr)
					c.File(fileName)
				}
				return
			}
		}

		// Query index for file infos.
		f, err := query.FileByHash(fileHash)

		if err != nil {
			c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
			return
		}

		// Find supported preview image if media file is not a JPEG or PNG.
		if f.NoJpeg() && f.NoPng() {
			if f, err = query.FileByPhotoUID(f.PhotoUID); err != nil {
				c.Data(http.StatusOK, "image/svg+xml", fileIconSvg)
				return
			}
		}

		// Return SVG icon as placeholder if file has errors.
		if f.FileError != "" {
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
			return
		}

		var fileName string

		// Resolve the original locally, fetching it from remote storage (e.g. S3)
		// into the local cache when it is not already present on disk.
		if resolved, _, resolveErr := get.ResolveLocalFile(f); resolveErr == nil {
			fileName = resolved
		} else {
			log.Errorf("%s: file %s is missing", logPrefix, clean.Log(f.FileName))
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)

			// Only flag the file as missing — and only trash the photo — when remote
			// storage confirms the original is gone. Never destroy data because of a
			// missing local cache copy or a transient remote-storage outage.
			if !get.RemoteMayHold(f) {
				logErr(logPrefix, f.Update("FileMissing", true))

				if f.AllFilesMissing() {
					log.Infof("%s: deleting photo, all files missing for %s", logPrefix, clean.Log(f.FileName))

					if _, err := f.RelatedPhoto().Delete(false); err != nil {
						log.Errorf("%s: %s while deleting %s", logPrefix, err, clean.Log(f.FileName))
					}
				}
			}

			return
		}

		// Choose the smallest fitting size if the original image is smaller.
		if size.Fit && f.Bounds().In(size.Bounds()) {
			size = thumb.FitBounds(f.Bounds())
			log.Tracef("%s: smallest fitting size for %s is %s (width %d, height %d)", logPrefix, clean.Log(f.FileName), size.Name, size.Width, size.Height)
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if size.ExceedsLimit() && !attachment {
			log.Debugf("%s: using original, size exceeds limit (width %d, height %d)", logPrefix, size.Width, size.Height)

			// Add HTTP cache header.
			AddImmutableCacheHeader(c)

			if watermarkText != "" {
				if wmData, ok := cache.Get(wmCacheKey); ok {
					c.Data(http.StatusOK, "image/jpeg", wmData.(ByteCache).Data)
				} else if wmBytes, wmErr := thumb.WatermarkFile(fileName, watermarkText); wmErr == nil {
					cache.SetDefault(wmCacheKey, ByteCache{Data: wmBytes})
					c.Data(http.StatusOK, "image/jpeg", wmBytes)
				} else {
					log.Warnf("%s: watermark failed: %s", logPrefix, wmErr)
					c.File(fileName)
				}
				return
			}

			// Return requested content.
			c.File(fileName)
			return
		}

		// thumbName is the thumbnail filename.
		var thumbName string

		// Try to find or create thumbnail image.
		if conf.ThumbUncached() || size.Uncached() {
			thumbName, err = size.FromFile(fileName, f.FileHash, conf.ThumbCachePath(), f.FileOrientation)
		} else {
			thumbName, err = size.FromCache(fileName, f.FileHash, conf.ThumbCachePath())
		}

		// Failed?
		if err != nil {
			log.Errorf("%s: %s", logPrefix, err)
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
			return
		} else if thumbName == "" {
			log.Errorf("%s: %s has empty thumb name - you may have found a bug", logPrefix, filepath.Base(fileName))
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
			return
		}

		// Cache thumbnail filename to reduce the number of index queries.
		cache.SetDefault(cacheKey, ThumbCache{thumbName, f.ShareBase(0)})
		log.Debugf("cached %s [%s]", cacheKey, time.Since(start))

		// Add HTTP cache header.
		AddImmutableCacheHeader(c)

		if watermarkText != "" {
			if wmData, ok := cache.Get(wmCacheKey); ok {
				c.Data(http.StatusOK, "image/jpeg", wmData.(ByteCache).Data)
			} else if wmBytes, wmErr := thumb.WatermarkFile(thumbName, watermarkText); wmErr == nil {
				cache.SetDefault(wmCacheKey, ByteCache{Data: wmBytes})
				c.Data(http.StatusOK, "image/jpeg", wmBytes)
			} else {
				log.Warnf("%s: watermark failed: %s", logPrefix, wmErr)
				if attachment {
					c.FileAttachment(thumbName, f.DownloadName(DownloadName(c), 0))
				} else {
					c.File(thumbName)
				}
			}
			return
		}

		// Return requested content.
		if attachment {
			c.FileAttachment(thumbName, f.DownloadName(DownloadName(c), 0))
		} else {
			c.File(thumbName)
		}
	})
}

// thumbWatermarkText resolves the watermark label for the session associated with the request token.
// Returns an empty string when no watermark should be applied (public mode, no shares, or download is permitted).
// The result is cached in the thumb cache for watermarkTextTTL to avoid repeated DB lookups per thumbnail.
func thumbWatermarkText(c *gin.Context) string {
	token := clean.UrlToken(c.Param("token"))
	if token == "" {
		token = clean.UrlToken(c.Query("t"))
	}

	// Resolve session ID from the preview or download token maps.
	sessionID := entity.PreviewToken.Get(token)
	if sessionID == "" {
		sessionID = entity.DownloadToken.Get(token)
	}
	if sessionID == "" {
		return ""
	}

	// Check the short-lived cache to avoid repeated DB queries on every thumbnail request.
	cache := get.ThumbCache()
	wmTextKey := CacheKey("wmtext", sessionID, "")
	if cached, ok := cache.Get(wmTextKey); ok {
		return cached.(string)
	}

	sess, err := entity.FindSession(sessionID)
	if err != nil || sess == nil {
		return ""
	}

	user := sess.GetUser()
	text := ""

	if user.IsRegistered() {
		// For registered users, check shares explicitly since UserShares is not auto-loaded.
		shares := entity.FindUserShares(user.GetUID())
		for _, share := range shares {
			if share.Perm&entity.PermDownload != 0 {
				continue
			}
			if share.LinkUID == "" {
				continue
			}
			if link := entity.FindLink(share.LinkUID); link != nil && link.CreatedBy != "" {
				if owner := entity.FindUserByUID(link.CreatedBy); owner != nil {
					text = ownerLabel(owner)
					break
				}
			}
		}
	} else {
		// Visitor session: find the owner from redeemed share link tokens.
		data := sess.GetData()
		if data != nil {
		outer:
			for _, token := range data.Tokens {
				for _, link := range entity.FindValidLinks(token, "") {
					if link.Perm&entity.PermDownload != 0 {
						continue
					}
					if link.CreatedBy == "" {
						continue
					}
					if owner := entity.FindUserByUID(link.CreatedBy); owner != nil {
						text = ownerLabel(owner)
						break outer
					}
				}
			}
		}
	}

	cache.Set(wmTextKey, text, watermarkTextTTL)
	return text
}

// ownerLabel returns the display name or email of the share owner, empty string if neither is set.
func ownerLabel(user *entity.User) string {
	if user.DisplayName != "" {
		return user.DisplayName
	}
	if user.UserEmail != "" {
		return user.UserEmail
	}
	return ""
}
