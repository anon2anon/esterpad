<template>
  <div id="app" class="container">
    <md-sidenav class="md-left md-fixed" ref="sidenav">
      <md-toolbar>
        <div class="md-toolbar-container">
          <p class="md-title">Esterpad</p>
        </div>
      </md-toolbar>
      <md-list>
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
      <span>{{ state.snackbarMsg }}</span>
      <md-button class="md-accent" md-theme="light-blue" @click.native="$refs.snackbar.close()">OK</md-button>
    </md-snackbar>
  </div>
</template>

<script>
import state from '@/state'

export default {
  name: 'app',
  data () {
    return {
      state: state,
      title: ''
    }
  },
  methods: {
    toggleSidenav () {
      this.$refs.sidenav.toggle()
    },
    closeSidenav () {
      this.$refs.sidenav.close()
    },
    signout () {
      state.sendMessage({
        Logout: {},
        CMessage: 'Logout'
      })
      state.isLoggedIn = false
      if (['/.login', '/.register'].indexOf(this.$route.path) < 0) {
        this.$router.push('/.login')
      }
    }
  },
  watch: {
    '$route' (to, from) {
      if (to.name !== 'Pad') {
        this.title = to.name
      } else {
        this.title = state.padId
      }
    },
    'state.snackbarMsg' (to, from) {
      if (to !== '') {
        this.$refs.snackbar.open()
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
   transition: all .4s !default cubic-bezier(.25, .8, .25, 1) !default !default;
 }

 .md-sidenav .md-sidenav-content {
   width: 280px !important;
   display: flex;
   flex-flow: column;
   overflow: hidden !important;
 }

 .md-list {
   flex: 1 1 auto;
 }

 @media (min-width: 1281px) {
   .md-title-left {
     margin-left: 8px !important;
   }

   .nav-trigger {
     display: none !important;
   }

   .container {
     padding-left: 280px;
   }

   .md-backdrop {
     opacity: 0;
     pointer-events: none;
   }

   .md-sidenav .md-sidenav-content {
     top: 0 !important;
     pointer-events: auto !important;
     transform: translate3d(0, 0, 0);
     box-shadow: 0 1px 5px rgba(0,0,0,.2), 0 2px 2px rgba(0,0,0,.14), 0 3px 1px -2px rgba(0,0,0,.12);
     left: 280px !important;
   }
 }
</style>
