<template>
  <v-dialog
    :model-value="visible"
    :max-width="680"
    @update:model-value="onDialogUpdate"
    @keydown.esc.stop="close"
  >
    <v-card>
      <v-toolbar density="comfortable" color="navigation">
        <v-toolbar-title>{{ $gettext("Reactions") }}</v-toolbar-title>
        <v-btn icon="mdi-close" :title="$gettext('Close')" @click.stop="close"></v-btn>
      </v-toolbar>

      <v-card-text class="pa-4">
        <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-4"></v-progress-linear>

        <!-- Emoji summary chips -->
        <div class="text-caption text-medium-emphasis mb-2">{{ $gettext("Reactions") }}</div>
        <div class="d-flex flex-wrap ga-1 mb-4">
          <v-chip
            v-for="item in allEmojis"
            :key="item.emoji"
            :color="newEmoji === item.emoji ? 'primary' : undefined"
            :variant="newEmoji === item.emoji ? 'flat' : 'outlined'"
            size="default"
            :closable="false"
            class="clickable reaction-chip"
            @click.stop="toggleEmoji(item.emoji)"
          >
            <span class="reaction-chip__emoji">{{ item.emoji }}</span>
            <span v-if="item.count > 0" class="ml-1">{{ item.count }}</span>
          </v-chip>
        </div>

        <!-- Add new reaction form -->
        <div class="text-caption text-medium-emphasis mb-2">{{ $gettext("Add a reaction") }}</div>
        <v-textarea
          v-model="newComment"
          :placeholder="$gettext('Comment (optional)')"
          density="compact"
          variant="outlined"
          hide-details
          auto-grow
          :rows="3"
          :max-rows="10"
          class="mb-3"
        ></v-textarea>
        <div class="d-flex justify-end mb-4">
          <v-btn
            variant="flat"
            color="primary"
            density="compact"
            :disabled="loading || (!newEmoji && !newComment)"
            @click.stop="addReaction"
          >
            {{ $gettext("Post") }}
          </v-btn>
        </div>

        <!-- My reactions -->
        <template v-if="mine.length > 0">
          <v-divider class="mb-3"></v-divider>
          <div class="text-caption text-medium-emphasis mb-2">{{ $gettext("My reactions") }}</div>
          <div class="mb-4">
            <div
              v-for="(item, index) in mine"
              :key="item.ID"
              class="d-flex align-start ga-2 py-2"
              :class="{ 'border-b': index < mine.length - 1 }"
            >
              <span v-if="item.Emoji" class="text-h6 line-height-1">{{ item.Emoji }}</span>
              <div class="flex-grow-1 min-width-0">
                <div v-if="item.Comment" class="text-body-2 text-medium-emphasis" style="white-space: pre-wrap; word-break: break-word;">{{ item.Comment }}</div>
              </div>
              <v-btn
                icon="mdi-delete-outline"
                size="x-small"
                variant="text"
                color="error"
                :title="$gettext('Delete')"
                :disabled="loading"
                @click.stop="deleteReaction(item.ID)"
              ></v-btn>
            </div>
          </div>
        </template>

        <!-- Scrollable history of all users' reactions -->
        <template v-if="details.length > 0 || total > 0">
          <v-divider class="mb-3"></v-divider>
          <div class="text-caption text-medium-emphasis mb-2">
            {{ $gettext("All reactions") }}
            <span v-if="total > 0" class="ml-1">({{ total }})</span>
          </div>

          <div
            ref="historyList"
            class="reaction-history-list"
            style="max-height: 320px; overflow-y: auto;"
            @scroll.passive="onHistoryScroll"
          >
            <div
              v-for="(item, index) in details"
              :key="index"
              class="d-flex align-start ga-2 py-2"
              :class="{ 'border-b': index < details.length - 1 }"
            >
              <span v-if="item.emoji" class="text-h6 line-height-1">{{ item.emoji }}</span>
              <div class="flex-grow-1 min-width-0">
                <div class="text-body-2 font-weight-medium">{{ item.user }}</div>
                <div v-if="item.comment" class="text-body-2 text-medium-emphasis" style="white-space: pre-wrap; word-break: break-word;">{{ item.comment }}</div>
              </div>
              <div class="text-caption text-medium-emphasis text-no-wrap">{{ formatDate(item.created_at) }}</div>
            </div>

            <!-- Load-more indicator -->
            <div v-if="loadingMore" class="text-center py-2">
              <v-progress-circular indeterminate color="primary" size="20"></v-progress-circular>
            </div>
            <div v-else-if="!hasMore && details.length > 0 && total > limit" class="text-caption text-medium-emphasis text-center py-2">
              {{ $gettext("All reactions loaded") }}
            </div>
          </div>
        </template>

        <div v-else-if="!loading && mine.length === 0" class="text-body-2 text-medium-emphasis text-center py-4">
          {{ $gettext("No reactions yet. Be the first!") }}
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script>
import $api from "common/api";
import { DateTime } from "luxon";

