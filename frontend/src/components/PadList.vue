<template>
  <div>
    <md-dialog-prompt
      md-title="Enter new pad name"
      md-ok-text="Create!"
      md-cancel-text="Cancel"
      v-model="newPadName"
      @close="onClose"
      ref="dialog">
    </md-dialog-prompt>
    <md-card>
      <md-card-content>
        <md-list>
          <md-list-item v-for="pad in state.padList" key="pad">
            <router-link exact :to="pad">{{ pad }}</router-link>
          </md-list-item>
          <md-list-item @click.native="createPad">
            Create new?
          </md-list-item>
        </md-list>
      </md-card-content>
    </md-card>
  </div>
</template>

<script>
import { state } from '@/globs'

export default {
  name: 'esterpad-padlist',
  data () {
    return {
      state: state,
      newPadName: ''
    }
  },
  methods: {
    createPad () {
      this.$refs.dialog.open()
    },
    onClose () {
      if (this.newPadName) {
        this.$router.push('/' + this.newPadName)
      }
    }
  }
}
</script>
