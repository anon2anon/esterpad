<template>
  <div>
    <md-dialog ref="dialog">
      <md-dialog-title>Enter new pad name</md-dialog-title>

      <md-dialog-content>
        <md-input-container md-inline :class="{ 'md-input-invalid': haveError }">
          <label>Name</label>
          <md-input v-model="newPadName" ref="newPadInput"></md-input>
          <span class="md-error">Dots and slashes are not allowed!</span>
        </md-input-container>
      </md-dialog-content>

      <md-dialog-actions>
        <md-button class="md-primary" @click.native="newPadCancel">Cancel</md-button>
        <md-button class="md-primary" @click.native="newPadOk">Ok</md-button>
      </md-dialog-actions>
    </md-dialog>
    <md-card class="card">
      <md-card-content>
        <md-list>
          <md-list-item v-for="pad in state.padList" key="pad">
            <router-link exact :to="pad">{{ pad }}</router-link>
          </md-list-item>
          <md-list-item @click.native="createPad" v-if="state.perms.mod">
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
  data () {
    return {
      state: state,
      newPadName: '',
      haveError: false
    }
  },
  watch: {
    newPadName (padName) {
      this.haveError = padName.indexOf('.') !== -1 || padName.indexOf('/') !== -1
    }
  },
  methods: {
    createPad () {
      this.$refs.dialog.open()
      // TODO: for some reason this doesn't work
      this.$nextTick(() => {
        this.$refs.newPadInput.$el.focus()
      })
    },
    newPadCancel () {
      this.newPadName = ''
      this.$refs.dialog.close()
    },
    newPadOk () {
      // TODO: check for pad existence here
      if (this.haveError) return
      if (this.newPadName) {
        this.$router.push('/' + this.newPadName)
      }
      this.$refs.dialog.close()
    }
  }
}
</script>

<style scoped>
 .card {
   margin: 15px;
 }
</style>
