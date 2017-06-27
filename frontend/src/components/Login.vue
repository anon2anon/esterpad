<template>
  <md-card class="card">
    <md-card-content>
      <md-input-container :class="{ 'md-input-invalid': emailInvalid }">
        <label>Email</label>
        <md-input @keyup.native.enter="sendLogin" v-model="email"></md-input>
        <span class="md-error">Email can not be empty</span>
      </md-input-container>
      <md-input-container md-has-password  :class="{ 'md-input-invalid': passwdInvalid }">
        <label>Password</label>
        <md-input @keyup.native.enter="sendLogin" v-model="passwd" type="password"></md-input>
        <span class="md-error">Password can not be empty</span>
      </md-input-container>
    </md-card-content>
    <md-card-actions>
      <md-button @click.native="sendLogin" class="md-raised">
        Login
      </md-button>
      <md-button @click.native="guestLogin" class="md-raised"
                 v-if="!state.isLoggedIn">
        Continue as guest
      </md-button>
    </md-card-actions>
  </md-card>
</template>

<script>
import { state, bus } from '@/globs'

export default {
  data () {
    return {
      email: '',
      passwd: '',
      emailInvalid: false,
      passwdInvalid: false,
      state: state
    }
  },
  beforeDestroy () {
    log.debug('login destroy')
    bus.$off('auth-error', this.loginError)
  },
  methods: {
    sendLogin () {
      this.emailInvalid = this.email === ''
      this.passwdInvalid = this.passwd === ''
      if (this.emailInvalid || this.passwdInvalid) return
      bus.$on('auth-error', this.loginError)
      bus.$emit('send', 'Login', {
        email: this.email,
        password: this.passwd
      })
    },
    guestLogin () {
      // TODO: read sessId
      bus.$emit('send', 'Session', {sessId: ''})
    },
    loginError () {
      bus.$emit('snack-msg', 'Invalid username or password')
    }
  }
}
</script>

<style scoped>
 .card {
   margin: 15px;
 }
</style>
