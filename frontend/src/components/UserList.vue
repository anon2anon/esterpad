<template>
  <div>
    <div v-for="user in userList" key="user">
      <md-layout class="user-item">
        <div class="avatar" :style="{ background: num2color(user.color) }"></div>
        <div>{{ user.nickname }}</div>
      </md-layout>
    </div>
  </div>
</template>

<script>
import { bus } from '@/globs'

export default {
  name: 'esterpad-userlist',
  data () {
    return {
      userList: []
    }
  },
  mounted () {
    console.log('mounted')
    bus.$on('user-info', this.userInfo)
    bus.$on('user-leave', this.userLeave)
  },
  methods: {
    userInfo (info) {
      console.log('user connected', info)
      if (info.online) {
        var tmp = this.userList.findIndex(
          i => i.userId === info.userId
        )
        if (tmp === -1) { // create new user
          this.userList.push(info)
        } else { // update existing
          this.userList.splice(tmp, 1, info)
        }
      }
    },
    userLeave (info) {
      console.log('user left', info)
      this.userList.splice(this.userList.findIndex(
        i => i.userId === info.userId
      ), 1)
    },
    num2color (num) {
      return '#' + ('000000' + num.toString(16)).slice(-6)
    }
  }
}
</script>

<style scoped>
 .container {
   display: block;
   height: 100%;
 }
 .avatar {
   display: inline-block;
   box-sizing: border-box;
   width: 20px;
   height: 20px;
   border-radius: 50%;
   border: 1px solid #000;
   box-shadow: 0 0 1px 0px white inset, 0 0 1px 0px white;
   -moz-border-radius: 50%;
   -webkit-border-radius: 50%;
   margin-right: 4px;
 }
 .user-item {
   padding: 3px;
   display: flex;
   align-items: center;
 }
</style>
