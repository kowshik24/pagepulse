const connEl = document.getElementById("connState");
const hostMetaEl = document.getElementById("hostMeta");
const buildMetaEl = document.getElementById("buildMeta");
const cpuEl = document.getElementById("cpu");
const memEl = document.getElementById("memory");
const diskEl = document.getElementById("disk");
const netEl = document.getElementById("network");
const cpuMetaEl = document.getElementById("cpuMeta");
const memMetaEl = document.getElementById("memoryMeta");
const diskMetaEl = document.getElementById("diskMeta");
const netMetaEl = document.getElementById("netMeta");
const resEl = document.getElementById("resources");
const themeToggleEl = document.getElementById("themeToggle");
const cpuLegendEl = document.getElementById("cpuLegend");
const memLegendEl = document.getElementById("memLegend");
const diskLegendEl = document.getElementById("diskLegend");
const netLegendEl = document.getElementById("netLegend");

const charts = {
  cpu: document.getElementById("cpuChart").getContext("2d"),
  mem: document.getElementById("memChart").getContext("2d"),
  disk: document.getElementById("diskChart").getContext("2d"),
  net: document.getElementById("netChart").getContext("2d"),
};
const miniCharts = {
  cpu: document.getElementById("cpuMini").getContext("2d"),
  mem: document.getElementById("memMini").getContext("2d"),
  disk: document.getElementById("diskMini").getContext("2d"),
  net: document.getElementById("netMini").getContext("2d"),
};

const themeKey = "pagepulse-theme";
const savedTheme = localStorage.getItem(themeKey);
if (savedTheme) document.documentElement.setAttribute("data-theme", savedTheme);

themeToggleEl.addEventListener("click", () => {
  const cur = document.documentElement.getAttribute("data-theme") === "dark" ? "light" : "dark";
  document.documentElement.setAttribute("data-theme", cur);
  localStorage.setItem(themeKey, cur);
});

function bytes(n) {
  if (n < 1024) return `${n.toFixed(0)} B`;
  const units = ["KB", "MB", "GB", "TB"];
  let i = -1;
  do { n /= 1024; i += 1; } while (n >= 1024 && i < units.length - 1);
  return `${n.toFixed(1)} ${units[i]}`;
}

function mbps(bps) {
  return (bps / (1024 * 1024)).toFixed(2);
}

function pct(v) {
  return `${Number(v || 0).toFixed(1)}%`;
}

function drawGrid(ctx, width, height, pad, lineColor) {
  ctx.strokeStyle = lineColor;
  ctx.lineWidth = 1;
  for (let i = 0; i <= 4; i++) {
    const y = pad + (i / 4) * (height - pad * 2);
    ctx.beginPath();
    ctx.moveTo(pad, y);
    ctx.lineTo(width - pad, y);
    ctx.stroke();
  }
}

function drawLine(ctx, values, color, minY = 0, maxY = 100, fill = false) {
  const width = ctx.canvas.width;
  const height = ctx.canvas.height;
  const pad = 20;

  ctx.clearRect(0, 0, width, height);
  const css = getComputedStyle(document.documentElement);
  drawGrid(ctx, width, height, pad, css.getPropertyValue("--line"));

  if (!values.length) return;
  const range = Math.max(maxY - minY, 1);
  const points = values.map((v, i) => {
    const x = pad + (i / Math.max(values.length - 1, 1)) * (width - pad * 2);
    const normalized = (Math.min(Math.max(v, minY), maxY) - minY) / range;
    const y = pad + (1 - normalized) * (height - pad * 2);
    return { x, y };
  });

  if (fill) {
    const grad = ctx.createLinearGradient(0, 0, 0, height);
    grad.addColorStop(0, `${color}66`);
    grad.addColorStop(1, `${color}00`);
    ctx.beginPath();
    ctx.moveTo(points[0].x, height - pad);
    points.forEach((p) => ctx.lineTo(p.x, p.y));
    ctx.lineTo(points[points.length - 1].x, height - pad);
    ctx.closePath();
    ctx.fillStyle = grad;
    ctx.fill();
  }

  ctx.beginPath();
  points.forEach((p, idx) => (idx === 0 ? ctx.moveTo(p.x, p.y) : ctx.lineTo(p.x, p.y)));
  ctx.strokeStyle = color;
  ctx.lineWidth = 2.2;
  ctx.stroke();
}

function drawMini(ctx, values, color, minY, maxY) {
  const width = ctx.canvas.width;
  const height = ctx.canvas.height;
  ctx.clearRect(0, 0, width, height);
  if (!values.length) return;
  const pad = 4;
  const range = Math.max(maxY - minY, 1);
  const points = values.map((v, i) => {
    const x = pad + (i / Math.max(values.length - 1, 1)) * (width - pad * 2);
    const normalized = (Math.min(Math.max(v, minY), maxY) - minY) / range;
    const y = pad + (1 - normalized) * (height - pad * 2);
    return { x, y };
  });
  ctx.beginPath();
  points.forEach((p, idx) => (idx === 0 ? ctx.moveTo(p.x, p.y) : ctx.lineTo(p.x, p.y)));
  ctx.strokeStyle = color;
  ctx.lineWidth = 2;
  ctx.stroke();
}

