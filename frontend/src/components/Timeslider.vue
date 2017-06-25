<template>
  <div>
    Timeslider for {{ state.padId }}
  </div>
</template>

<script>
import { state } from '@/globs'

export default {
  data () {
    return {
      state: state
    }
  },
  methods: {
  }
}
</script>
