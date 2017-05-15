export let num2color = function (num) {
  return '#' + ('000000' + num.toString(16)).slice(-6)
}

export let color2num = function (_color) {
  let color = _color.trim()
  if (color.startsWith('#')) color = color.substr(1)
  return parseInt(color, 16)
}
