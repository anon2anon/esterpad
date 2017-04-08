<template>
  <md-card>
    <md-card-content>
      <md-input-container>
        <label>Nickname</label>
        <md-input @keyup.native.enter="register" v-model="nickname"></md-input>
      </md-input-container>
      <md-input-container>
        <label>Email</label>
        <md-input @keyup.native.enter="register" v-model="email"></md-input>
      </md-input-container>
      <md-input-container md-has-password>
        <label>Password</label>
        <md-input @keyup.native.enter="register" v-model="passwd" type="password"></md-input>
      </md-input-container>
      <md-input-container md-has-password>
        <label>Repeat password</label>
        <md-input @keyup.native.enter="register" v-model="passwd2" type="password"></md-input>
      </md-input-container>
    </md-card-content>
    <md-card-actions>
      <md-button @click.native="register" class="md-raised">
        Register
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
      nickname: '',
      email: '',
      passwd: '',
      passwd2: ''
    }
  },
  methods: {
    register () {
      // validate passwds or smth
      state.sendMessage('Register', {
        email: this.email,
        nickname: this.nickname,
        password: this.pass
      })
    }
  }
}
</script>
