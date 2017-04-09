<template>
  <md-layout md-row>
    <md-layout md-flex="75" class="bordered-right">
      <esterpad-editor></esterpad-editor>
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
export default {
  name: 'esterpad-pad'
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
