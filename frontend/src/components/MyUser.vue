<template>
  <div>
    {{ userName }}
    <div class="avatar" :style="{ background: userColor }"></div>
  </div>
</template>

<script>
export default {
  name: 'esterpad-myuser',
  props: {
    userName: String,
    userColor: String
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
