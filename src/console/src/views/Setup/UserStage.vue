<template>
  <div>
    <h2>Admin user</h2>
    <p class="subheader">
      Create the administrator user account that you will use to sign in. You
      can always change this later. This provides complete access to your entire
      cluster, so please ensure it is as secure as possible.
    </p>

    <div class="form">
      <div
        class="profile"
        :style="{ backgroundImage: `url('${userProfileSrc}')` }"
        :class="{ disabled: busy }"
        @click="clickProfile()"
      >
        <div class="overlay" :class="{ disabled: busy }">
          <img v-if="user.profile" src="@/assets/icon/trash-white.svg" />
          <img v-else src="@/assets/icon/file-add-white.svg" />
        </div>
      </div>

      <input
        :style="{ display: 'none' }"
        type="file"
        accept="image/*"
        ref="profileInput"
        @change="e => (user.profile = e.target.files[0])"
      />

      <label>Username</label>
      <input
        :disabled="busy"
        ref="usernameField"
        v-model="user.username"
        type="text"
        name="username"
        maxlength="20"
        placeholder="Username"
      />

      <label>Email address</label>
      <input
        :disabled="busy"
        v-model="user.email"
        type="email"
        name="email"
        maxlength="80"
        placeholder="Email address"
      />

      <label>Password</label>
      <input
        :disabled="busy"
        v-model="user.password"
        type="password"
        name="password"
        placeholder="Password"
      />

      <label>Confirm password</label>
      <input
        :disabled="busy"
        v-model="user.confirmPassword"
        type="password"
        name="password"
        placeholder="Password"
      />
    </div>

    <Button
      class="green"
      text="Continue"
      :disabled="!validUser"
      :busy="busy"
      @click="createUser"
    />
  </div>
</template>

<script>
import Button from "@/components/Button";
import defaultProfileImage from "@/assets/icon/blank-profile.svg";

export default {
  components: { Button },

  data() {
    return {
      user: {
        profile: null, // An image
        username: "",
        email: "",
        password: "",
        confirmPassword: ""
      },

      busy: false // Whether or not processing is taking place
    };
  },

  mounted() {
    this.$refs.usernameField.focus(); // Focus the username field on start.
  },

  computed: {
    // Return the source URL for the image.
    userProfileSrc() {
      const { profile } = this.user;
      return profile ? URL.createObjectURL(profile) : defaultProfileImage;
    },

    // Whether or not the user settings, as they are, are valid.
    validUser() {
      return true;
    }
  },

  watch: {
    // Sanitize the user fields on input.
    user: {
      deep: true,
      handler() {
        const { username, email } = this.user;

        this.user.username = username
          .toLowerCase() // Convert completely to lowercase
          .split(" ") // Remove all spaces
          .join("")
          .split(/[^a-zA-Z0-9]/) // Only allow alphanumeric characters
          .join("");

        this.user.email = email
          .toLowerCase() // Convert completely to lowercase
          .split(" ") // Remove all spaces
          .join("");
      }
    }
  },

  methods: {
    // When the user profile is clicked on.
    clickProfile() {
      if (this.busy) return;

      this.$refs.profileInput.value = ""; // Clear the file input first
      if (this.user.profile) this.user.profile = null;
      else this.$refs.profileInput.click();
    },

    // Perform the user creation operation and make the request to the API.
    async createUser() {
      this.busy = !this.busy;
    }
  }
};
</script>

<style lang="scss" scoped>
.form {
  display: flex;
  flex-direction: column;
  max-width: 300px;
  margin: 30px auto;
  background-color: rgba(255, 255, 255, 0.5);
  border-radius: 4px;
  padding: 40px;
  text-align: left;

  .profile {
    width: 90px;
    height: 90px;
    border-radius: 1000px;
    background-color: #c8d6e5;
    margin: 0 auto;

    background-size: cover;
    background-position: center;
    background-repeat: no-repeat;
    overflow: hidden;

    transition: transform 0.2s;

    &:not(.disabled) {
      &:active {
        transform: scale(0.9);
      }
    }

    .overlay {
      width: 100%;
      height: 100%;
      background-color: rgba(0, 0, 0, 0.3);
      opacity: 0;

      display: flex;
      align-items: center;
      justify-content: center;

      img {
        width: 30px;
      }

      transition: opacity 0.2s;

      &:not(.disabled) {
        cursor: pointer;
        &:hover {
          opacity: 1;
        }
      }
    }
  }

  label {
    margin-top: 20px;
    margin-bottom: 8px;
    font-weight: bold;
  }
}
</style>
