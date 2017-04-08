<template>
  <div>User list here</div>
</template>

<script>
import { bus } from '@/globs'

export default {
  name: 'esterpad-userlist',
  mounted () {
    console.log('mounted')
    bus.$on('user-info', this.userInfo)
    bus.$on('user-leave', this.userLeave)
  },
  methods: {
    userInfo (info) {
      console.log(info)
    },
    userLeave (info) {
      console.log(info)
    }
  }
}
</script>
