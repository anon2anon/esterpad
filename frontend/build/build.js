require('./check-versions')()

process.env.NODE_ENV = 'production'

var ora = require('ora')
var rm = require('rimraf')
var path = require('path')
var chalk = require('chalk')
var webpack = require('webpack')
var config = require('../config')
var webpackConfig = require('./webpack.prod.conf')

// Build protobuf
var protobuf = require('protobufjs')
var fs = require('fs')

protobuf.load('./clientmessages.proto', function(e, p) {
  var compiledProto = JSON.stringify(p.toJSON())
  fs.writeFile('src/assets/proto.json', compiledProto, function(err) {
    if (err) return console.log(err)
    console.log('Compiled proto!')

// I'm really sorry for this. I'll try to fix this ASAP.
var spinner = ora('building for production...')
spinner.start()

rm(path.join(config.build.assetsRoot, config.build.assetsSubDirectory), err => {
  if (err) throw err
  webpack(webpackConfig, function (err, stats) {
    spinner.stop()
    if (err) throw err
    process.stdout.write(stats.toString({
      colors: true,
      modules: false,
      children: false,
      chunks: false,
      chunkModules: false
    }) + '\n\n')

    console.log(chalk.cyan('  Build complete.\n'))
    console.log(chalk.yellow(
      '  Tip: built files are meant to be served over an HTTP server.\n' +
      '  Opening index.html over file:// won\'t work.\n'
    ))
  })
})

  })
})


