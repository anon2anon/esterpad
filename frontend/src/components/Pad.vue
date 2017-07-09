<template>
  <div>
    <div id="left-panel" class="split split-horizontal">
      <router-view></router-view>
    </div>
    <div id="right-panel" class="split split-horizontal">
      <div id="userlist">
        <esterpad-userlist></esterpad-userlist>
      </div>
      <div id="chat">
        <esterpad-chat></esterpad-chat>
      </div>
    </div>
  </div>
</template>

<script>
import Split from 'split.js'

import Chat from '@/components/Chat'
import UserList from '@/components/UserList'

export default {
  components: {
    'esterpad-userlist': UserList,
    'esterpad-chat': Chat
  },
  mounted () {
    Split(['#left-panel', '#right-panel'], {
      sizes: [75, 25]
    })
    Split(['#userlist', '#chat'], {
      sizes: [40, 60],
      direction: 'vertical'
    })
  }
}
</script>

<style>
 .split {
   -webkit-box-sizing: border-box;
   -moz-box-sizing: border-box;
   box-sizing: border-box;

   overflow-y: auto;
   overflow-x: hidden;
 }

 .gutter {
   background-color: #eee;

   background-repeat: no-repeat;
   background-position: 50%;
 }

 .gutter.gutter-horizontal {
   background-image: url('../assets/vertical.png');
   cursor: ew-resize;
 }

 .gutter.gutter-vertical {
   background-image: url('../assets/horizontal.png');
   cursor: ns-resize;
 }

  .split.split-horizontal, .gutter.gutter-horizontal {
    height: 100%;
    float: left;
  }
</style>
