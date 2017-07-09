var commit = require('child_process')
    .execSync('git rev-parse --short HEAD')
    .toString().trim()
var version = require("../package.json").version

module.exports = {
  NODE_ENV: '"production"',
  VERSION: JSON.stringify(version),
  COMMIT: JSON.stringify(commit)
}
