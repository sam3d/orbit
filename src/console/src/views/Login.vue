<template>
  <div class="container">
    <img src="@/assets/logo/gradient.svg" class="logo" />

    <form class="form" @submit.prevent="login">
      <div class="profile" :style="profileStyle"></div>

      <label>Username or email address</label>
      <input
        type="text"
        name="username"
        ref="usernameField"
        placeholder="Username or email address"
        v-model="user.identifier"
        :disabled="busy"
        @change="updateProfile"
      />

      <label>Password</label>
      <input
        ref="passwordField"
        v-model="user.password"
        type="password"
        name="password"
        placeholder="Password"
        :disabled="busy"
      />

      <!-- Blank submit to ensure the form will process a submit -->
      <input type="submit" style="display: none;" />
    </form>

    <Button
      text="Log in"
      class="purple"
      @click="login"
      :busy="busy"
      :disabled="!valid"
    />
  </div>
</template>

<script>
import defaultProfile from "@/assets/icon/blank-profile.svg";

export default {
  data() {
    return {
      profile: defaultProfile, // The profile image URL

      // The data to be submitted.
      user: {
        identifier: "", // The current identifier for the user
        password: ""
      },

      busy: false // Whether or not we're processing data
    };
  },

  mounted() {
    // Don't show this page if the user is logged in.
    if (this.$store.state.token) this.$router.push("/");

    // Focus the correct field.
    this.$refs.usernameField.focus();
  },

  methods: {
    async updateProfile() {
      const path = `/user/${this.user.identifier}/profile`;
      const res = await this.$api.get(path);
      if (res.status !== 200) {
        // No profile image or user not found, set it to the default.
        this.profile = defaultProfile;
        return;
      }

      // Otherwise, set the profile data to the correct URL.
      this.profile = "/api" + path;
    },

    async login() {
      if (this.busy || !this.valid) return;
      this.busy = true;

      // Make the request.
      const opts = { redirect: false };
      const res = await this.$api.post("/user/login", this.user, opts);
      if (res.status !== 200) {
        // Stop processing and show error.
        this.busy = false;
        alert(res.data);

        // Focus on the correct field depending on the error.
        await this.$nextTick();
        const field = res.data.includes("password")
          ? "passwordField"
          : "usernameField";
        this.$refs[field].focus();

        return;
      }

      // Set the token data and retrieve the user data.
      localStorage.setItem("token", res.data);
      await this.$store.dispatch("updateUser");

      // Redirect to the dashboard.
      this.$router.push("/");
    }
  },

  computed: {
    profileStyle() {
      return { backgroundImage: `url("${this.profile}")` };
    },

    // Simply ensure each field has enough data in it.
    valid() {
      return this.user.identifier && this.user.password.length >= 3;
    }
  },

  watch: {
    user: {
      handler() {
        this.user.identifier = this.user.identifier
          .toLowerCase()
          .split(" ")
          .join("");
      },
      deep: true
    }
  }
};
</script>

<style lang="scss" scoped>
.container {
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  padding: 20px;

  height: 100vh;
  width: 100%;
}

@mixin fadeIn($delay: 0s) {
  animation: fadeIn 0.9s $delay forwards ease;
  opacity: 0;

  @keyframes fadeIn {
    from {
      transform: scale(0.9);
      opacity: 0;
    }

    to {
      transform: scale(1);
      opacity: 1;
    }
  }
}

.logo {
  width: 80px;
  @include fadeIn();
}

.form {
  background-color: rgba(255, 255, 255, 0.5);
  padding: 40px;
  border-radius: 4px;

  width: 100%;
  max-width: 380px;

  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  margin: 50px 0;

  @include fadeIn(0.3s);

  label {
    font-weight: bold;
    margin-top: 25px;
    margin-bottom: 8px;
  }

  .profile {
    width: 90px;
    height: 90px;
    border-radius: 90px;
    background-color: rgba(0, 0, 0, 0.1);
    align-self: center;
    margin-bottom: 15px;

    background-size: cover;
    background-position: center;
    background-repeat: no-repeat;
  }
}

.button {
  @include fadeIn(0.6s);
}
</style>
