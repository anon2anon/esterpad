<template>
  <md-layout :md-gutter="8">
    <md-layout md-flex="75">
      <esterpad-editor></esterpad-editor>
    </md-layout>
    <md-layout md-column :md-gutter="8">
      <md-layout md-flex="40">
        <esterpad-userlist></esterpad-userlist>
      </md-layout>
      <md-layout>
        Chat
      </md-layout>
    </md-layout>
  </md-layout>
</template>

<script>
export default {
  name: 'esterpad-pad'
}
</script>
