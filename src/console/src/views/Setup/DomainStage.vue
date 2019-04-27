<template>
  <div>
    <h2>Domain &amp; security</h2>
    <p class="subheader">
      The domain name that you use to access Orbit. This includes the main
      dashboard and also where any git repositories will be hosted. This should
      be secured before adding your first user account.
    </p>

    <div class="domain-group">
      <input
        name="domain"
        type="text"
        placeholder="orbit.example.com"
        autocomplete="off"
        autocorrect="off"
        autocapitalize="off"
        spellcheck="false"
        v-model="domain"
        ref="input"
        size="30"
        :disabled="busy"
        @keyup.enter="addDomain"
      />
    </div>

    <a
      href="#"
      @click.prevent="domain = urlDomain"
      v-if="urlDomain && urlDomain !== domain"
      >Use URL Domain ({{ urlDomain }})</a
    >

    <div class="certificate-group">
      <h4>Add an SSL certificate</h4>
      <p class="subheader">
        This will encrypt all of your communications to Orbit. How would you
        like to add an SSL certificate?
      </p>
      <div class="options" :class="{ disabled: busy }">
        <div
          class="option blue"
          :class="{ selected: certMethod === 'upload' }"
          @click="!busy && (certMethod = 'upload')"
        >
          <img />

          <h5>Upload</h5>
          <p>Upload a certificate from your computer</p>
        </div>

        <div
          class="option green"
          :class="{ selected: certMethod === 'letsencrypt' }"
          @click="!busy && (certMethod = 'letsencrypt')"
        >
          <img />

          <h5>LetsEncrypt</h5>
          <p>Obtain a free certificate from LetsEncrypt</p>
        </div>

        <div
          class="option red"
          :class="{ selected: certMethod === 'none' }"
          @click="!busy && (certMethod = 'none')"
        >
          <img />

          <h5>None</h5>
          <p>Don't use an SSL certificate</p>
        </div>
      </div>

      <div class="warning" v-if="certMethod === 'none'">
        <h5 class="error">Not recommended</h5>
        <p>
          Not having a certificate for the domain means that anything you do
          will be sent over plain text. That means anybody could steal your
          username and password.
        </p>
      </div>
    </div>

    <div class="button-group">
      <Button
        :class="{
          green: certMethod === 'letsencrypt',
          blue: certMethod === 'upload',
          red: certMethod === 'none'
        }"
        :disabled="!validCert || !validDomain"
        :busy="busy"
        text="Continue"
        @click="addDomain"
      />
    </div>
  </div>
</template>

<script>
import validator from "validator";

import Button from "@/components/Button";
import Spinner from "@/components/Spinner";

