<template>
  <div>
    <h2>Admin user</h2>
    <p class="subheader">
      Create the administrator user account that you will use to sign in. You
      can always change this later. This provides complete access to your entire
      cluster, so please ensure it is as secure as possible.
    </p>

    <div class="form">
      <div class="profile"></div>

      <label>Username</label>
      <input
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
      <input type="password" name="password" placeholder="Password" />
    </div>

    <Button class="green" text="Sign up" />
  </div>
</template>

<script>
import Button from "@/components/Button";

export default {
  components: { Button },

  data() {
    return {
      user: {
        username: "",
        email: "",
        password: ""
      }
    };
  },

  computed: {
    // Generate a placeholder email based on the domain name and the user name
    // specified.
    placeholderEmail() {
      const { username } = this.user;
      const [domain] = document.location.host.split(":");
      return `${username ? username : "admin"}@${domain}`;
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

  methods: {}
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
    background-color: #000;
    margin: 0 auto;
  }

  label {
    margin-top: 20px;
    margin-bottom: 5px;
  }
}
</style>