const EMOJIS = ["❤️", "👍", "😍", "🥰", "🔥", "🎉", "✨", "🌈"];

export default {
  name: "PReactionDialog",
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
      loadingMore: false,
      reactions: [],
      details: [],
      mine: [],
      total: 0,
      limit: 10,
      newEmoji: "",
      newComment: "",
    };
  },
  computed: {
    allEmojis() {
      return EMOJIS.map((emoji) => {
        const found = this.reactions.find((r) => r.emoji === emoji);
        return { emoji, count: found ? found.count : 0 };
      });
    },
    hasMore() {
      return this.details.length < this.total;
    },
  },
  watch: {
    visible: {
      immediate: true,
      handler(val) {
        if (val && this.model?.UID) {
          this.load();
        }
      },
    },
    model(val) {
      if (this.visible && val?.UID) {
        this.load();
      }
    },
  },
  methods: {
    load() {
      if (!this.model?.UID) {
        return;
      }

      this.loading = true;
      this.details = [];
      this.total = 0;
      this.newComment = "";
      this.newEmoji = "";

      $api
        .get(`photos/${this.model.UID}/react`, { params: { offset: 0 } })
        .then((r) => {
          this.reactions = r.data.reactions || [];
          this.details = r.data.details || [];
          this.mine = r.data.mine || [];
          this.total = r.data.total || 0;
          this.limit = r.data.limit || 10;
        })
        .catch(() => {
          this.$notify.error(this.$gettext("Could not load reactions"));
        })
        .finally(() => {
          this.loading = false;
        });
    },
    loadMore() {
      if (this.loadingMore || !this.hasMore || !this.model?.UID) {
        return;
      }

      this.loadingMore = true;
      const offset = this.details.length;

      $api
        .get(`photos/${this.model.UID}/react`, { params: { offset } })
        .then((r) => {
          const newDetails = r.data.details || [];
          this.details = [...this.details, ...newDetails];
          this.total = r.data.total || this.total;
        })
        .catch(() => {
          this.$notify.error(this.$gettext("Could not load more reactions"));
        })
        .finally(() => {
          this.loadingMore = false;
        });
    },
    onHistoryScroll(e) {
      const el = e.target;
      if (el.scrollTop + el.clientHeight >= el.scrollHeight - 60) {
        this.loadMore();
      }
    },
    toggleEmoji(emoji) {
      this.newEmoji = this.newEmoji === emoji ? "" : emoji;
    },
    addReaction() {
      if (!this.model?.UID || (!this.newEmoji && !this.newComment)) {
        return;
      }

      this.loading = true;

      $api
        .post(`photos/${this.model.UID}/react`, {
          emoji: this.newEmoji,
          comment: this.newComment,
        })
        .then(() => {
          this.load();
        })
        .catch(() => {
          this.$notify.error(this.$gettext("Could not save reaction"));
          this.loading = false;
        });
    },
    deleteReaction(id) {
      if (!this.model?.UID || !id) {
        return;
      }

      this.loading = true;

      $api
        .delete(`photos/${this.model.UID}/react/${id}`)
        .then(() => {
          this.load();
        })
        .catch(() => {
          this.$notify.error(this.$gettext("Could not delete reaction"));
          this.loading = false;
        });
    },
    formatDate(iso) {
      if (!iso) {
        return "";
      }

      return DateTime.fromISO(iso).toRelative() || "";
    },
    onDialogUpdate(val) {
      if (!val) {
        this.close();
      }
    },
    close() {
      this.$emit("close");
    },
  },
};
</script>
