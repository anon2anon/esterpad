<template>
  <md-card>
    <md-card-content>
      <md-input-container :class="{ 'md-input-invalid': nicknameInvalid }">
        <label>Nickname</label>
        <md-input @keyup.native.enter="register" v-model="nickname"></md-input>
        <span class="md-error">Nickname can not be empty</span>
      </md-input-container>
      <md-input-container :class="{ 'md-input-invalid': emailInvalid }">
        <label>Email</label>
        <md-input @keyup.native.enter="register" v-model="email"></md-input>
        <span class="md-error">Email can not be empty</span>
      </md-input-container>
      <md-input-container md-has-password :class="{ 'md-input-invalid': passwdInvalid }">
        <label>Password</label>
        <md-input @keyup.native.enter="register" v-model="passwd" type="password"></md-input>
        <span class="md-error">Password can not be empty</span>
      </md-input-container>
      <md-input-container md-has-password :class="{ 'md-input-invalid': passwd2Invalid }">
        <label>Repeat password</label>
        <md-input @keyup.native.enter="register" v-model="passwd2" type="password"></md-input>
        <span class="md-error" ref="repeatErrorLabel">Password can not be empty</span>
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
import { bus } from '@/globs'

export default {
  name: 'esterpad-login',
  data () {
    return {
      nickname: '',
      email: '',
      passwd: '',
      passwd2: '',
      nicknameInvalid: false,
      emailInvalid: false,
      passwdInvalid: false,
      passwd2Invalid: false
    }
  },
  methods: {
    register () {
      this.nicknameInvalid = this.nickname === ''
      this.emailInvalid = this.email === ''
      this.passwdInvalid = this.passwd === ''
      this.passwd2Invalid = (this.passwd2 === '' ||
                             this.passwd !== this.passwd2)
      if (this.passwd2 === '') {
        this.$refs.repeatErrorLabel.innerHTML = 'Password can not be empty'
      } else if (this.passwd !== this.passwd2) {
        this.$refs.repeatErrorLabel.innerHTML = 'Passwords do not match'
      }
      if (this.nicknameInvalid || this.emailInvalid ||
          this.passwdInvalid || this.passwd2Invalid) return

      bus.$on('auth-error', function () { bus.$emit('snack-msg', 'User already exists') })
      bus.$emit('send', 'Register', {
        email: this.email,
        nickname: this.nickname,
        password: this.passwd
      })
    }
  }
}
</script>
