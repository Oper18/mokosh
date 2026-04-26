<template>
  <div v-if="visible" class="p-fab" style="position: fixed; bottom: 0; right: 0; z-index: 100;">
    <!-- Create Album inline dialog -->
    <v-dialog v-model="createAlbum.visible" max-width="400" @keydown.esc="closeCreateAlbum">
      <v-card>
        <v-card-title class="text-h6 pa-4">{{ $gettext("New Album") }}</v-card-title>
        <v-card-text class="pa-4 pt-0">
          <v-text-field
            ref="albumTitle"
            v-model="createAlbum.title"
            :disabled="createAlbum.loading"
            :placeholder="$gettext('Album Title')"
            :label="$gettext('Title')"
            variant="outlined"
            density="comfortable"
            autofocus
            hide-details
            class="input-album-title"
            @keyup.enter="onCreateAlbum"
          ></v-text-field>
        </v-card-text>
        <v-card-actions class="pa-4 pt-0 d-flex justify-end ga-2">
          <v-btn variant="text" :disabled="createAlbum.loading" class="action-cancel" @click.stop="closeCreateAlbum">
            {{ $gettext("Cancel") }}
          </v-btn>
          <v-btn
            color="highlight"
            variant="flat"
            :disabled="createAlbum.loading || createAlbum.title.trim() === ''"
            class="action-confirm"
            @click.stop="onCreateAlbum"
          >
            {{ $gettext("Create") }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Speed dial FAB -->
    <v-speed-dial
      v-model="expanded"
      class="p-fab-dial"
      location="top"
      transition="slide-y-reverse-transition"
      offset="12"
    >
      <template #activator="{ props }">
        <v-btn
          v-bind="props"
          icon="mdi-plus"
          size="52"
          color="highlight"
          variant="elevated"
          density="comfortable"
          class="action-menu opacity-95 ma-5"
          :title="$gettext('Create')"
        ></v-btn>
      </template>

      <v-btn
        v-if="canUpload"
        key="action-upload"
        icon="mdi-cloud-upload"
        color="upload"
        variant="elevated"
        density="comfortable"
        :title="$gettext('Upload')"
        class="action-upload"
        @click.stop="onUpload"
      ></v-btn>
      <v-btn
        v-if="canCreateAlbum"
        key="action-create-album"
        icon="mdi-image-album"
        color="accent"
        variant="elevated"
        density="comfortable"
        :title="$gettext('New Album')"
        class="action-create-album"
        @click.stop="openCreateAlbum"
      ></v-btn>
    </v-speed-dial>
  </div>
</template>

<script>
import Album from "model/album";

export default {
  name: "PFab",
  data() {
    return {
      expanded: false,
      rtl: this.$config.isRtl(),
      selection: this.$clipboard.selection,
      albumSelection: this.$albumClipboard.selection,
      createAlbum: {
        visible: false,
        loading: false,
        title: "",
      },
    };
  },
  computed: {
    visible() {
      // Hide on auth pages (login, register) and when not authenticated.
      if (this.$route.meta?.hideNav) {
        return false;
      }

      // Hide when items are selected (clipboard toolbar takes over).
      if (this.selection.length > 0 || this.albumSelection.length > 0) {
        return false;
      }

      return this.$session.auth && (this.canCreateAlbum || this.canUpload);
    },
    canCreateAlbum() {
      return this.$config.allow("albums", "manage") || this.$config.allow("albums", "create");
    },
    canUpload() {
      return !this.$config.get("readonly") && this.$config.feature("upload");
    },
  },
  methods: {
    onUpload() {
      this.expanded = false;
      this.$event.publish("dialog.upload", {});
    },
    openCreateAlbum() {
      this.expanded = false;
      this.createAlbum.title = "";
      this.createAlbum.visible = true;
    },
    closeCreateAlbum() {
      this.createAlbum.visible = false;
      this.createAlbum.title = "";
    },
    onCreateAlbum() {
      const title = this.createAlbum.title.trim();

      if (title === "" || this.createAlbum.loading) {
        return;
      }

      this.createAlbum.loading = true;

      new Album({ Title: title, Favorite: false })
        .save()
        .then((album) => {
          this.$notify.success(this.$gettext("Album created"));
          this.closeCreateAlbum();

          // Navigate to the new album if not already on the albums page.
          if (album?.UID) {
            this.$router.push({ name: "album", params: { album: album.UID, slug: album.Slug || "view" } }).catch(() => {});
          }
        })
        .catch(() => {
          this.$notify.error(this.$gettext("Failed to create album"));
        })
        .finally(() => {
          this.createAlbum.loading = false;
        });
    },
  },
};
</script>

