<template>
  <div>
    <div v-for="user in userList" key="user">
      {{ user.nickname }} {{ user.color }}
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
      console.log(info)
      if (info.online) {
        this.userList.push(info)
      }
    },
    userLeave (info) {
      console.log(info)
    }
  }
}
</script>
