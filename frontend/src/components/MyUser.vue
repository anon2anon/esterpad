<template>
  <div>
    <span contenteditable @keydown.prevent.enter="nameChanged" ref="userName">
      {{ userName }}
    </span>
    <div class="avatar" :style="{ background: userColor }"></div>
  </div>
</template>

<script>
import { state, bus } from '@/globs'
import { color2num } from '@/helpers'

export default {
  props: {
    userName: String,
    userColor: String
  },
  methods: {
    nameChanged () {
      this.$refs.userName.blur()
      state.userName = this.$refs.userName.textContent.trim()
      bus.$emit('send', 'EditUser', {
        changemask: 1,
        nickname: state.userName
      })
    },
    colorChanged () {
      bus.$emit('send', 'EditUser', {
        changemask: 2,
        color: color2num(this.userColor)
      })
    }
  }
}
</script>

<style scoped>
 .avatar {
   display: inline-block;
   box-sizing: border-box;
   width: 1em;
   height: 1em;
   border: 1px solid #000;
   box-shadow: 0 0 1px 0px white inset, 0 0 1px 0px white;
   -moz-border-radius: 50%;
   -webkit-border-radius: 50%;
   border-radius: 50%;
   margin-right: 4px;
 }
</style>
