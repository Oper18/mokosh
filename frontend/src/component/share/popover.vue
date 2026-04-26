<template>
  <div
    v-if="visible"
    class="p-share-popover"
    style="position: fixed; top: 56px; right: 8px; z-index: 10001; width: 360px; max-width: calc(100vw - 16px);"
    @click.stop
  >
    <v-card elevation="8" rounded="lg">
      <v-card-title class="d-flex justify-space-between align-center px-4 py-2">
        <span class="text-subtitle-2">{{ $gettext("Share") }}</span>
        <v-btn
          icon="mdi-close"
          variant="text"
          density="compact"
          size="small"
          :title="$gettext('Close')"
          @click.stop="close"
        ></v-btn>
      </v-card-title>
      <v-divider></v-divider>

      <v-card-text class="pa-2">
        <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-1"></v-progress-linear>

        <div v-if="!loading && links.length === 0" class="text-caption text-medium-emphasis text-center py-2">
          {{ $gettext("No shared links yet.") }}
        </div>

        <div v-for="(link, index) in links" :key="link.UID">
          <v-divider v-if="index > 0" class="my-2"></v-divider>

          <!-- URL copy row -->
          <v-text-field
            :model-value="link.url()"
            readonly
            hide-details
            density="compact"
            variant="outlined"
            append-inner-icon="mdi-content-copy"
            style="cursor: pointer;"
            :title="$gettext('Copy link')"
            @click.stop="$util.copyText(link.url())"
            @click:append-inner.stop="$util.copyText(link.url())"
          ></v-text-field>

          <!-- Name row -->
          <v-text-field
            v-model="link.Name"
            :readonly="!isOwner"
            hide-details
            density="compact"
            variant="outlined"
            autocorrect="off"
            autocapitalize="none"
            autocomplete="off"
            :label="$gettext('Name')"
            :placeholder="$gettext('Display name')"
            class="mt-1 input-name"
          ></v-text-field>

          <!-- Expiry row -->
          <v-select
            v-model="link.Expires"
            hide-details
            density="compact"
            variant="outlined"
            :label="$gettext('Expires')"
            :items="expiresOptions"
            item-title="text"
            item-value="value"
            class="mt-1"
          ></v-select>

          <!-- Permissions -->
          <div class="text-caption text-medium-emphasis mt-2 mb-1 px-1">{{ $gettext("Permissions") }}</div>
          <div class="d-flex flex-wrap ga-1 px-1">
            <v-chip
              v-for="perm in permOptions"
              :key="perm.value"
              :color="hasPermission(link, perm.value) ? 'primary' : undefined"
              :variant="hasPermission(link, perm.value) ? 'flat' : 'outlined'"
              size="small"
              :closable="false"
              style="cursor: pointer;"
              @click.stop="togglePermission(link, perm.value)"
            >
              {{ perm.label }}
            </v-chip>
          </div>

          <!-- Actions row -->
          <div class="d-flex justify-space-between align-center mt-2">
            <v-btn
              icon="mdi-delete"
              variant="text"
              density="compact"
              size="small"
              color="error"
              :title="$gettext('Delete')"
              @click.stop="remove(index)"
            ></v-btn>
            <v-btn
              variant="flat"
              color="primary"
              density="compact"
              size="small"
              :title="$gettext('Save')"
              @click.stop="update(link)"
            >
              {{ $gettext("Save") }}
            </v-btn>
          </div>
        </div>
      </v-card-text>

      <v-divider></v-divider>
      <v-card-actions class="px-3 py-2">
        <v-btn
          prepend-icon="mdi-link-plus"
          variant="text"
          density="compact"
          color="primary"
          :disabled="loading"
          @click.stop="add"
        >
          {{ $gettext("Add Link") }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </div>
</template>

<script>
import * as options from "options/options";

// Permission bitmask constants (must match internal/entity/auth_user_share.go).
const PermNone = 1;
const PermView = 2;
const PermReact = 4;
const PermComment = 8;
const PermUpload = 16;
const PermEdit = 32;
const PermShare = 64;

export default {
  name: "PSharePopover",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    model: {
      type: Object,
      default: () => null,
    },
  },
  emits: ["close"],
  data() {
    return {
      loading: false,
      links: [],
      isOwner: this.$session.isUser(),
      expiresOptions: options.Expires(),
      permOptions: [
        { value: PermView, label: this.$gettext("View") },
        { value: PermReact, label: this.$gettext("React") },
        { value: PermComment, label: this.$gettext("Comment") },
        { value: PermUpload, label: this.$gettext("Upload") },
        { value: PermEdit, label: this.$gettext("Edit") },
        { value: PermShare, label: this.$gettext("Share") },
      ],
    };
  },
  watch: {
    visible: {
      immediate: true,
      handler(show) {
        if (show) {
          this.load();
        }
      },
    },
    model() {
      if (this.visible) {
        this.load();
      }
    },
  },
  methods: {
    hasPermission(link, bit) {
      // Perm === 0 means default (view only); treat as PermView set.
      if (link.Perm === 0) {
        return bit === PermView;
      }
      return (link.Perm & bit) !== 0;
    },
    togglePermission(link, bit) {
      let perm = link.Perm === 0 ? PermView : link.Perm;
      if (perm & bit) {
        perm = perm & ~bit;
      } else {
        perm = perm | bit;
      }
      // Ensure at least PermView is always set (can't share without view).
      if (!(perm & PermView)) {
        perm = perm | PermView;
      }
      link.Perm = perm;
    },
    load() {
      if (!this.model) {
        return;
      }
      this.links = [];
      this.loading = true;
      this.model
        .links()
        .then((resp) => {
          this.links = resp.models;
        })
        .finally(() => {
          this.loading = false;
        });
    },
    add() {
      if (!this.model) {
        return;
      }
      this.loading = true;
      this.model
        .createLink()
        .then((r) => {
          this.links.push(r);
        })
        .finally(() => {
          this.loading = false;
        });
    },
    update(link) {
      if (!link || !this.model) {
        return;
      }
      // Normalise Perm=0 to PermView before saving.
      if (link.Perm === 0) {
        link.Perm = PermView;
      }
      this.loading = true;
      this.model
        .updateLink(link)
        .then(() => {
          this.$notify.success(this.$gettext("Changes successfully saved"));
        })
        .finally(() => {
          this.loading = false;
        });
    },
    remove(index) {
      const link = this.links[index];
      if (!link || !this.model) {
        return;
      }
      this.loading = true;
      this.model
        .removeLink(link)
        .then(() => {
          this.$notify.success(this.$gettext("Changes successfully saved"));
          this.links.splice(index, 1);
        })
        .finally(() => {
          this.loading = false;
        });
    },
    close() {
      this.$emit("close");
    },
  },
};
</script>
