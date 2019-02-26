/* eslint-disable no-console */
var protobuf = require('protobufjs')
var fs = require('fs')

process.env.VUE_APP_COMMIT = require('child_process').execSync('git rev-parse --short HEAD').toString().trim()

if (process.env.NODE_ENV == 'production') {
  process.env.VUE_APP_VERSION = require('./package.json').version
} else {
  process.env.VUE_APP_VERSION = 'dev'
}

module.exports = {
  configureWebpack: config => {
    protobuf.load('../internal/proto/esterpad.proto', function(e, p) {
      if (e) return console.log(e)
      var compiledProto = JSON.stringify(p.toJSON())
      fs.writeFile('src/assets/proto.json', compiledProto, function(err) {
        if (err) return console.log(err)
        console.log('Compiled proto!')
      })
      return config
    })
  },
  runtimeCompiler: true
}