function drawDualLine(ctx, valuesA, colorA, valuesB, colorB) {
  const width = ctx.canvas.width;
  const height = ctx.canvas.height;
  const pad = 20;
  ctx.clearRect(0, 0, width, height);
  const css = getComputedStyle(document.documentElement);
  drawGrid(ctx, width, height, pad, css.getPropertyValue("--line"));

  const merged = [...valuesA, ...valuesB];
  if (!merged.length) return;
  const maxY = Math.max(...merged, 1);
  drawLine(ctx, valuesA, colorA, 0, maxY, false);
  drawLineOverlay(ctx, valuesB, colorB, 0, maxY);
}

function drawLineOverlay(ctx, values, color, minY, maxY) {
  const width = ctx.canvas.width;
  const height = ctx.canvas.height;
  const pad = 20;
  if (!values.length) return;
  const range = Math.max(maxY - minY, 1);
  const points = values.map((v, i) => {
    const x = pad + (i / Math.max(values.length - 1, 1)) * (width - pad * 2);
    const normalized = (Math.min(Math.max(v, minY), maxY) - minY) / range;
    const y = pad + (1 - normalized) * (height - pad * 2);
    return { x, y };
  });
  ctx.beginPath();
  points.forEach((p, idx) => (idx === 0 ? ctx.moveTo(p.x, p.y) : ctx.lineTo(p.x, p.y)));
  ctx.strokeStyle = color;
  ctx.lineWidth = 2.2;
  ctx.stroke();
}

function renderSummary(sum) {
  hostMetaEl.textContent = `${sum.hostname} • ${sum.os}/${sum.arch} • uptime ${Math.floor(sum.uptime_sec / 60)}m`;
  cpuEl.textContent = pct(sum.cpu.usage_pct);
  memEl.textContent = pct(sum.memory.used_pct);
  diskEl.textContent = pct(sum.disk.used_pct);
  netEl.textContent = `↑ ${mbps(sum.network.tx_bytes_per_sec)} Mbps / ↓ ${mbps(sum.network.rx_bytes_per_sec)} Mbps`;

  cpuMetaEl.textContent = `Current ${pct(sum.cpu.usage_pct)}`;
  memMetaEl.textContent = `${bytes(sum.memory.used_bytes)} / ${bytes(sum.memory.total_bytes)}`;
  diskMetaEl.textContent = `${bytes(sum.disk.used_bytes)} / ${bytes(sum.disk.total_bytes)}`;
  netMetaEl.textContent = `tx ${bytes(sum.network.tx_bytes_per_sec)}/s • rx ${bytes(sum.network.rx_bytes_per_sec)}/s`;

  const cpuVals = (sum.trends.cpu_usage_pct || []).map((p) => p.val);
  const memVals = (sum.trends.memory_pct || []).map((p) => p.val);
  const diskVals = cpuVals.map(() => sum.disk.used_pct || 0);
  const netRxVals = (sum.trends.network_rx || []).map((p) => p.val / (1024 * 1024));
  const netTxVals = (sum.trends.network_tx || []).map((p) => p.val / (1024 * 1024));

  drawLine(charts.cpu, cpuVals, "#2d8dff", 0, 100, false);
  drawLine(charts.mem, memVals, "#2ee6a4", 0, 100, true);
  drawLine(charts.disk, diskVals, "#ffc544", 0, 100, false);
  drawDualLine(charts.net, netTxVals, "#9e6cff", netRxVals, "#3d94ff");
  drawMini(miniCharts.cpu, cpuVals, "#2d8dff", 0, 100);
  drawMini(miniCharts.mem, memVals, "#2ee6a4", 0, 100);
  drawMini(miniCharts.disk, diskVals, "#ffc544", 0, 100);
  drawMini(miniCharts.net, netRxVals, "#8f74ff", 0, Math.max(...netRxVals, ...netTxVals, 1));

  cpuLegendEl.textContent = `Current ${pct(sum.cpu.usage_pct)}`;
  memLegendEl.textContent = `${bytes(sum.memory.used_bytes)} / ${bytes(sum.memory.total_bytes)}`;
  diskLegendEl.textContent = `${bytes(sum.disk.used_bytes)} / ${bytes(sum.disk.total_bytes)}`;
  netLegendEl.textContent = `↑ ${mbps(sum.network.tx_bytes_per_sec)} Mbps / ↓ ${mbps(sum.network.rx_bytes_per_sec)} Mbps`;
}

async function loadResources() {
  const res = await fetch("/api/v1/resources");
  const data = await res.json();
  resEl.textContent = JSON.stringify(data, null, 2);
}

async function loadVersion() {
  const res = await fetch("/api/v1/version");
  const data = await res.json();
  buildMetaEl.textContent = `${data.version} (${data.commit}) • ${data.build_time}`;
}

function setState(online) {
  connEl.textContent = online ? "live" : "offline";
  connEl.className = `badge ${online ? "online" : "offline"}`;
}

function connectSSE() {
  const es = new EventSource("/api/v1/stream");
  es.addEventListener("summary", (ev) => {
    setState(true);
    renderSummary(JSON.parse(ev.data));
  });
  es.onerror = () => {
    setState(false);
    es.close();
    setTimeout(connectSSE, 1500);
  };
}

loadResources().catch(() => {
  resEl.textContent = "failed to load resources";
});
loadVersion().catch(() => {
  buildMetaEl.textContent = "build metadata unavailable";
});
connectSSE();
