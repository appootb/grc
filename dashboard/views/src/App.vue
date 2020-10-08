<template>
  <v-app>
    <v-navigation-drawer app expand-on-hover>
      <v-list>
        <!-- <v-list-item class="px-2">
          <v-list-item-avatar>
            <v-img src="logo.png"></v-img>
          </v-list-item-avatar>
        </v-list-item> -->

        <v-list-item link @click="swithDarkMode">
          <v-list-item-content>
            <v-list-item-title class="title"><v-img src="logo.png"></v-img></v-list-item-title>
            <v-list-item-subtitle>Remote Configuration</v-list-item-subtitle>
          </v-list-item-content>
        </v-list-item>
      </v-list>
      <v-divider />
      <v-list nav dense>
        <template v-for="item in menus">
          <v-list-item link :key="item.path" :to="item.path">
            <v-list-item-icon><v-icon>{{ item.icon }}</v-icon></v-list-item-icon>
            <v-list-item-title>{{ item.title }}</v-list-item-title>
          </v-list-item>
        </template>
      </v-list>
    </v-navigation-drawer>

    <v-app-bar app>
      <v-toolbar-title>Remote Configuration - Dashboard</v-toolbar-title>
      <v-spacer />
      <v-avatar><img src="logo.png" /></v-avatar>
    </v-app-bar>

    <v-main>
      <v-container fluid>
        <router-view />
      </v-container>
    </v-main>

    <v-footer app>
      <v-spacer/>
      <div>grc dashboard &copy; {{ new Date().getFullYear() }} appootb.dev</div>
    </v-footer>
  </v-app>
</template>

<script>
export default {
  data: () => ({
    dark: false,
    menus: [
      { icon: 'mdi-tune', title: 'Configurations', path: '/' },
    ]
  }),

  created() {
    if (localStorage.dark === 'true') {
      this.$vuetify.theme.dark = true
    } else {
      this.$vuetify.theme.dark = false
    }
  },

  methods: {
    swithDarkMode() {
      let dark = localStorage.dark === 'true'
      localStorage.dark = !dark
      this.$vuetify.theme.dark = !dark
    },
  }
}
</script>