<template>
  <v-container id="auth-register" theme="login" fluid fill-height class="auth-register wallpaper background-welcome pa-6" :style="wallpaper()">
    <v-theme-provider theme="login">
      <v-row id="auth-layout" class="auth-layout">
        <v-col cols="12" sm="9" md="6" lg="5" xl="3">
          <v-form ref="form" class="auth-register-form" accept-charset="UTF-8" @submit.prevent="onRegister">
            <v-card id="auth-register-box" class="elevation-12 auth-register-box pa-1 blur-7">
              <v-card-text>
                <p-auth-header></p-auth-header>
                <v-spacer></v-spacer>
                <v-row align="start" dense>
                  <v-col cols="12" class="pb-1">
                    <v-text-field
                      id="register-name"
                      v-model="displayName"
                      :disabled="loading"
                      :placeholder="$gettext('Display Name')"
                      :autofocus="true"
                      name="display-name"
                      variant="solo"
                      density="comfortable"
                      type="text"
                      hide-details
                      autocorrect="off"
                      autocapitalize="words"
                      autocomplete="name"
                      class="input-display-name text-selectable"
                      prepend-inner-icon="mdi-account-outline"
                      @keyup.enter="onRegister"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="12" class="pb-1">
                    <v-text-field
                      id="register-username"
                      v-model="username"
                      :disabled="loading"
                      :placeholder="$gettext('Username')"
                      name="username"
                      variant="solo"
                      density="comfortable"
                      type="text"
                      hide-details
                      autocorrect="off"
                      autocapitalize="none"
                      autocomplete="username"
                      class="input-username text-selectable"
                      prepend-inner-icon="mdi-account"
                      @keyup.enter="onRegister"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="12" class="pb-1">
                    <v-text-field
                      id="register-email"
                      v-model="email"
                      :disabled="loading"
                      :placeholder="$gettext('E-Mail Address')"
                      name="email"
                      variant="solo"
                      density="comfortable"
                      type="email"
                      hide-details
                      autocorrect="off"
                      autocapitalize="none"
                      autocomplete="email"
                      class="input-email text-selectable"
                      prepend-inner-icon="mdi-email-outline"
                      @keyup.enter="onRegister"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="12" class="pb-1">
                    <v-text-field
                      id="register-password"
                      v-model="password"
                      :disabled="loading"
                      :type="showPassword ? 'text' : 'password'"
                      :placeholder="$gettext('Password')"
                      name="new-password"
                      variant="solo"
                      density="comfortable"
                      hide-details
                      autocorrect="off"
                      autocapitalize="none"
                      autocomplete="new-password"
                      class="input-password text-selectable"
                      :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
                      prepend-inner-icon="mdi-lock-outline"
                      @click:append-inner="showPassword = !showPassword"
                      @keyup.enter="onRegister"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="12" class="pb-1">
                    <v-text-field
                      id="register-password-confirm"
                      v-model="passwordConfirm"
                      :disabled="loading"
                      :type="showPassword ? 'text' : 'password'"
                      :placeholder="$gettext('Confirm Password')"
                      name="confirm-password"
                      variant="solo"
                      density="comfortable"
                      hide-details
                      autocorrect="off"
                      autocapitalize="none"
                      autocomplete="new-password"
                      class="input-password-confirm text-selectable"
                      prepend-inner-icon="mdi-lock-check-outline"
                      @keyup.enter="onRegister"
                    ></v-text-field>
                  </v-col>
                  <v-col cols="12" class="auth-actions">
                    <div class="action-buttons auth-buttons pb-1 d-flex ga-3 align-center justify-center">
                      <v-btn :block="$vuetify.display.xs" color="highlight" variant="outlined" class="action-login opacity-80" @click.stop.prevent="onLogin">
                        {{ $gettext(`Sign in`) }}
                      </v-btn>
                      <v-btn
                        :disabled="registerDisabled"
                        :block="$vuetify.display.xs"
                        color="highlight"
                        variant="flat"
                        class="action-confirm"
                        @click.stop.prevent="onRegister"
                      >
                        {{ $gettext(`Create Account`) }}
                        <v-icon :icon="$config.isRtl() ? 'mdi-chevron-left' : 'mdi-chevron-right'" end></v-icon>
                      </v-btn>
                    </div>
                  </v-col>
                </v-row>
              </v-card-text>
            </v-card>
          </v-form>
        </v-col>
      </v-row>
      <p-auth-footer></p-auth-footer>
    </v-theme-provider>
  </v-container>
</template>

<script>
import PAuthHeader from "component/auth/header.vue";
import PAuthFooter from "component/auth/footer.vue";

export default {
  name: "PPageRegister",
  components: {
    PAuthHeader,
    PAuthFooter,
  },
  data() {
    return {
      loading: false,
      displayName: "",
      username: "",
      email: "",
      password: "",
      passwordConfirm: "",
      showPassword: false,
      wallpaperUri: this.$config.values.wallpaperUri,
    };
  },
  computed: {
    registerDisabled() {
      if (this.loading) {
        return true;
      }

      return this.username.trim() === "" || this.password.trim() === "" || this.passwordConfirm.trim() === "";
    },
  },
  mounted() {
    this.$view.enter(this, this.$refs?.form, 'input[value=""], button.action-confirm');
  },
  unmounted() {
    this.$view.leave(this);
  },
  methods: {
    wallpaper() {
      if (this.wallpaperUri) {
        return `background-image: url(${this.wallpaperUri});`;
      }

      return "";
    },
    onLogin() {
      this.$router.push({ name: "login" });
    },
    onRegister() {
      const username = this.username.trim();
      const password = this.password.trim();
      const passwordConfirm = this.passwordConfirm.trim();

      if (username === "" || password === "") {
        return;
      }

      if (password !== passwordConfirm) {
        this.$notify.warn(this.$gettext("Passwords do not match"));
        return;
      }

      this.loading = true;

      this.$session
        .register({
          UserName: username,
          UserEmail: this.email.trim(),
          DisplayName: this.displayName.trim(),
          Password: password,
        })
        .then(() => this.$session.login(username, password))
        .then(() => {
          const homeRoute = this.$router.resolve({ name: this.$session.getDefaultRoute() });
          this.$session.followLoginRedirectUrl(homeRoute.href);
        })
        .catch((e) => {
          const msg = e?.response?.data?.message || e?.response?.data?.error;
          if (msg) {
            this.$notify.error(msg);
          } else {
            this.$notify.error(this.$gettext("Registration failed, please try again"));
          }
          this.loading = false;
        });
    },
  },
};
</script>
