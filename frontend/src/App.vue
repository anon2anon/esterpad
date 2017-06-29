<template>
  <div id="app" class="container">
    <md-sidenav class="md-left main-sidenav" ref="sidenav">
      <md-toolbar>
        <div class="md-toolbar-container">
          <p class="md-title">Esterpad</p>
        </div>
      </md-toolbar>
      <md-list @click.native="closeSidenav" class="menulist">
        <md-list-item v-if="!state.isLoggedIn || !state.perms.notGuest">
          <router-link to="/.login">Login</router-link>
        </md-list-item>
        <md-list-item v-if="!state.isLoggedIn || !state.perms.notGuest">
          <router-link to="/.register">Register</router-link>
        </md-list-item>
        <md-list-item v-if="state.isLoggedIn">
          <router-link to="/.padlist">Pad List</router-link>
        </md-list-item>
        <md-list-item v-if="state.isLoggedIn">
          <router-link to="/.options">Options</router-link>
        </md-list-item>
        <md-list-item v-if="state.isLoggedIn && state.perms.mod">
          <router-link to="/.users">Users</router-link>
        </md-list-item>
        <md-list-item v-if="state.isLoggedIn && state.perms.admin">
          <router-link to="/.admin">Admin</router-link>
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

      <p style="margin-left: 10px" class="md-title-left md-title">{{ title }}</p>
      <esterpad-myuser v-if="state.isLoggedIn"
                       :user-name="state.userName" :user-color="state.userColor"
                       class="md-title">
      </esterpad-myuser>
    </md-toolbar>

    <router-view class="rview"></router-view>

    <md-snackbar md-position="bottom right" ref="snackbar" :md-duration="2000">
      <span>{{ snckMsg }}</span>
      <md-button class="md-accent" md-theme="light-blue" @click.native="$refs.snackbar.close()">OK</md-button>
    </md-snackbar>

    <transition name="fade">
      <div v-if="state.loading" class="loading">
        <md-spinner :md-size="100" md-indeterminate class="md-warn"></md-spinner>
      </div>
    </transition>
  </div>
</template>

<script>
import MyUser from '@/components/MyUser'
import { state, bus } from '@/globs'

export default {
  components: {
    'esterpad-myuser': MyUser
  },
  data () {
    return {
      state: state,
      title: '',
      snckMsg: ''
    }
  },
  mounted () {
    bus.$on('snack-msg', this.snackbarMsg)
    this.updateTitle(this.$route)
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
    },
    updateTitle (to) {
      if (to.name) {
        this.title = to.name
      } else {
        this.title = state.padId
      }
    }
  },
  watch: {
    '$route' (to, from) {
      this.updateTitle(to)
    }
  }
}
</script>

<style>
 html, body, .rview {
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

 .main-sidenav > .md-sidenav-content {
   width: 200px !important;
   display: flex;
   flex-flow: column;
   overflow: hidden;
 }

 @media (min-width: 1281px) {
   .container {
     padding-left: 200px;
   }

   .main-sidenav > .md-sidenav-content {
     top: 0 !important;
     pointer-events: auto !important;
     transform: translate3d(0, 0, 0) !important;
     box-shadow: 0 1px 5px rgba(0,0,0,.2), 0 2px 2px rgba(0,0,0,.14), 0 3px 1px -2px rgba(0,0,0,.12) !important;
   }

   .main-sidenav > .md-backdrop {
     opacity: 0 !important;
     pointer-events: none !important;
   }

   .nav-trigger {
     display: none !important;
   }
 }

 .menulist {
   overflow-y: auto;
 }

 /* Absolute Center Spinner */
.loading {
  position: fixed;
  z-index: 999;
  height: 100px;
  width: 100px;
  overflow: show;
  margin: auto;
  top: 0;
  left: 0;
  bottom: 0;
  right: 0;
}

/* Transparent Overlay */
.loading:before {
  content: '';
  display: block;
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0,0,0,0.3);
}

.fade-enter-active, .fade-leave-active {
  transition: opacity .5s
}
.fade-enter, .fade-leave-to {
  opacity: 0
}
</style>
