<template>
  <div class="user-outer-container">
    <div class="user-inner-container">
      <div v-for="user in userList" key="user" >
        <md-layout class="user-item" :title="user.ip ? user.ip + '\n' + user.userAgent : ''">
          <div class="avatar" :style="{ background: num2color(user.color) }"></div>
          <div :style="{ color: userColor(user.perms) }">{{ user.nickname }}</div>
        </md-layout>
      </div>
    </div>
  </div>
</template>

<script>
import { bus } from '@/globs'
import { num2color, permsMask } from '@/helpers'

export default {
  data () {
    return {
      userList: []
    }
  },
  mounted () {
    log.debug('userlist mounted')
    bus.$on('user-info', this.userInfo)
    bus.$on('user-leave', this.userLeave)
  },
  beforeDestroy () {
    log.debug('userlist destroy')
    bus.$off('user-info', this.userInfo)
    bus.$off('user-leave', this.userLeave)
  },
  methods: {
    userInfo (info) {
      log.debug('user connected', info)

      let tmp = this.userList.findIndex(
        i => i.userId === info.userId
      )

      if (tmp !== -1) { // update existing
        Object.assign(this.userList[tmp], info)
      } else if (info.online) { // create new
        this.userList.push(info)
      }
    },
    userLeave (info) {
      log.debug('user left', info)
      this.userList.splice(this.userList.findIndex(
        i => i.userId === info.userId
      ), 1)
    },
    num2color: num2color,
    userColor (perms) {
      log.debug(!(perms & permsMask.notGuest))
      if (!(perms & permsMask.notGuest)) {
        return '#999'
      } else if (perms & permsMask.admin) {
        return '#831'
      } else if (perms & permsMask.mod) {
        return '#138'
      } else if (perms & permsMask.whitewash) {
        return '#163'
      }
      return '#000'
    }
  }
}
</script>

<style scoped>
 .user-outer-container {
   min-height: 100%;
   display: flex;
   flex-flow: column nowrap;
   flex: 1;
   white-space: pre-wrap;
 }

 .user-inner-container {
   flex: 1 1 0;
   overflow-y: scroll;
   word-break: break-all;
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
