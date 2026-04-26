<template>
  <div class="p-reaction-panel">
    <v-divider class="my-4"></v-divider>
    <div class="px-4 pb-2">
      <div class="text-subtitle-2 mb-3">{{ $gettext("Reactions") }}</div>

      <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-3"></v-progress-linear>

      <!-- Emoji summary chips -->
      <div class="d-flex flex-wrap ga-1 mb-3">
        <v-chip
          v-for="item in allEmojis"
          :key="item.emoji"
          :color="newEmoji === item.emoji ? 'primary' : undefined"
          :variant="newEmoji === item.emoji ? 'flat' : 'outlined'"
          size="small"
          :closable="false"
          class="clickable reaction-chip"
          :title="item.emoji"
          @click.stop="toggleEmoji(item.emoji)"
        >
          <span class="reaction-chip__emoji">{{ item.emoji }}</span>
          <span v-if="item.count > 0" class="ml-1 text-caption">{{ item.count }}</span>
        </v-chip>
      </div>

      <!-- Add reaction form -->
      <v-textarea
        v-model="newComment"
        :placeholder="$gettext('Comment (optional)')"
        density="compact"
        variant="outlined"
        hide-details
        auto-grow
        :rows="2"
        :max-rows="5"
        class="mb-2"
      ></v-textarea>

      <div class="d-flex justify-end mb-3">
        <v-btn
          variant="flat"
          color="primary"
          density="compact"
          size="small"
          :disabled="loading || (!newEmoji && !newComment)"
          @click.stop="addReaction"
        >
          {{ $gettext("Post") }}
        </v-btn>
      </div>

      <!-- My reactions with delete buttons -->
      <template v-if="mine.length > 0">
        <div class="text-caption text-medium-emphasis mb-1">{{ $gettext("My reactions") }}</div>
        <div
          v-for="item in mine"
          :key="item.ID"
          class="d-flex align-start ga-2 py-1"
        >
          <span v-if="item.Emoji" class="text-body-1">{{ item.Emoji }}</span>
          <div class="flex-grow-1 min-width-0 text-body-2 text-medium-emphasis" style="white-space: pre-wrap; word-break: break-word;">{{ item.Comment }}</div>
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
      </template>
    </div>
  </div>
</template>

<script>
import $api from "common/api";

const EMOJIS = ["❤️", "👍", "😍", "🥰", "🔥", "🎉", "✨", "🌈"];

export default {
  name: "PReactionPanel",
  props: {
    model: {
      type: Object,
      default: () => null,
    },
  },
  data() {
    return {
      loading: false,
      reactions: [],
      mine: [],
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
  },
  watch: {
    model: {
      immediate: true,
      handler(val) {
        if (val?.UID) {
          this.load();
        }
      },
    },
  },
  methods: {
    load() {
      if (!this.model?.UID) {
        return;
      }

      this.loading = true;

      $api
        .get(`photos/${this.model.UID}/react`)
        .then((r) => {
          this.reactions = r.data.reactions || [];
          this.mine = r.data.mine || [];
        })
        .catch(() => {
          this.$notify.error(this.$gettext("Could not load reactions"));
        })
        .finally(() => {
          this.loading = false;
        });
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
          this.newEmoji = "";
          this.newComment = "";
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
  },
};
</script>
