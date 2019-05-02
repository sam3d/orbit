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
        @click="clickProfile()"
      >
        <div class="overlay">
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
        ref="usernameField"
        v-model="user.username"
        type="text"
        name="username"
        placeholder="admin"
      />

      <label>Email address</label>
      <input
        v-model="user.email"
        type="email"
        name="email"
        :placeholder="placeholderEmail"
      />

      <label>Password</label>
      <input
        v-model="user.password"
        type="password"
        name="password"
        placeholder="Password"
      />

      <label>Confirm password</label>
      <input
        v-model="user.confirmPassword"
        type="password"
        name="password"
        placeholder="Password"
      />
    </div>

    <Button class="green" text="Continue" />
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
      }
    };
  },

  mounted() {
    this.$refs.usernameField.focus(); // Focus the username field on start.
  },

  computed: {
    // Generate a placeholder email based on the domain name and the user name
    // specified.
    placeholderEmail() {
      const { username } = this.user;
      const [domain] = document.location.host.split(":");
      return `${username ? username : "admin"}@${domain}`;
    },

    // Return the source URL for the image.
    userProfileSrc() {
      const { profile } = this.user;
      return profile ? URL.createObjectURL(profile) : defaultProfileImage;
    }
  },

  watch: {
    // Sanitize the user fields on input.
    user: {
      deep: true,
      handler() {
        const { username, email } = this.user;

        this.user.username = username
          .toLowerCase()
          .split(" ")
          .join("");

        this.user.email = email
          .toLowerCase()
          .split(" ")
          .join("");
      }
    }
  },

  methods: {
    // When the user profile is clicked on.
    clickProfile() {
      this.$refs.profileInput.value = ""; // Clear the file input first
      if (this.user.profile) this.user.profile = null;
      else this.$refs.profileInput.click();
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

    &:active {
      transform: scale(0.9);
    }

    .overlay {
      width: 100%;
      height: 100%;
      background-color: rgba(0, 0, 0, 0.3);
      opacity: 0;

      display: flex;
      align-items: center;
      justify-content: center;

      transition: opacity 0.2s;
      cursor: pointer;

      img {
        width: 30px;
      }

      &:hover {
        opacity: 1;
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
