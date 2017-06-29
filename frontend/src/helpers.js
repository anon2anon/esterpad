export let num2color = function (num) {
  return '#' + ('000000' + num.toString(16)).slice(-6)
}

export let color2num = function (_color) {
  let color = _color.trim()
  if (color.startsWith('#')) color = color.substr(1)
  return parseInt(color, 16)
}

export let textColor = function (_color) {
  let colorInt = _color
  if (typeof _color === 'string') {
    colorInt = color2num(_color)
  }
  let r = ~~(colorInt / (1 << 16))
  let g = (~~(colorInt / (1 << 8))) % (1 << 8)
  let b = colorInt % (1 << 8)
  let l = (Math.max(r, g, b) + Math.min(r, g, b)) / 2 / 255
  return l < 0.5 ? '#fff' : '#000'
}

export let permsMask = {
  notGuest: 1,
  chat: (1 << 1),
  write: (1 << 2),
  edit: (1 << 3),
  whitewash: (1 << 4),
  mod: (1 << 5),
  admin: (1 << 6)
}
