function parseYAML(text) {
  const lines = text.split(/\r?\n/);
  const sections = [];
  let current = null;
  lines.forEach(line => {
    if (line.trim().startsWith('-')) {
      if (current) sections.push(current);
      current = {};
      const m = line.match(/-\s*name:\s*(.*)/);
      if (m) current.name = m[1].replace(/"/g, '').trim();
    } else if (/color:/.test(line)) {
      const m = line.match(/color:\s*(.*)/);
      if (m) current.color = m[1].replace(/"/g, '').trim();
    } else if (/started:/.test(line)) {
      const m = line.match(/started:\s*(.*)/);
      if (m) current.started = m[1].replace(/"/g, '').trim();
    }
  });
  if (current) sections.push(current);
  return sections;
}

function parseTime(str) {
  const [h, m] = str.split(':').map(Number);
  return h * 60 + m;
}

function computeAngles(sections) {
  const day = 24 * 60;
  sections.forEach((s, i) => {
    s.startMin = parseTime(s.started);
    let end = parseTime(sections[(i + 1) % sections.length].started);
    if (end <= s.startMin) end += day;
    s.endMin = end;
  });
}

function loadSections() {
  return fetch('sections.yaml')
    .then(res => res.text())
    .then(text => parseYAML(text));
}

function drawClock(sections) {
  const canvas = document.getElementById('clock');
  const ctx = canvas.getContext('2d');

  function resize() {
    canvas.width = canvas.clientWidth;
    canvas.height = canvas.clientHeight;
  }
  resize();
  window.addEventListener('resize', resize);

  function drawArrow(angle, radius) {
    const center = Math.min(canvas.width, canvas.height) / 2;
    ctx.save();
    ctx.translate(center, center);
    ctx.rotate(angle);

    const shaft = radius * 0.6;
    const head = radius * 0.25;
    const width = radius * 0.1;

    ctx.lineWidth = width;
    ctx.lineCap = 'round';
    ctx.strokeStyle = '#000';

    ctx.beginPath();
    ctx.moveTo(0, 0);
    ctx.lineTo(shaft, 0);
    ctx.stroke();

    ctx.beginPath();
    ctx.moveTo(shaft, -head * 0.6);
    ctx.lineTo(shaft + head, 0);
    ctx.lineTo(shaft, head * 0.6);
    ctx.closePath();
    ctx.fillStyle = '#000';
    ctx.fill();

    ctx.restore();
  }

  function draw() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    const radius = Math.min(canvas.width, canvas.height) / 2 * 0.9;
    const center = Math.min(canvas.width, canvas.height) / 2;
    const now = new Date();
    const min24 = now.getHours() * 60 + now.getMinutes() + now.getSeconds() / 60;
    const day = 24 * 60;
    const min = (now.getHours() % 24) * 60 + now.getMinutes() + now.getSeconds() / 60;
    const day24 = 24 * 60;

    let currentIndex = sections.length - 1;
    sections.forEach((s, i) => {
      const startAngle = (s.startMin / day) * 2 * Math.PI - Math.PI / 2;
      const endAngle = (s.endMin / day) * 2 * Math.PI - Math.PI / 2;
      const isCurrent = min24 >= s.startMin && min24 < s.endMin;
      if (isCurrent) currentIndex = i;
      ctx.beginPath();
      const r = isCurrent ? radius * 1.05 : radius;
      ctx.moveTo(center, center);
      ctx.arc(center, center, r, startAngle, endAngle, false);
      ctx.closePath();
      ctx.fillStyle = s.color;
      ctx.globalAlpha = isCurrent ? 1 : 0.5;
      ctx.fill();
      if (isCurrent) {
        ctx.lineWidth = 4;
        ctx.strokeStyle = '#fff';
        ctx.stroke();
      }
    });

    ctx.globalAlpha = 1;
    const angle = (min / day24) * 2 * Math.PI - Math.PI / 2;
    drawArrow(angle, radius);

    requestAnimationFrame(draw);
  }
  draw();
}

loadSections().then(sections => {
  computeAngles(sections);
  drawClock(sections);
});
