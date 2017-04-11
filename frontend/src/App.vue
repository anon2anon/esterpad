<template>
  <div id="app" class="container">
    <md-sidenav class="md-left" ref="sidenav">
      <md-toolbar>
        <div class="md-toolbar-container">
          <p class="md-title">Esterpad</p>
        </div>
      </md-toolbar>
      <md-list @click.native="closeSidenav">
        <md-list-item v-if="!state.isLoggedIn || !state.perms.notGuest">
          <router-link exact to="/.login">Login</router-link>
        </md-list-item>
        <md-list-item v-if="!state.isLoggedIn || !state.perms.notGuest">
          <router-link exact to="/.register">Register</router-link>
        </md-list-item>
        <md-list-item v-if="state.isLoggedIn">
          <router-link exact to="/.padlist">Pad List</router-link>
        </md-list-item>
        <md-list-item v-if="state.isLoggedIn" @click.native="signout">
          Sign Out
        </md-list-item>
      </md-list>
      <footer>
        &copy; 2016-2017 Anon2Anon
      </footer>
    </md-sidenav>

    <md-toolbar>
      <md-button class="md-icon-button nav-trigger" @click.native="toggleSidenav">
        <md-icon>menu</md-icon>
      </md-button>

      <p class="md-title-left md-title">{{ title }}</p>
      <esterpad-myuser v-if="state.isLoggedIn"
                       :user-name="state.userName" :user-color="state.userColor"
                       class="md-title">
      </esterpad-myuser>
    </md-toolbar>

    <router-view></router-view>

    <md-snackbar md-position="bottom right" ref="snackbar" :md-duration="2000">
      <span>{{ snckMsg }}</span>
      <md-button class="md-accent" md-theme="light-blue" @click.native="$refs.snackbar.close()">OK</md-button>
    </md-snackbar>
  </div>
</template>

<script>
import { state, bus } from '@/globs'

export default {
  name: 'app',
  data () {
    return {
      state: state,
      title: '',
      snckMsg: ''
    }
  },
  mounted () {
    bus.$on('auth-error', this.snackbarMsg)
  },
  methods: {
    toggleSidenav () {
      this.$refs.sidenav.toggle()
    },
    closeSidenav () {
      this.$refs.sidenav.close()
    },
    signout () {
      bus.$emit('send', 'Logout', {})
      state.isLoggedIn = false
      state.sessId = ''
      if (['/.login', '/.register'].indexOf(this.$route.path) < 0) {
        this.$router.push('/.login')
      }
    },
    snackbarMsg (msg) {
      this.snckMsg = msg
      this.$refs.snackbar.open()
    }
  },
  watch: {
    '$route' (to, from) {
      if (to.name !== 'Pad') {
        this.title = to.name
      } else {
        this.title = state.padId
      }
    }
  }
}
</script>

<style>
 html,
 body {
   height: 100%;
   overflow: hidden;
 }

 body {
   display: flex;
 }

 .md-title-left {
   flex: 1;
 }

 footer {
   height: 20px;
   line-height: 20px;
   font-size: 14px;
   color: #999;
   margin: 4px 7px;
 }

 .container {
   min-height: 100%;
   display: flex;
   flex-flow: column nowrap;
   flex: 1;
 }

 .md-sidenav .md-sidenav-content {
   display: flex;
   flex-flow: column;
 }

 .md-list {
   flex: 1 1 0;
 }
</style>
