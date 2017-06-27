<template>
  <md-layout md-row>
    <md-layout md-flex="75" class="bordered-right">
      <router-view></router-view>
    </md-layout>
    <md-layout md-column>
      <md-layout md-flex="40" class="bordered-bottom">
        <esterpad-userlist></esterpad-userlist>
      </md-layout>
      <md-layout>
        <esterpad-chat></esterpad-chat>
      </md-layout>
    </md-layout>
  </md-layout>
</template>

<script>
import Chat from '@/components/Chat'
import UserList from '@/components/UserList'

export default {
  components: {
    'esterpad-userlist': UserList,
    'esterpad-chat': Chat
  }
}
</script>

<style scoped>
 .bordered-right {
   border-right-style: solid;
   border-width: 1px;
 }
 .bordered-bottom {
   border-bottom-style: solid;
   border-width: 1px;
 }
</style>