export default {
  components: {
    Button,
    Spinner
  },

  data() {
    return {
      domain: "orbit.samholmes.net",
      busy: false,

      certMethod: "letsencrypt",
      certificateFile: null, // The certificate file if method is "upload"
      privateKeyFile: null // The private key file if the method is "upload"
    };
  },

  mounted() {
    this.focus();
  },

  computed: {
    // urlDomain returns the domain name of the current page, if there is
    // one.
    urlDomain() {
      const [domain] = document.location.host.split(":"); // Strip the port
      if (validator.isFQDN(domain)) return domain;
    },

    // This checks whether the domain provided is actually valid.
    validDomain() {
      return validator.isFQDN(this.domain);
    },

    // This checks whether the SSL certificate option is valid.
    validCert() {
      return true;
    }
  },

  methods: {
    // Add the domain name to the store.
    async addDomain() {
      if (!this.validCert || !this.validDomain || this.busy) return;
      this.busy = true;

      // Globally scoped variables that we retrieve over the course of the API
      // access.
      let routerID, certificateID, namespaceID;

      /**
       * Retrieve the orbit-system namespace ID.
       */
      {
        const res = await this.$api.get("/namespaces", { redirect: false });
        if (res.status !== 200) {
          this.busy = false;
          alert(res.data);
          return;
        }
        namespaceID = res.data.find(ns => ns.name === "orbit-system").id;
      }

      /**
       * Create the router.
       */
      {
        const body = {
          domain: this.domain,
          namespace_id: namespaceID,
          app_id: "console"
        };
        const opts = { redirect: false };
        const res = await this.$api.post("/router", body, opts);
        if (res.status !== 201) {
          this.busy = false;
          alert(res.data);
          return;
        }
        routerID = res.data;
        console.log(`Created router with ID ${routerID}`);
      }

      // If we are not adding a certificate, we don't need to do anything
      // further.
      if (this.certMethod === "none") {
        this.busy === false;
        return;
      }

      /**
       * Create the certificate.
       */
      {
        const body =
          this.certMethod === "letsencrypt"
            ? { auto_renew: true }
            : this.certMethod === "upload"
            ? {
                full_chain: this.certificateFile,
                private_key: this.privateKeyFile
              }
            : {};
        body.namespace_id = namespaceID; // Ensure we set the correct namespace.
        const opts = { redirect: false };
        const res = await this.$api.post("/certificate", body, opts);
        if (res.status !== 201) {
          this.busy = false;
          alert(res.data);
          return;
        }
        certificateID = res.data;
        console.log(`Created certificate with ID ${certificateID}`);
      }

      /**
       * Add the certificate ID to the existing router object.
       */
      {
        const path = `/router/${routerID}`;
        const body = { certificate_id: certificateID };
        const opts = { redirect: false };
        const res = await this.$api.put(path, body, opts);
        if (res.status !== 200) {
          this.busy = false;
          alert(res.data);
          return;
        }
      }

      // TODO: Wait for load balancer update and then continue on.
    },

    // Focus on the input element.
    async focus() {
      await this.$nextTick();
      this.$refs.input.focus();
    }
  }
};
</script>

<style lang="scss" scoped>
.domain-group {
  margin-top: 30px;
  align-items: center;
  justify-content: center;
  display: flex;
}

.certificate-group {
  margin-top: 40px;
  display: inline-flex;
  flex-direction: column;
  align-items: center;

  padding: 30px;
  border-radius: 4px;
  background-color: rgba(0, 0, 0, 0.03);
  box-shadow: inset 0 2px 6px 0 rgba(0, 0, 0, 0.05);

  max-width: 700px;

  h4 {
    font-size: 18px;
    font-weight: 500;
  }

  p.subheader {
    font-size: 16px;
    max-width: 500px;
    line-height: 1.4rem;
    margin-top: 10px;
  }

  .options {
    margin-top: 20px;
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    grid-gap: 20px;

    @media (max-width: 600px) {
      grid-template-columns: 1fr;
    }

    .option {
      border: solid 1px #fff;
      border-radius: 4px;
      padding: 14px;
      background-color: #fff;

      cursor: pointer;

      img {
        width: 40px;
        height: 40px;
      }

      h5 {
        font-size: 17px;
        font-weight: 500;
        margin-top: 14px;

        transition: color 0.2s;
      }

      p {
        font-size: 14px;
        line-height: 1.2rem;
        margin-top: 6px;
        opacity: 0.7;
      }

      transition: border-color 0.2s, transform 0.2s;

      &:hover {
        transform: scale(1.05);
      }

      &:active {
        transform: scale(0.95);
      }

      &.green {
        &.selected {
          border: solid 1px #1dd1a1;
          h5 {
            color: #1dd1a1;
          }
        }
      }

      &.blue {
        &.selected {
          border: solid 1px #0abde3;
          h5 {
            color: #0abde3;
          }
        }
      }

      &.red {
        &.selected {
          border: solid 1px #ee5253;
          h5 {
            color: #ee5253;
          }
        }
      }
    }

    // Disabled occurs when there is an operation processing, and we need to
    // mimic input being disabled.
    &.disabled {
      .option {
        cursor: default;
        opacity: 0.6;

        &:hover {
          transform: none;
        }

        &:active {
          transform: none;
        }
      }
    }
  }

  .warning {
    h5 {
      font-weight: 500;
      display: inline-block;
      padding: 5px 14px;
      margin-top: 20px;
      text-transform: uppercase;
    }

    p {
      font-size: 16px;
      max-width: 500px;
      margin-top: 10px;
      line-height: 1.5rem;
      color: #ee5253;
    }
  }
}

a {
  margin-top: 15px;
  display: block;
}

.button-group {
  margin-top: 30px;
}
</style>
