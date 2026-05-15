const connEl = document.getElementById("connState");
const hostMetaEl = document.getElementById("hostMeta");
const buildMetaEl = document.getElementById("buildMeta");
const cpuEl = document.getElementById("cpu");
const memEl = document.getElementById("memory");
const diskEl = document.getElementById("disk");
const netEl = document.getElementById("network");
const resEl = document.getElementById("resources");
const themeToggleEl = document.getElementById("themeToggle");
const canvas = document.getElementById("trendChart");
const ctx = canvas.getContext("2d");

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

function drawTrends(sum) {
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  const pad = 28;
  const w = canvas.width - pad * 2;
  const h = canvas.height - pad * 2;

  const cpu = sum.trends.cpu_usage_pct || [];
  if (!cpu.length) return;

  const points = cpu.map((p, i) => ({
    x: pad + (i / Math.max(cpu.length - 1, 1)) * w,
    y: pad + h - (Math.min(Math.max(p.val, 0), 100) / 100) * h,
  }));

  ctx.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue("--grid");
  ctx.lineWidth = 1;
  for (let i = 0; i <= 4; i++) {
    const y = pad + (i / 4) * h;
    ctx.beginPath();
    ctx.moveTo(pad, y);
    ctx.lineTo(pad + w, y);
    ctx.stroke();
  }

  ctx.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue("--accent");
  ctx.lineWidth = 2;
  ctx.beginPath();
  points.forEach((p, idx) => idx === 0 ? ctx.moveTo(p.x, p.y) : ctx.lineTo(p.x, p.y));
  ctx.stroke();
}

function renderSummary(sum) {
  hostMetaEl.textContent = `${sum.hostname} • ${sum.os}/${sum.arch} • uptime ${Math.floor(sum.uptime_sec / 60)}m`;
  cpuEl.textContent = `${sum.cpu.usage_pct.toFixed(1)}%`;
  memEl.textContent = `${sum.memory.used_pct.toFixed(1)}% (${bytes(sum.memory.used_bytes)} / ${bytes(sum.memory.total_bytes)})`;
  diskEl.textContent = `${sum.disk.used_pct.toFixed(1)}% (${bytes(sum.disk.used_bytes)} / ${bytes(sum.disk.total_bytes)})`;
  netEl.textContent = `↓ ${bytes(sum.network.rx_bytes_per_sec)}/s  ↑ ${bytes(sum.network.tx_bytes_per_sec)}/s`;
  drawTrends(sum);
}

async function loadResources() {
  const res = await fetch("/api/v1/resources");
  const data = await res.json();
  resEl.textContent = JSON.stringify(data, null, 2);
}

async function loadVersion() {
  const res = await fetch("/api/v1/version");
  const data = await res.json();
  buildMetaEl.textContent = `build: ${data.version} (${data.commit}) • ${data.build_time} • ${data.go_version}`;
}

function setState(online) {
  connEl.textContent = online ? "online" : "offline";
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
  buildMetaEl.textContent = "build: unavailable";
});
connectSSE();
