<template>
  <LoadingSlider v-if="loading" />
  <div class="sidebar-screen" v-else>
    <h1>Provision a new certificate</h1>

    <label>Domains</label>

    <p class="explain">
      Add or remove the domains you wish to attach to this certificate.
    </p>

    <div class="domains">
      <div class="domain" v-for="domain in domains">
        <span>{{ domain }}</span>
        <div class="delete" v-if="!busy" @click="removeDomain(domain)"></div>
        <Spinner v-else class="spinner" />
      </div>
    </div>

    <div class="input-and-button">
      <input
        v-model="domain"
        type="text"
        placeholder="example.com"
        class="defined"
        @keypress.enter="addDomain"
        :disabled="busy"
      />

      <Button
        text="Add domain"
        :disabled="!domain || busy"
        class="blue"
        @click="addDomain"
      />
    </div>

    <label>Certificate source</label>
    <p class="explain">Where is this certificate coming from?</p>
    <div class="options">
      <div
        class="option"
        :class="{ active: !upload, disabled: busy }"
        @click="setUpload(false)"
      >
        LetsEncrypt
      </div>
      <div
        class="option"
        :class="{ active: upload, disabled: busy }"
        @click="setUpload(true)"
      >
        Upload
      </div>
    </div>

    <template v-if="upload">
      <Button
        text="Upload certificate"
        @click="busy = !busy"
        :busy="busy"
        class="green final"
        :disabled="!domains.length"
      />
    </template>

    <template v-else>
      <label>DNS configuration</label>
      <p class="explain">
        For each of these domains, ensure that they have the following DNS
        settings. Otherwise, LetsEncrypt will fail.
      </p>

      <table class="dns">
        <tr>
          <th>Type</th>
          <th>Name</th>
          <th>Value</th>
          <th>TTL</th>
        </tr>
        <template v-for="node in nodes">
          <tr>
            <td>A</td>
            <td>@</td>
            <td>{{ node.address }}</td>
            <td>30</td>
          </tr>
          <tr>
            <td>A</td>
            <td>*</td>
            <td>{{ node.address }}</td>
            <td>30</td>
          </tr>
        </template>
      </table>

      <Button
        text="Provision certificate"
        @click="getAcmeCert"
        :busy="busy"
        class="green final"
        :disabled="!domains.length"
      />
    </template>
  </div>
</template>

<script>
import Spinner from "@/components/Spinner";

export default {
  components: { Spinner },

  data() {
    return {
      domain: "",
      domains: [],
      loading: true,
      busy: false,
      upload: false,
      nodes: []
    };
  },

  mounted() {
    this.load();
  },

  methods: {
    setUpload(bool) {
      if (this.busy) return;
      this.upload = bool;
    },

    addDomain() {
      if (!this.domain) return;
      this.domains.push(this.domain);
      this.domain = "";
    },

    async load() {
      const res = await this.$api.get("/nodes");
      if (res.status !== 200) return alert(res.data);
      const nodes = res.data.filter(node =>
        node.node_roles.includes("LOAD_BALANCER")
      );
      this.nodes = nodes;
      this.loading = false;
    },

    removeDomain(domainToRemove) {
      this.domains = this.domains.filter(domain => domain !== domainToRemove);
    },

    async getAcmeCert() {
      this.busy = true;

      const body = {
        auto_renew: true,
        domains: this.domains,
        namespace: this.$namespace()
      };

      const res = await this.$api.post("/certificate", body);
      if (res.status !== 201) {
        this.busy = false;
        alert(res.data);
        return;
      }
      console.log(res.data);
      await this.renewCerts();
      this.busy = false;
      this.$reload();
      this.$push("/certificates");
    },

    // Perform a renewal on all of the certificates in the store. This will only
    // realistically perform a single certificate retrieval but it ensures that
    // LetsEncrypt provides us with the correct certificate for that domain.
    async renewCerts() {
      const { status } = await this.$api.post("/certificates/renew");
    }
  },

  watch: {
    domain(value) {
      this.domain = value
        .toLowerCase()
        .split(" ")
        .join("");
    }
  }
};
</script>

<style lang="scss" scoped>
.domains {
  width: 100%;
  text-align: left;
  border-radius: 4px;
  overflow: hidden;
  cursor: default;

  .domain {
    display: flex;
    justify-content: space-between;
    align-items: center;

    transition: background-color 0.2s;
    padding: 10px;
    &:not(:last-of-type) {
      border-bottom: solid 1px #ddd;
    }
    &:hover {
      background-color: #fafafa;

      .delete {
        opacity: 1;
        transform: scale(1);
      }
    }
  }

  .delete {
    width: 20px;
    height: 20px;
    background-image: url("~@/assets/icon/exit.svg");
    background-size: 12px;
    background-position: center;
    background-repeat: no-repeat;
    opacity: 0;
    transform: scale(0);
    cursor: pointer;
    transition: opacity 0.2s, transform 0.2s;
  }
}

table.dns {
  text-align: left;
  width: 100%;
  font-size: 15px;
  margin-top: 10px;

  th,
  td {
    padding: 5px;
  }

  th {
    font-weight: bold;
  }

  td {
    font-family: "Source Code Pro", sans-serif;
  }
}

label:not(:first-of-type) {
  margin-top: 30px;
}

.input-and-button {
  margin-top: 20px;
}

.spinner {
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 3px solid transparent;
  border-radius: 50%;
  border-top-color: #feca57;
  border-bottom-color: #feca57;
}

.options {
  display: flex;

  .option {
    padding: 10px 14px;
    border: solid 1px #ddd;
    border-radius: 4px;

    &:not(:last-of-type) {
      margin-right: 10px;
    }
    transition: all 0.2s;

    cursor: not-allowed;
    opacity: 0.7;

    &:not(.disabled) {
      opacity: 1;

      cursor: pointer;
      &:hover {
        transform: scale(1.05);
      }
      &:active {
        transform: scale(0.95);
      }
    }

    &.active {
      border: solid 1px #1dd1a1;
      color: #1dd1a1;
    }
  }
}

.button.final {
  margin-top: 30px;
}
</style>
