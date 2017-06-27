<template>
  <md-card>
    <md-card-content>
      <md-input-container>
        <label>Email</label>
        <md-input @keyup.native.enter="sendLogin" v-model="email"></md-input>
      </md-input-container>
      <md-input-container md-has-password>
        <label>Password</label>
        <md-input @keyup.native.enter="sendLogin" v-model="passwd" type="password"></md-input>
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
import state from '@/state'

export default {
  name: 'esterpad-login',
  data () {
    return {
      email: '',
      passwd: '',
      state: state
    }
  },
  methods: {
    sendLogin () {
      state.sendMessage({
        Login: {
          email: this.email,
          password: this.pass
        },
        CMessage: 'Login'
      })
    },
    guestLogin () {
      state.sendMessage({
        Session: {
          sessId: '' // TODO: read sessId
        },
        CMessage: 'Session'
      })
    }
  }
}
</script>
