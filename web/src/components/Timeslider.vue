<template>
  <div class="editor flex">
    <vue-slider ref="slider" v-model="revision"
                :min="0" :max="maxRevision"
                tooltip="always" tooltip-dir="bottom"
                style="z-index: 90" @callback="revChange">
    </vue-slider>
    <div class="toolbar">
      <div class="toolbar-buttons">
        <div class="button-group">
          <router-link to=".." class="link-black" append>
            <div class="button"><i class="material-icons">mode_edit</i></div>
          </router-link>
          <div class="button" @click="restoreRevision"><i class="material-icons">restore_page</i></div>
        </div>
      </div>
    </div>
    <div ref="cm"></div>
  </div>
</template>

<script>
import { state, bus } from '@/globs'
import CodeMirror from 'codemirror'
import 'codemirror/lib/codemirror.css'
import CodemirrorAdapter from '@/ot/CodemirrorAdapter'
import TextOperation from '@/ot/TextOperation'
import CSSManager from '@/lib/cssmanager'
import { textColor } from '@/helpers'
import vueSlider from 'vue-slider-component'

export default {
  data () {
    return {
      cma: null,
      revision: 0,
      maxRevision: 0,
      cssManager: null,
      state: state,
      waitingForRestore: false
    }
  },
  components: {
    vueSlider
  },
  mounted () {
    log.debug('timeslider mounted')

    // isn't it a RC?
    this.reinitCM(state.padId)
    this.cssManager = new CSSManager()

    bus.$on('pad-id-changed', this.reinitCM)
    bus.$on('document', this.recvDocument)
    bus.$on('new-delta', this.newDelta)
    bus.$on('user-leave', this.userLeave)
    bus.$on('color-update', this.updateColor)

    if (state.userId !== 0) {
      this.updateColor(state.userId, state.userColor)
    }
  },
  beforeDestroy () {
    log.debug('timeslider destroy')

    bus.$off('pad-id-changed', this.reinitCM)
    bus.$off('document', this.recvDocument)
    bus.$off('new-delta', this.newDelta)
    bus.$off('user-leave', this.userLeave)
    bus.$off('color-update', this.updateColor)
  },
  methods: {
    reinitCM (padId) {
      log.debug('reinitCM', padId)
      bus.$emit('send', 'EnterPad', {name: padId})

      if (this.cma) {
        log.debug('Clearing editor')
        this.cma.clear()
      } else {
        log.debug('Creating editor')

        let cm = CodeMirror(this.$refs.cm, {
          value: '', // (TODO: make cool spinner here)
          tabSize: 4,
          mode: 'text/plain',
          lineNumbers: true,
          lineWrapping: true,
          readOnly: 'nocursor'
        })

        this.cma = new CodemirrorAdapter(cm)
      }
    },
    revChange (val) {
      log.debug('Requesting revision', val)
      bus.$emit('send', 'RevisionRequest', {revision: val})
    },
    recvDocument (doc) {
      log.debug('recv doc', doc)

      if (this.revision === 0) this.revision = doc.revision
      this.maxRevision = Math.max(this.maxRevision, doc.revision)

      if (this.cma) {
        log.debug('Clearing editor')
        this.cma.clear()
      }

      let to = (new TextOperation()).fromProtobuf(doc)
      log.debug('Converted doc', to)

      this.cma.applyOperation(to)
    },
    newDelta (delta) {
      this.maxRevision = Math.max(this.maxRevision, delta.id)
      if (this.waitingForRestore) {
        this.revision = this.maxRevision
        this.waitingForRestore = false
      }
    },
    updateColor (userId, newColor) {
      this.cssManager.selectorStyle('.author-' + userId).background = newColor
      let fgColor = textColor(newColor)
      this.cssManager.selectorStyle('.author-' + userId).color = fgColor
    },
    userLeave (info) {
      log.warn('TODO: handle user leave in editor', info)
    },
    restoreRevision () {
      if (this.waitingForRestore) return
      log.debug('Restoring revision', this.revision)
      bus.$emit('send', 'RestoreRevision', {rev: this.revision})
      this.waitingForRestore = true
    }
  }
}
</script>

<style scoped>

  .editor{
    display: grid;
    grid-template-rows: 50px 45px 1fr;
    height: 100%;
  }
</style>
