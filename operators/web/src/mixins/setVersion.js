import Vue from "vue";

export const setVersion = {
  data() {
    return {
      gitVersion: ""
    }
  },
  async beforeMount() {
    await this.fetchVersion();
  },
  methods: {
    async fetchVersion() {
      const response = await Vue.axios.get("/version");
      if (response.data == null) {
        console.error("Failed getting git version info");
        return;
      }

      console.log("Git version:", response.data);
      this.gitVersion = response.data.gitVersion;
    }
  }
}