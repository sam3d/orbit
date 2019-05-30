<template>
  <div>
    <h2>Admin user</h2>
    <p class="subheader">
      Create the administrator user account that you will use to sign in. You
      can always change this later. This provides complete access to your entire
      cluster, so please ensure it is as secure as possible.
    </p>

    <form class="form" @submit.prevent="createUser">
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
        style="display: none"
        type="file"
        accept="image/*"
        ref="profileInput"
        @change="e => (user.profile = e.target.files[0])"
      />

      <label>Name</label>
      <input
        :disabled="busy"
        ref="nameField"
        v-model="user.name"
        type="text"
        name="name"
        placeholder="Name"
      />

      <label>Username</label>
      <input
        :disabled="busy"
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

      <input type="submit" style="display: none;" />
    </form>

    <div class="error" v-if="error">{{ error }}</div>
    <br />
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
import defaultProfileImage from "@/assets/icon/blank-profile.svg";

export default {
  data() {
    return {
      user: {
        profile: null, // An image
        name: "",
        username: "",
        email: "",
        password: "",
        confirmPassword: ""
      },

      error: "", // If there was an error processing
      busy: false // Whether or not processing is taking place
    };
  },

  mounted() {
    this.$refs.nameField.focus(); // Focus the name field on start.
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
      if (this.busy) return;
      this.busy = true;

      this.error = ""; // Clear error on resubmit

      // Construct the request.
      const body = new FormData();
      body.append("name", this.user.name);
      body.append("username", this.user.username);
      body.append("password", this.user.password);
      body.append("email", this.user.email);

      if (this.user.profile) body.append("profile", this.user.profile);

      // Construct and submit the request.
      const headers = { "Content-Type": "multipart/form-data" };
      const opts = { redirect: false, headers };
      const res = await this.$api.post("/user", body, opts);

      // If there is an error, handle it.
      if (res.status !== 201) {
        this.busy = false;
        this.error = res.data;
        return;
      }

      // Otherwise, emit the completion of this stage.
      this.$emit("complete");
    }
  },

  computed: {
    // Return the source URL for the image.
    userProfileSrc() {
      const { profile } = this.user;
      return profile ? URL.createObjectURL(profile) : defaultProfileImage;
    },

    validUser() {
      const { user } = this;

      // Return the overall error.
      return (
        user.name &&
        user.username &&
        user.email &&
        user.password &&
        user.confirmPassword &&
        user.password === user.confirmPassword
      );
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

.button {
  display: inline-block !important;
}

.error {
  display: inline-block;
  margin-bottom: 30px;
}
</style>
