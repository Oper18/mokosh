<template>
  <svg
    class="v-icon__svg"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path v-for="(d, i) in paths" :key="i" :d="d"></path>
  </svg>
</template>
<script>
// Mokosh navigation icon pack (24x24, stroked). Rendered via the Vuetify
// "mokosh" icon set, e.g. <v-icon>mokosh:albums</v-icon>.
// search: ["M11 4C14.866 4 18 7.134 18 11C18 14.866 14.866 18 11 18C7.134 18 4 14.866 4 11C4 7.134 7.134 4 11 4Z", "M20 20L16.5 16.5", "M11 8V14", "M8 11H14"],
const PATHS = {
  search: [
    "M9.5 3A6.5 6.5 0 0 1 16 9.5C16 10.61 15.72 11.66 15.22 12.59L15.5 12.87H16.29L21 17.59L19.59 19L14.87 14.29V13.5L14.59 13.22C13.66 13.72 12.61 14 11.5 14A6.5 6.5 0 0 1 5 7.5A6.5 6.5 0 0 1 11.5 1A6.5 6.5 0 0 1 18 7.5A6.5 6.5 0 0 1 11.5 14M11.5 3A4.5 4.5 0 0 0 7 7.5A4.5 4.5 0 0 0 11.5 12A4.5 4.5 0 0 0 16 7.5A4.5 4.5 0 0 0 11.5 3Z"
  ],
  albums: ["M6 4H17C18.1 4 19 4.9 19 6V18C19 19.1 18.1 20 17 20H6C4.9 20 4 19.1 4 18V6C4 4.9 4.9 4 6 4Z", "M9 4V12L12 10L15 12V4"],
  media: ["M12 3L18 7V17L12 21L6 17V7L12 3Z", "M10 9L15 12L10 15V9Z"],
  stats: ["M5 19V12", "M10 19V8", "M15 19V11", "M20 19V5", "M5 7L10 4L15 8L20 3"],
  places: ["M12 21C12 21 18 15.5 18 10C18 6.686 15.314 4 12 4C8.686 4 6 6.686 6 10C6 15.5 12 21 12 21Z", "M10 8L14 10L12 14L10 8Z"],
  calendar: [
    "M7 3V6",
    "M17 3V6",
    "M4 8H20",
    "M5 5H19C19.552 5 20 5.448 20 6V19C20 19.552 19.552 20 19 20H5C4.448 20 4 19.552 4 19V6C4 5.448 4.448 5 5 5Z",
    "M9 12H15",
    "M9 16H13",
  ],
  moments: ["M4 12H20", "M8 12V12", "M12 12V12", "M16 12V12", "M12 5L13.5 8.5L17 10L13.5 11.5L12 15L10.5 11.5L7 10L10.5 8.5L12 5Z"],
  folders: ["M4 8C4 6.895 4.895 6 6 6H10L12 8H18C19.105 8 20 8.895 20 10V17C20 18.105 19.105 19 18 19H6C4.895 19 4 18.105 4 17V8Z"],
  settings: [
    "M12 8C14.209 8 16 9.791 16 12C16 14.209 14.209 16 12 16C9.791 16 8 14.209 8 12C8 9.791 9.791 8 12 8Z",
    "M12 3V5",
    "M12 19V21",
    "M21 12H19",
    "M5 12H3",
    "M18.364 5.636L16.95 7.05",
    "M7.05 16.95L5.636 18.364",
    "M18.364 18.364L16.95 16.95",
    "M7.05 7.05L5.636 5.636",
  ],
};

export default {
  name: "IconMokosh",
  props: {
    // The icon name, supplied by Vuetify as the part after "mokosh:".
    icon: {
      type: String,
      default: "",
    },
  },
  computed: {
    paths() {
      return PATHS[this.icon] || [];
    },
  },
};
</script>