package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/SirsiMaster/sirsi-pantheon/internal/stele"
)

// ── Shared Layout ───────────────────────────────────────────────────────

// pageShell wraps page-specific content in the shared dashboard layout.
// activePage is the nav item to highlight (e.g., "/", "/scan", "/guard").
func pageShell(title, activePage, bodyContent string) string {
	navItems := []struct {
		Path  string
		Glyph string
		Label string
	}{
		{"/", "☥", "Overview"},
		{"/scan", "𓁢", "Scan"},
		{"/ghosts", "𓂓", "Ghosts"},
		{"/guard", "🛡", "Guard"},
		{"/notifications", "🔔", "Notifications"},
		{"/horus", "𓂀", "Horus"},
		{"/vault", "🏛", "Vault"},
	}

	var navHTML strings.Builder
	for _, n := range navItems {
		cls := "nav-item"
		if n.Path == activePage {
			cls += " active"
		}
		navHTML.WriteString(fmt.Sprintf(
			`<a href="%s" class="%s"><span class="nav-glyph">%s</span><span class="nav-label">%s</span></a>`,
			n.Path, cls, n.Glyph, n.Label,
		))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s — Sirsi Pantheon</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
@font-face{font-family:'Cinzel';font-style:normal;font-weight:400;font-display:swap;
src:local('Cinzel Regular'),local('Cinzel-Regular')}
body{background:%s;color:%s;font-family:Inter,-apple-system,system-ui,'Segoe UI',sans-serif;
display:flex;min-height:100vh;overflow-x:hidden}
::-webkit-scrollbar{width:6px}
::-webkit-scrollbar-track{background:transparent}
::-webkit-scrollbar-thumb{background:rgba(200,169,81,.2);border-radius:3px}
::-webkit-scrollbar-thumb:hover{background:rgba(200,169,81,.4)}

/* Sidebar */
.sidebar{width:220px;min-height:100vh;background:rgba(6,6,15,.96);border-right:1px solid %s;
padding:24px 0;display:flex;flex-direction:column;position:fixed;left:0;top:0;bottom:0;z-index:10}
.sidebar-brand{padding:0 20px 24px;border-bottom:1px solid %s}
.sidebar-brand h1{font-family:Cinzel,Georgia,'Times New Roman',serif;font-size:16px;font-weight:400;
color:%s;letter-spacing:3px;text-transform:uppercase}
.sidebar-brand p{font-size:9px;color:%s;letter-spacing:1.5px;margin-top:4px;text-transform:uppercase}
.sidebar-nav{flex:1;padding:16px 0}
.nav-item{display:flex;align-items:center;padding:10px 20px;color:%s;text-decoration:none;
font-size:13px;letter-spacing:.3px;transition:all .2s;border-left:2px solid transparent}
.nav-item:hover{background:rgba(200,169,81,.06);color:%s}
.nav-item.active{background:rgba(200,169,81,.08);color:%s;border-left-color:%s}
.nav-glyph{width:24px;font-size:15px;margin-right:10px;text-align:center}
.sidebar-footer{padding:16px 20px;border-top:1px solid %s;font-size:9px;color:#333;letter-spacing:1px}

/* Main */
.main{margin-left:220px;flex:1;padding:32px 40px;min-height:100vh}
.page-title{font-family:Cinzel,Georgia,'Times New Roman',serif;font-size:22px;font-weight:400;
color:%s;letter-spacing:2px;margin-bottom:24px}
.page-subtitle{font-size:11px;color:%s;letter-spacing:1.5px;text-transform:uppercase;margin-bottom:20px}

/* Cards */
.card{background:%s;border:1px solid %s;border-radius:12px;padding:20px 24px;margin-bottom:16px;
backdrop-filter:blur(12px);box-shadow:0 4px 20px rgba(0,0,0,.3);transition:border-color .2s,box-shadow .2s,transform .15s}
.card-link{cursor:pointer;text-decoration:none;display:block;color:inherit}
.card-link:hover .card,.card.clickable:hover{border-color:rgba(200,169,81,.45);
box-shadow:0 4px 24px rgba(200,169,81,.1);transform:translateY(-2px)}
.card-title{font-size:10px;color:%s;letter-spacing:2px;text-transform:uppercase;margin-bottom:12px;font-weight:600}
.card-value{font-size:28px;font-weight:300;color:%s}
.card-label{font-size:11px;color:%s;margin-top:4px}

/* Grid */
.grid{display:grid;gap:16px}
.grid-2{grid-template-columns:repeat(2,1fr)}
.grid-3{grid-template-columns:repeat(3,1fr)}
.grid-4{grid-template-columns:repeat(4,1fr)}

/* Table */
.tbl{width:100%%;border-collapse:collapse}
.tbl th{text-align:left;font-size:10px;color:%s;letter-spacing:1.5px;text-transform:uppercase;
padding:10px 14px;border-bottom:1px solid %s;font-weight:600}
.tbl td{padding:10px 14px;font-size:13px;border-bottom:1px solid rgba(200,169,81,.06);color:%s}
.tbl tr:hover td{background:rgba(200,169,81,.03)}

/* Severity badges */
.badge{display:inline-block;padding:2px 8px;border-radius:4px;font-size:10px;font-weight:600;letter-spacing:.5px}
.badge-success{background:rgba(68,255,136,.12);color:%s}
.badge-error{background:rgba(255,68,68,.12);color:%s}
.badge-warning{background:rgba(255,215,0,.12);color:%s}
.badge-info{background:rgba(81,169,200,.12);color:#51A9C8}

/* Pulse animation */
@keyframes pulse{0%%,100%%{opacity:.6}50%%{opacity:1}}
.pulse{animation:pulse 2s ease-in-out infinite}

/* Search */
.search-box{background:rgba(6,6,15,.6);border:1px solid %s;border-radius:8px;padding:10px 16px;
color:%s;font-size:14px;width:100%%;outline:none;transition:border-color .2s}
.search-box:focus{border-color:%s}
.search-box::placeholder{color:#444}

/* Empty state */
.empty{text-align:center;padding:60px 20px;color:#444;font-size:13px;letter-spacing:.5px}
.empty-glyph{font-size:40px;margin-bottom:12px;opacity:.3}
</style>
</head>
<body>
<div class="sidebar">
 <div class="sidebar-brand">
  <h1>☥ Pantheon</h1>
  <p>Infrastructure Dashboard</p>
 </div>
 <nav class="sidebar-nav">%s</nav>
 <div class="sidebar-footer">LOCAL ONLY • 127.0.0.1:%d</div>
</div>
<div class="main">%s</div>
</body>
</html>`,
		title,
		ColorBg, ColorWhite,
		ColorBorder, ColorBorder,
		ColorGold, ColorDim,
		ColorDim, ColorWhite, ColorGold, ColorGold,
		ColorBorder,
		ColorGold, ColorDim,
		ColorBgPanel, ColorBorder,
		ColorGold, ColorGold, ColorDim,
		ColorGold, ColorBorder, ColorWhite,
		ColorGreen, ColorRed, ColorYellow,
		ColorBorder, ColorWhite, ColorGold,
		navHTML.String(),
		DashboardPort,
		bodyContent,
	)
}

// safeTextJS is a JavaScript helper function injected into pages that need to render
// dynamic data. It escapes HTML entities to prevent XSS when inserting into the DOM.
const safeTextJS = `function esc(s){if(!s)return'';const d=document.createElement('div');d.textContent=s;return d.innerHTML}`

// ── Overview Page ───────────────────────────────────────────────────────

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	statsJSON := "null"
	if s.cfg.StatsFn != nil {
		if data, err := s.cfg.StatsFn(); err == nil {
			statsJSON = string(data)
		}
	}

	recentJSON := "[]"
	if s.cfg.NotifyDB != nil {
		if recent, err := s.cfg.NotifyDB.Recent(8); err == nil && recent != nil {
			if b, err := json.Marshal(recent); err == nil {
				recentJSON = string(b)
			}
		}
	}

	body := fmt.Sprintf(`
<h1 class="page-title">Command Center</h1>

<!-- ── Status Bar ─────────────────────────────────────────────── -->
<div class="grid grid-4" style="margin-bottom:20px">
 <a href="/guard" class="card-link"><div class="card"><div class="card-title">RAM</div>
  <div class="card-value" id="ram-val" style="font-size:22px">—</div>
  <div class="card-label" id="ram-label"></div></div></a>
 <a href="/notifications" class="card-link"><div class="card"><div class="card-title">Git</div>
  <div class="card-value" id="git-val" style="font-size:22px">—</div>
  <div class="card-label" id="git-label"></div></div></a>
 <a href="/scan" class="card-link"><div class="card"><div class="card-title">Deities</div>
  <div class="card-value" id="deity-val" style="font-size:22px">0</div>
  <div class="card-label" id="deity-label">None</div></div></a>
 <a href="/horus" class="card-link"><div class="card"><div class="card-title">Platform</div>
  <div class="card-value" id="accel-val" style="font-size:22px">—</div>
  <div class="card-label" id="accel-label"></div></div></a>
</div>

<!-- ── Action Buttons ─────────────────────────────────────────── -->
<h2 class="page-subtitle">Actions</h2>
<div class="grid grid-4" style="margin-bottom:20px" id="actions-grid">
 <button class="action-btn" data-cmd="scan" id="btn-scan"><span class="action-glyph">𓁢</span>Scan</button>
 <button class="action-btn" data-cmd="ghosts" id="btn-ghosts"><span class="action-glyph">𓂓</span>Ghost Hunt</button>
 <button class="action-btn" data-cmd="doctor" id="btn-doctor"><span class="action-glyph">𓁐</span>Doctor</button>
 <button class="action-btn" data-cmd="quality" id="btn-quality"><span class="action-glyph">𓆄</span>Quality</button>
 <button class="action-btn" data-cmd="network" id="btn-network"><span class="action-glyph">🌐</span>Network</button>
 <button class="action-btn" data-cmd="hardware" id="btn-hardware"><span class="action-glyph">⚡</span>Hardware</button>
 <button class="action-btn" data-cmd="dedup" id="btn-dedup"><span class="action-glyph">🔍</span>Duplicates</button>
 <button class="action-btn" data-cmd="guard" id="btn-guard"><span class="action-glyph">🛡</span>Guard</button>
</div>

<!-- ── Live Terminal ──────────────────────────────────────────── -->
<div style="display:flex;align-items:center;gap:12px;margin-bottom:8px">
 <h2 class="page-subtitle" style="margin-bottom:0;flex:1">Terminal</h2>
 <span id="run-status" style="font-size:11px;color:#444;letter-spacing:.5px"></span>
 <button id="term-clear" style="background:none;border:1px solid rgba(200,169,81,.15);color:#666;
  font-size:10px;padding:4px 10px;border-radius:4px;cursor:pointer;letter-spacing:.5px;
  transition:all .2s" onmouseover="this.style.borderColor='rgba(200,169,81,.4)';this.style.color='#C8A951'"
  onmouseout="this.style.borderColor='rgba(200,169,81,.15)';this.style.color='#666'">CLEAR</button>
</div>
<div class="card" id="terminal" style="min-height:280px;max-height:50vh;overflow-y:auto;font-family:'SF Mono',Menlo,Consolas,monospace;
 font-size:12px;padding:16px;line-height:1.6;background:rgba(3,3,8,.95);border-color:rgba(200,169,81,.08)">
 <div class="term-line term-dim">☥ Pantheon ready — click an action above to begin</div>
</div>

<!-- ── Recent Activity (compact) ─────────────────────────────── -->
<div style="margin-top:20px">
 <a href="/notifications" style="text-decoration:none"><h2 class="page-subtitle" style="cursor:pointer;transition:color .2s"
  onmouseover="this.style.color='#C8A951'" onmouseout="this.style.color=''">Recent Activity ›</h2></a>
 <div class="card" style="padding:0;overflow:hidden">
  <table class="tbl" id="recent-tbl">
   <thead><tr><th>Source</th><th>Summary</th><th>Status</th><th>Time</th></tr></thead>
   <tbody id="recent-body"></tbody>
  </table>
 </div>
</div>

<style>
.action-btn{background:rgba(6,6,15,.88);border:1px solid rgba(200,169,81,.12);border-radius:10px;
 color:#FAFAFA;padding:16px;font-size:13px;cursor:pointer;text-align:center;letter-spacing:.3px;
 transition:all .2s;display:flex;flex-direction:column;align-items:center;gap:8px;font-family:inherit}
.action-btn:hover{border-color:rgba(200,169,81,.45);background:rgba(200,169,81,.06);
 box-shadow:0 4px 20px rgba(200,169,81,.08);transform:translateY(-1px)}
.action-btn:active{transform:translateY(0);box-shadow:none}
.action-btn.running{border-color:#C8A951;color:#C8A951;animation:pulse 1.5s ease-in-out infinite}
.action-btn:disabled{opacity:.4;cursor:not-allowed;transform:none!important;box-shadow:none!important}
.action-glyph{font-size:24px;line-height:1}
.term-line{margin:0;white-space:pre-wrap;word-break:break-all}
.term-dim{color:#555}
.term-out{color:#ccc}
.term-ok{color:#44FF88}
.term-err{color:#FF4444}
.term-info{color:#C8A951}
</style>

<script>
(function(){
'use strict';
%s
const S=%s,R=%s;
const sevBadge=s=>({success:'badge-success',error:'badge-error',warning:'badge-warning'}[s]||'badge-info');
const sevIcon=s=>({success:'✅',error:'❌',warning:'⚠️',info:'ℹ️'}[s]||'ℹ️');
const ago=ts=>{if(!ts)return'—';const d=Date.now()-new Date(ts).getTime();
if(d<60e3)return Math.floor(d/1e3)+'s ago';if(d<3600e3)return Math.floor(d/6e4)+'m ago';
if(d<864e5)return Math.floor(d/36e5)+'h ago';return Math.floor(d/864e5)+'d ago'};

/* ── Stats ────────────────────────────────────────────── */
function renderStats(s){
 if(!s)return;
 document.getElementById('ram-val').textContent=(s.ram_icon||'')+' '+Math.round(s.ram_percent||0)+'%%';
 document.getElementById('ram-label').textContent=(s.ram_pressure||'unknown');
 document.getElementById('git-val').textContent=(s.osiris_icon||'')+' '+(s.uncommitted_files||0)+' dirty';
 document.getElementById('git-label').textContent=(s.git_branch||'')+(s.time_since_commit?' • '+s.time_since_commit:'');
 document.getElementById('deity-val').textContent=s.deity_count||0;
 document.getElementById('deity-label').textContent=(s.active_deities||[]).join(', ')||'None';
 document.getElementById('accel-val').textContent=s.accel_icon||'💻';
 document.getElementById('accel-label').textContent=s.primary_accelerator||'Unknown';
}

/* ── Activity table ───────────────────────────────────── */
function renderRecent(items){
 const tb=document.getElementById('recent-body');
 tb.textContent='';
 if(!items||!items.length){const tr=document.createElement('tr');const td=document.createElement('td');
  td.colSpan=4;td.className='empty';td.textContent='No activity yet';tr.appendChild(td);tb.appendChild(tr);return}
 items.forEach(function(n){
  const tr=document.createElement('tr');tr.style.cursor='pointer';
  tr.addEventListener('click',function(){window.location.href='/notifications'});
  const tdSrc=document.createElement('td');tdSrc.style.fontWeight='600';tdSrc.textContent=n.source;
  const tdSum=document.createElement('td');const summary=(n.summary||'');
  tdSum.textContent=summary.length>80?summary.substring(0,77)+'…':summary;
  const tdSev=document.createElement('td');const badge=document.createElement('span');
  badge.className='badge '+sevBadge(n.severity);badge.textContent=sevIcon(n.severity)+' '+n.severity;tdSev.appendChild(badge);
  const tdTime=document.createElement('td');tdTime.style.cssText='color:#666;font-size:11px;white-space:nowrap';
  tdTime.textContent=ago(n.timestamp);
  tr.appendChild(tdSrc);tr.appendChild(tdSum);tr.appendChild(tdSev);tr.appendChild(tdTime);tb.appendChild(tr)});
}

/* ── Terminal ─────────────────────────────────────────── */
const term=document.getElementById('terminal');
const statusEl=document.getElementById('run-status');
let running=false;

function termLine(text,cls){
 const d=document.createElement('div');
 d.className='term-line '+(cls||'term-out');
 d.textContent=text;
 term.appendChild(d);
 if(term.children.length>500)term.removeChild(term.firstChild);
 term.scrollTop=term.scrollHeight;
}

document.getElementById('term-clear').addEventListener('click',function(){
 term.textContent='';
 termLine('☥ Pantheon ready — click an action above to begin','term-dim');
});

/* ── Action buttons ───────────────────────────────────── */
function setRunning(key,label){
 running=true;
 statusEl.textContent='⟳ Running '+label+'…';
 statusEl.style.color='#C8A951';
 document.querySelectorAll('.action-btn').forEach(function(b){
  if(b.dataset.cmd===key){b.classList.add('running')}
  else{b.disabled=true}
 });
}

function setIdle(){
 running=false;
 statusEl.textContent='';
 document.querySelectorAll('.action-btn').forEach(function(b){
  b.classList.remove('running');
  b.disabled=false;
 });
}

document.querySelectorAll('.action-btn').forEach(function(btn){
 btn.addEventListener('click',function(){
  if(running)return;
  const cmd=this.dataset.cmd;
  const label=this.textContent.trim();
  termLine('','term-dim');
  termLine('▸ '+label,'term-info');
  fetch('/api/run?cmd='+encodeURIComponent(cmd),{method:'POST'})
   .then(function(r){
    if(!r.ok)return r.json().then(function(e){throw new Error(e.error||'failed')});
    setRunning(cmd,label);
   })
   .catch(function(e){termLine('✗ '+e.message,'term-err')});
 });
});

/* ── SSE event stream ─────────────────────────────────── */
if(typeof EventSource!=='undefined'){
 const es=new EventSource('/api/events');
 es.addEventListener('run_start',function(e){
  try{const d=JSON.parse(e.data);setRunning(d.key,d.label)}catch(x){}
 });
 es.addEventListener('run_output',function(e){
  try{const d=JSON.parse(e.data);termLine(d.line)}catch(x){}
 });
 es.addEventListener('run_complete',function(e){
  try{const d=JSON.parse(e.data);
   if(d.status==='success'){
    termLine('✓ '+d.label+' completed ('+d.duration_ms+'ms)','term-ok');
   }else{
    termLine('✗ '+d.label+' failed: '+(d.error||'unknown'),'term-err');
   }
   setIdle();
   /* Refresh activity table after command completes */
   fetch('/api/notifications?limit=8').then(function(r){return r.json()}).then(renderRecent).catch(function(){});
  }catch(x){setIdle()}
 });
}

/* ── Init + polling ───────────────────────────────────── */
renderStats(S);renderRecent(R);
setInterval(function(){
 fetch('/api/stats').then(function(r){return r.json()}).then(renderStats).catch(function(){});
 fetch('/api/notifications?limit=8').then(function(r){return r.json()}).then(renderRecent).catch(function(){});
},10000);

/* Check if something is already running on page load */
fetch('/api/run/status').then(function(r){return r.json()}).then(function(d){
 if(d.running)setRunning(d.current,d.current);
}).catch(function(){});
})();
</script>`, safeTextJS, statsJSON, recentJSON)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, pageShell("Overview", "/", body))
}

// ── Notifications Page ──────────────────────────────────────────────────

func (s *Server) handleNotifications(w http.ResponseWriter, r *http.Request) {
	initialJSON := "[]"
	if s.cfg.NotifyDB != nil {
		if items, err := s.cfg.NotifyDB.Recent(200); err == nil && items != nil {
			if b, err := json.Marshal(items); err == nil {
				initialJSON = string(b)
			}
		}
	}

	var count int64
	if s.cfg.NotifyDB != nil {
		count, _ = s.cfg.NotifyDB.Count()
	}

	body := fmt.Sprintf(`
<h1 class="page-title">Notification History</h1>
<p class="page-subtitle">%d total notifications</p>

<div style="display:flex;gap:12px;margin-bottom:20px">
 <input type="text" class="search-box" id="filter-input" placeholder="Filter by source, summary, or severity..." style="flex:1">
 <select id="sev-filter" style="background:rgba(6,6,15,.6);border:1px solid %s;border-radius:8px;
  padding:8px 12px;color:%s;font-size:13px;outline:none">
  <option value="">All Severities</option>
  <option value="success">Success</option>
  <option value="error">Error</option>
  <option value="warning">Warning</option>
  <option value="info">Info</option>
 </select>
</div>

<div class="card" style="padding:0;overflow:hidden">
 <table class="tbl">
  <thead><tr><th>Time</th><th>Source</th><th>Action</th><th>Summary</th><th>Status</th><th>Duration</th></tr></thead>
  <tbody id="ntf-body"></tbody>
 </table>
</div>

<script>
(function(){
'use strict';
%s
const D=%s;
const sevBadge=s=>({success:'badge-success',error:'badge-error',warning:'badge-warning'}[s]||'badge-info');
const sevIcon=s=>({success:'✅',error:'❌',warning:'⚠️',info:'ℹ️'}[s]||'ℹ️');
const fmtDur=ms=>{if(!ms)return'—';if(ms<1000)return ms+'ms';if(ms<60000)return(ms/1000).toFixed(1)+'s';return(ms/60000).toFixed(1)+'m'};
const fmtTime=ts=>{if(!ts)return'—';const d=new Date(ts);return d.toLocaleDateString('en-US',{month:'short',day:'numeric'})+' '+d.toLocaleTimeString('en-US',{hour:'2-digit',minute:'2-digit'})};

let filtered=D;
function render(){
 const tb=document.getElementById('ntf-body');
 tb.textContent='';
 if(!filtered.length){const tr=document.createElement('tr');const td=document.createElement('td');
  td.colSpan=6;td.className='empty';td.textContent='No notifications match your filter';
  tr.appendChild(td);tb.appendChild(tr);return}
 filtered.forEach(function(n){
  const tr=document.createElement('tr');
  const tdTime=document.createElement('td');tdTime.style.cssText='color:#666;font-size:11px;white-space:nowrap';
  tdTime.textContent=fmtTime(n.timestamp);
  const tdSrc=document.createElement('td');tdSrc.style.fontWeight='600';tdSrc.textContent=n.source;
  const tdAct=document.createElement('td');tdAct.textContent=n.action;
  const tdSum=document.createElement('td');tdSum.textContent=n.summary||'—';
  const tdSev=document.createElement('td');const badge=document.createElement('span');
  badge.className='badge '+sevBadge(n.severity);badge.textContent=sevIcon(n.severity)+' '+n.severity;tdSev.appendChild(badge);
  const tdDur=document.createElement('td');tdDur.style.cssText='color:#666;font-size:11px';tdDur.textContent=fmtDur(n.duration_ms);
  tr.appendChild(tdTime);tr.appendChild(tdSrc);tr.appendChild(tdAct);tr.appendChild(tdSum);tr.appendChild(tdSev);tr.appendChild(tdDur);
  tb.appendChild(tr)});
}

function applyFilter(){
 const q=document.getElementById('filter-input').value.toLowerCase();
 const sev=document.getElementById('sev-filter').value;
 filtered=D.filter(function(n){
  if(sev&&n.severity!==sev)return false;
  if(!q)return true;
  return(n.source||'').toLowerCase().includes(q)||(n.summary||'').toLowerCase().includes(q)||
   (n.action||'').toLowerCase().includes(q)});
 render();
}

document.getElementById('filter-input').addEventListener('input',applyFilter);
document.getElementById('sev-filter').addEventListener('change',applyFilter);
render();
})();
</script>`, count, ColorBorder, ColorWhite, safeTextJS, initialJSON)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, pageShell("Notifications", "/notifications", body))
}

// ── Guard Page ──────────────────────────────────────────────────────────

func (s *Server) handleGuard(w http.ResponseWriter, r *http.Request) {
	body := fmt.Sprintf(`
<h1 class="page-title">🛡 Guard — System Monitor</h1>

<div class="grid grid-3" style="margin-bottom:20px">
 <div class="card">
  <div class="card-title">RAM Usage</div>
  <canvas id="ram-chart" width="320" height="100" style="width:100%%;height:100px"></canvas>
 </div>
 <div class="card">
  <div class="card-title">RAM Pressure</div>
  <div class="card-value pulse" id="ram-pct" style="font-size:48px;text-align:center;padding:12px 0">—</div>
  <div class="card-label" id="ram-state" style="text-align:center"></div>
 </div>
 <div class="card">
  <div class="card-title">System Health</div>
  <div class="card-value" id="health-score" style="font-size:48px;text-align:center;padding:12px 0">—</div>
  <div class="card-label" id="health-label" style="text-align:center">
   <button class="clean-btn" id="btn-doctor" style="margin-top:4px">Run Doctor</button></div>
 </div>
</div>

<h2 class="page-subtitle">Process Slayer</h2>
<div class="grid grid-3" style="margin-bottom:12px">
 <button class="action-btn slay-btn" data-target="node" style="flex-direction:row;padding:10px 16px;gap:6px;font-size:12px">
  📦 Kill Node</button>
 <button class="action-btn slay-btn" data-target="electron" style="flex-direction:row;padding:10px 16px;gap:6px;font-size:12px">
  ⚡ Kill Electron</button>
 <button class="action-btn slay-btn" data-target="docker" style="flex-direction:row;padding:10px 16px;gap:6px;font-size:12px">
  🐳 Kill Docker</button>
 <button class="action-btn slay-btn" data-target="lsp" style="flex-direction:row;padding:10px 16px;gap:6px;font-size:12px">
  🔤 Kill LSP</button>
 <button class="action-btn slay-btn" data-target="build" style="flex-direction:row;padding:10px 16px;gap:6px;font-size:12px">
  🔨 Kill Builds</button>
 <button class="action-btn slay-btn" data-target="ai" style="flex-direction:row;padding:10px 16px;gap:6px;font-size:12px">
  🤖 Kill AI</button>
</div>
<div id="slay-result" style="font-size:12px;margin-bottom:20px;color:#666"></div>

<div id="doctor-results" style="display:none">
 <h2 class="page-subtitle">Health Diagnostics</h2>
 <div class="card" style="padding:0;overflow:hidden">
  <table class="tbl"><thead><tr><th></th><th>Check</th><th>Result</th></tr></thead>
   <tbody id="doctor-body"></tbody></table>
 </div>
</div>

<h2 class="page-subtitle" style="margin-top:20px">Alert History</h2>
<div class="card" style="padding:0;overflow:hidden">
 <table class="tbl">
  <thead><tr><th>Time</th><th>Summary</th><th>Status</th></tr></thead>
  <tbody id="guard-body"><tr><td colspan="3" class="empty">Loading...</td></tr></tbody>
 </table>
</div>

<script>
(function(){
'use strict';
const ramHistory=[];const maxPoints=60;
const canvas=document.getElementById('ram-chart');const ctx=canvas.getContext('2d');
function drawChart(){
 const w=canvas.width,h=canvas.height;ctx.clearRect(0,0,w,h);if(!ramHistory.length)return;
 ctx.strokeStyle='rgba(200,169,81,.08)';ctx.lineWidth=1;
 for(let y=0;y<=100;y+=25){const py=h-y/100*h;ctx.beginPath();ctx.moveTo(0,py);ctx.lineTo(w,py);ctx.stroke()}
 ctx.strokeStyle='%s';ctx.lineWidth=2;ctx.beginPath();
 ramHistory.forEach(function(v,i){const x=i/(maxPoints-1)*w,y=h-v/100*h;i===0?ctx.moveTo(x,y):ctx.lineTo(x,y)});
 ctx.stroke();const grad=ctx.createLinearGradient(0,0,0,h);
 grad.addColorStop(0,'rgba(200,169,81,.15)');grad.addColorStop(1,'transparent');
 ctx.fillStyle=grad;ctx.lineTo((ramHistory.length-1)/(maxPoints-1)*w,h);ctx.lineTo(0,h);ctx.fill()}
function refresh(){
 fetch('/api/stats').then(r=>r.json()).then(function(s){
  ramHistory.push(s.ram_percent||0);if(ramHistory.length>maxPoints)ramHistory.shift();drawChart();
  document.getElementById('ram-pct').textContent=Math.round(s.ram_percent||0)+'%%';
  document.getElementById('ram-state').textContent=(s.ram_pressure||'unknown')+' pressure';
 }).catch(function(){});
 fetch('/api/notifications?source=isis&limit=10').then(r=>r.json()).then(function(items){
  const tb=document.getElementById('guard-body');tb.textContent='';
  if(!items||!items.length){const tr=document.createElement('tr');const td=document.createElement('td');
   td.colSpan=3;td.className='empty';td.textContent='No alerts';tr.appendChild(td);tb.appendChild(tr);return}
  items.forEach(function(n){const tr=document.createElement('tr');
   const t=document.createElement('td');t.style.cssText='color:#666;font-size:11px';
   t.textContent=new Date(n.timestamp).toLocaleTimeString();
   const s=document.createElement('td');s.textContent=n.summary;
   const b=document.createElement('td');const badge=document.createElement('span');
   badge.className='badge badge-'+({success:'success',error:'error',warning:'warning'}[n.severity]||'info');
   badge.textContent=n.severity;b.appendChild(badge);
   tr.appendChild(t);tr.appendChild(s);tr.appendChild(b);tb.appendChild(tr)});
 }).catch(function(){});
}
document.getElementById('btn-doctor').addEventListener('click',function(){
 const btn=this;btn.textContent='Running...';btn.disabled=true;
 fetch('/api/doctor').then(r=>r.json()).then(function(rpt){
  document.getElementById('health-score').textContent=rpt.Score+'/100';
  document.getElementById('health-label').textContent=rpt.Score>=75?'Healthy':rpt.Score>=50?'Degraded':'Critical';
  document.getElementById('doctor-results').style.display='';
  const tb=document.getElementById('doctor-body');tb.textContent='';
  (rpt.Findings||[]).forEach(function(f){const tr=document.createElement('tr');
   const i=document.createElement('td');i.textContent=({0:'✅',1:'ℹ️',2:'⚠️',3:'🔴'}[f.Severity]||'⚪');
   const c=document.createElement('td');c.style.fontWeight='600';c.textContent=f.Check;
   const m=document.createElement('td');m.textContent=f.Message;
   tr.appendChild(i);tr.appendChild(c);tr.appendChild(m);tb.appendChild(tr)});
  btn.textContent='Run Doctor';btn.disabled=false;
 }).catch(function(){btn.textContent='Error';btn.disabled=false});
});
document.querySelectorAll('.slay-btn').forEach(function(btn){
 btn.addEventListener('click',function(){
  const target=this.dataset.target;const resEl=document.getElementById('slay-result');
  btn.disabled=true;
  fetch('/api/slay?target='+target+'&dry_run=false',{method:'POST'}).then(r=>r.json()).then(function(d){
   btn.disabled=false;resEl.textContent='';
   const msg=document.createElement('span');
   if(d.killed>0){msg.style.color='#44FF88';msg.textContent='✓ Killed '+d.killed+' '+target+' processes'}
   else{msg.style.color='#888';msg.textContent='No '+target+' processes found'}
   resEl.appendChild(msg);
  }).catch(function(){btn.disabled=false});
 });
});
refresh();setInterval(refresh,20000);
})();
</script>`, ColorGold)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, pageShell("Guard", "/guard", body))
}

// ── Scan Page ───────────────────────────────────────────────────────────

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	body := `
<h1 class="page-title">𓁢 Scan Results</h1>

<!-- Summary cards -->
<div class="grid grid-4" style="margin-bottom:20px" id="scan-summary">
 <div class="card"><div class="card-title">Total Waste</div>
  <div class="card-value" id="total-waste" style="font-size:22px">—</div></div>
 <div class="card"><div class="card-title">Findings</div>
  <div class="card-value" id="total-findings" style="font-size:22px">—</div></div>
 <div class="card"><div class="card-title">Rules Ran</div>
  <div class="card-value" id="rules-ran" style="font-size:22px">—</div></div>
 <div class="card"><div class="card-title">Last Scan</div>
  <div class="card-value" id="scan-time" style="font-size:16px">—</div></div>
</div>

<!-- Action bar -->
<div style="display:flex;gap:12px;margin-bottom:20px;align-items:center">
 <button class="action-btn" id="btn-rescan" style="flex-direction:row;padding:10px 20px;gap:6px">
  <span class="action-glyph" style="font-size:16px">𓁢</span> Run Scan</button>
 <button class="action-btn" id="btn-clean-safe" style="flex-direction:row;padding:10px 20px;gap:6px;display:none">
  <span class="action-glyph" style="font-size:16px">🧹</span> Clean Safe Items</button>
 <span id="scan-status" style="font-size:11px;color:#666;flex:1;text-align:right"></span>
</div>

<!-- Category breakdown -->
<div id="categories"></div>

<!-- Findings table -->
<div id="findings-area"></div>

<!-- Empty state -->
<div id="empty-state" class="card">
 <div class="empty"><div class="empty-glyph">𓁢</div>No scan results yet. Click "Run Scan" to begin.</div>
</div>

<style>
.cat-header{display:flex;align-items:center;gap:12px;padding:14px 20px;cursor:pointer;
 border-bottom:1px solid rgba(200,169,81,.06);transition:background .15s}
.cat-header:hover{background:rgba(200,169,81,.04)}
.cat-chevron{transition:transform .2s;color:#555;font-size:12px}
.cat-chevron.open{transform:rotate(90deg)}
.cat-name{font-size:13px;font-weight:600;flex:1;color:#FAFAFA}
.cat-meta{font-size:11px;color:#888}
.finding-row{display:flex;align-items:center;gap:12px;padding:10px 20px 10px 44px;
 border-bottom:1px solid rgba(200,169,81,.04);font-size:12px;transition:background .15s}
.finding-row:hover{background:rgba(200,169,81,.03)}
.finding-sev{width:20px;text-align:center}
.finding-desc{flex:1;color:#ccc}
.finding-path{color:#666;font-family:monospace;font-size:11px;max-width:300px;overflow:hidden;
 text-overflow:ellipsis;white-space:nowrap;direction:rtl;text-align:left}
.finding-size{color:#C8A951;font-weight:600;min-width:70px;text-align:right}
.finding-action{min-width:60px;text-align:right}
.clean-btn{background:none;border:1px solid rgba(200,169,81,.2);color:#888;font-size:10px;
 padding:3px 8px;border-radius:4px;cursor:pointer;transition:all .2s}
.clean-btn:hover{border-color:#C8A951;color:#C8A951}
.clean-btn.done{border-color:#44FF88;color:#44FF88;cursor:default}
.clean-btn.err{border-color:#FF4444;color:#FF4444;cursor:default}
</style>

<script>
(function(){
'use strict';

const catIcons={general:'📁',vms:'🖥',dev:'🔧',ai:'🤖',ides:'💻',cloud:'☁️',storage:'💾'};
const catLabels={general:'General',vms:'Virtualization',dev:'Developer',ai:'AI & ML',
 ides:'IDEs & Editors',cloud:'Cloud & Infra',storage:'Storage'};
const sevIcon=s=>({safe:'🟢',caution:'🟡',warning:'🟠'}[s]||'⚪');
const fmtSize=b=>{if(b>=1073741824)return(b/1073741824).toFixed(1)+' GB';
 if(b>=1048576)return(b/1048576).toFixed(1)+' MB';if(b>=1024)return(b/1024).toFixed(1)+' KB';return b+' B'};
const ago=ts=>{if(!ts)return'—';const d=Date.now()-new Date(ts).getTime();
 if(d<60e3)return'just now';if(d<3600e3)return Math.floor(d/6e4)+'m ago';
 if(d<864e5)return Math.floor(d/36e5)+'h ago';return Math.floor(d/864e5)+'d ago'};

let scanData=null;
let openCats={};

function loadFindings(){
 fetch('/api/findings').then(r=>r.json()).then(function(data){
  if(data.error&&!data.findings?.length){
   document.getElementById('empty-state').style.display='';
   document.getElementById('categories').textContent='';
   document.getElementById('findings-area').textContent='';
   return;
  }
  scanData=data;
  document.getElementById('empty-state').style.display='none';
  renderSummary(data);
  renderCategories(data);
 }).catch(function(){});
}

function renderSummary(data){
 document.getElementById('total-waste').textContent=fmtSize(data.total_size||0);
 document.getElementById('total-findings').textContent=(data.findings||[]).length;
 document.getElementById('rules-ran').textContent=data.rules_ran||0;
 document.getElementById('scan-time').textContent=ago(data.timestamp);
 const cleanBtn=document.getElementById('btn-clean-safe');
 const safeCount=(data.findings||[]).filter(f=>f.severity==='safe').length;
 if(safeCount>0){cleanBtn.style.display='';cleanBtn.textContent='🧹 Clean '+safeCount+' Safe Items'}
 else{cleanBtn.style.display='none'}
}

function renderCategories(data){
 const cats={};const findings=data.findings||[];
 findings.forEach(function(f,i){f._idx=i;if(!cats[f.category])cats[f.category]={findings:[],size:0};
  cats[f.category].findings.push(f);cats[f.category].size+=f.size_bytes});

 const sorted=Object.entries(cats).sort(function(a,b){return b[1].size-a[1].size});
 const container=document.getElementById('categories');
 container.textContent='';

 sorted.forEach(function(pair){
  const cat=pair[0],info=pair[1];
  const card=document.createElement('div');card.className='card';card.style.cssText='padding:0;overflow:hidden;margin-bottom:12px';

  const header=document.createElement('div');header.className='cat-header';
  const chevron=document.createElement('span');chevron.className='cat-chevron';chevron.textContent='▸';
  const icon=document.createElement('span');icon.textContent=catIcons[cat]||'📦';icon.style.fontSize='18px';
  const name=document.createElement('span');name.className='cat-name';
  name.textContent=(catLabels[cat]||cat)+' ('+info.findings.length+')';
  const meta=document.createElement('span');meta.className='cat-meta';meta.textContent=fmtSize(info.size);
  header.appendChild(chevron);header.appendChild(icon);header.appendChild(name);header.appendChild(meta);

  const body=document.createElement('div');body.style.display=openCats[cat]?'':'none';

  info.findings.forEach(function(f){
   const row=document.createElement('div');row.className='finding-row';
   const sev=document.createElement('span');sev.className='finding-sev';sev.textContent=sevIcon(f.severity);
   const desc=document.createElement('span');desc.className='finding-desc';desc.textContent=f.description;
   const path=document.createElement('span');path.className='finding-path';path.textContent=f.path;path.title=f.path;
   const size=document.createElement('span');size.className='finding-size';size.textContent=f.size_human||fmtSize(f.size_bytes);
   const action=document.createElement('span');action.className='finding-action';
   if(f.severity==='safe'||f.severity==='caution'){
    const btn=document.createElement('button');btn.className='clean-btn';btn.textContent='Clean';
    btn.dataset.idx=f._idx;
    btn.addEventListener('click',function(e){e.stopPropagation();cleanFinding(btn,f._idx)});
    action.appendChild(btn);
   }
   row.appendChild(sev);row.appendChild(desc);row.appendChild(path);row.appendChild(size);row.appendChild(action);
   body.appendChild(row);
  });

  if(openCats[cat])chevron.classList.add('open');
  header.addEventListener('click',function(){
   const open=body.style.display==='none';
   body.style.display=open?'':'none';
   chevron.classList.toggle('open',open);
   openCats[cat]=open;
  });

  card.appendChild(header);card.appendChild(body);container.appendChild(card);
 });
}

function cleanFinding(btn,idx){
 if(btn.classList.contains('done')||btn.classList.contains('err'))return;
 btn.textContent='...';
 fetch('/api/clean',{method:'POST',headers:{'Content-Type':'application/json'},
  body:JSON.stringify({indices:[idx],dry_run:false})
 }).then(r=>r.json()).then(function(d){
  if(d.cleaned>0){btn.textContent='✓ '+d.freed_human;btn.classList.add('done')}
  else{btn.textContent='Skip';btn.classList.add('err')}
 }).catch(function(){btn.textContent='Err';btn.classList.add('err')});
}

/* Clean all safe items */
document.getElementById('btn-clean-safe').addEventListener('click',function(){
 if(!scanData)return;
 const safeIdx=[];
 (scanData.findings||[]).forEach(function(f,i){if(f.severity==='safe')safeIdx.push(i)});
 if(!safeIdx.length)return;
 const btn=this;btn.textContent='🧹 Cleaning...';btn.disabled=true;
 fetch('/api/clean',{method:'POST',headers:{'Content-Type':'application/json'},
  body:JSON.stringify({indices:safeIdx,dry_run:false})
 }).then(r=>r.json()).then(function(d){
  btn.textContent='✓ Freed '+d.freed_human;btn.style.borderColor='#44FF88';btn.style.color='#44FF88';
  setTimeout(function(){loadFindings()},1500);
 }).catch(function(){btn.textContent='Error';btn.style.color='#FF4444';btn.disabled=false});
});

/* Re-scan */
document.getElementById('btn-rescan').addEventListener('click',function(){
 const btn=this;const status=document.getElementById('scan-status');
 btn.disabled=true;btn.textContent='𓁢 Scanning...';status.textContent='';
 fetch('/api/run?cmd=scan',{method:'POST'}).then(function(r){
  if(!r.ok)return r.json().then(function(e){throw new Error(e.error)});
  status.textContent='Scan running...';status.style.color='#C8A951';

  /* Poll for completion */
  const poll=setInterval(function(){
   fetch('/api/run/status').then(r=>r.json()).then(function(d){
    if(!d.running){
     clearInterval(poll);
     btn.disabled=false;btn.textContent='𓁢 Run Scan';
     status.textContent='Scan complete';status.style.color='#44FF88';
     setTimeout(function(){status.textContent=''},3000);
     loadFindings();
    }
   }).catch(function(){});
  },1000);
 }).catch(function(e){btn.disabled=false;btn.textContent='𓁢 Run Scan';
  status.textContent=e.message;status.style.color='#FF4444'});
});

loadFindings();
})();
</script>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, pageShell("Scan", "/scan", body))
}

// ── Ghosts Page ─────────────────────────────────────────────────────────

func (s *Server) handleGhosts(w http.ResponseWriter, r *http.Request) {
	body := `
<h1 class="page-title">𓂓 Ghost Detection</h1>
<div style="display:flex;gap:12px;margin-bottom:20px;align-items:center">
 <button class="action-btn" id="btn-hunt" style="flex-direction:row;padding:10px 20px;gap:6px">
  <span class="action-glyph" style="font-size:16px">𓂓</span> Run Ghost Hunt</button>
 <span id="hunt-status" style="font-size:11px;color:#666;flex:1;text-align:right"></span>
</div>

<div class="grid grid-3" style="margin-bottom:20px">
 <div class="card"><div class="card-title">Ghosts Found</div>
  <div class="card-value" id="ghost-count" style="font-size:22px">—</div></div>
 <div class="card"><div class="card-title">Total Waste</div>
  <div class="card-value" id="ghost-waste" style="font-size:22px">—</div></div>
 <div class="card"><div class="card-title">Residual Files</div>
  <div class="card-value" id="ghost-files" style="font-size:22px">—</div></div>
</div>

<div id="ghosts-list"></div>
<div id="empty-state" class="card">
 <div class="empty"><div class="empty-glyph">𓂓</div>Click "Run Ghost Hunt" to scan for dead app remnants</div>
</div>

<script>
(function(){
'use strict';
const fmtSize=b=>{if(b>=1073741824)return(b/1073741824).toFixed(1)+' GB';
 if(b>=1048576)return(b/1048576).toFixed(1)+' MB';if(b>=1024)return(b/1024).toFixed(1)+' KB';return b+' B'};

function loadGhosts(){
 const status=document.getElementById('hunt-status');
 status.textContent='Scanning...';status.style.color='#C8A951';
 document.getElementById('btn-hunt').disabled=true;
 fetch('/api/ghosts').then(r=>r.json()).then(function(ghosts){
  status.textContent='';document.getElementById('btn-hunt').disabled=false;
  if(!ghosts.length){document.getElementById('empty-state').style.display='';return}
  document.getElementById('empty-state').style.display='none';
  let totalSize=0,totalFiles=0;
  ghosts.forEach(function(g){totalSize+=g.total_size;totalFiles+=g.total_files});
  document.getElementById('ghost-count').textContent=ghosts.length;
  document.getElementById('ghost-waste').textContent=fmtSize(totalSize);
  document.getElementById('ghost-files').textContent=totalFiles;
  renderGhosts(ghosts);
 }).catch(function(e){status.textContent='Error: '+e.message;status.style.color='#FF4444';
  document.getElementById('btn-hunt').disabled=false});
}

function renderGhosts(ghosts){
 const container=document.getElementById('ghosts-list');container.textContent='';
 ghosts.sort(function(a,b){return b.total_size-a.total_size});
 ghosts.forEach(function(g){
  const card=document.createElement('div');card.className='card';card.style.cssText='padding:0;overflow:hidden;margin-bottom:12px';
  const header=document.createElement('div');header.className='cat-header';
  const chevron=document.createElement('span');chevron.className='cat-chevron';chevron.textContent='▸';
  const name=document.createElement('span');name.className='cat-name';
  name.textContent='👻 '+g.app_name+(g.in_launch_services?' (still in Launch Services)':'');
  const meta=document.createElement('span');meta.className='cat-meta';
  meta.textContent=g.total_files+' files · '+fmtSize(g.total_size);
  const cleanBtn=document.createElement('button');cleanBtn.className='clean-btn';
  cleanBtn.textContent='Clean';cleanBtn.style.marginLeft='12px';
  cleanBtn.addEventListener('click',function(e){e.stopPropagation();cleanGhost(cleanBtn,g.app_name)});
  header.appendChild(chevron);header.appendChild(name);header.appendChild(meta);header.appendChild(cleanBtn);

  const body=document.createElement('div');body.style.display='none';
  g.residuals.forEach(function(r){
   const row=document.createElement('div');row.className='finding-row';
   const type=document.createElement('span');type.className='finding-desc';type.style.color='#999';
   type.textContent=r.type;type.style.minWidth='140px';type.style.flex='none';
   const path=document.createElement('span');path.className='finding-path';path.textContent=r.path;path.title=r.path;path.style.flex='1';
   const size=document.createElement('span');size.className='finding-size';size.textContent=fmtSize(r.size_bytes);
   row.appendChild(type);row.appendChild(path);row.appendChild(size);body.appendChild(row);
  });

  header.addEventListener('click',function(){
   const open=body.style.display==='none';body.style.display=open?'':'none';
   chevron.classList.toggle('open',open)});
  card.appendChild(header);card.appendChild(body);container.appendChild(card);
 });
}

function cleanGhost(btn,appName){
 if(btn.classList.contains('done'))return;btn.textContent='...';
 fetch('/api/ghosts/clean',{method:'POST',headers:{'Content-Type':'application/json'},
  body:JSON.stringify({app_name:appName,dry_run:false})
 }).then(r=>r.json()).then(function(d){
  btn.textContent='✓ '+d.freed_human;btn.classList.add('done');btn.style.borderColor='#44FF88';btn.style.color='#44FF88';
 }).catch(function(){btn.textContent='Err';btn.style.color='#FF4444'});
}

document.getElementById('btn-hunt').addEventListener('click',loadGhosts);
})();
</script>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, pageShell("Ghosts", "/ghosts", body))
}

// ── Horus Page ──────────────────────────────────────────────────────────

func (s *Server) handleHorus(w http.ResponseWriter, r *http.Request) {
	body := `
<h1 class="page-title">𓂀 Horus — Code Graph</h1>

<div style="display:flex;gap:12px;margin-bottom:20px;align-items:center">
 <input type="text" class="search-box" id="horus-search" placeholder="Search symbols (e.g. Server, *Handler, func)..." style="flex:1">
 <button class="action-btn" id="btn-horus-scan" style="flex-direction:row;padding:10px 20px;gap:6px;white-space:nowrap">
  𓂀 Scan Project</button>
</div>

<div class="grid grid-4" style="margin-bottom:20px" id="horus-stats">
 <div class="card"><div class="card-title">Files</div><div class="card-value" id="h-files" style="font-size:22px">—</div></div>
 <div class="card"><div class="card-title">Packages</div><div class="card-value" id="h-pkgs" style="font-size:22px">—</div></div>
 <div class="card"><div class="card-title">Types</div><div class="card-value" id="h-types" style="font-size:22px">—</div></div>
 <div class="card"><div class="card-title">Functions</div><div class="card-value" id="h-funcs" style="font-size:22px">—</div></div>
</div>

<div id="search-results" style="display:none">
 <h2 class="page-subtitle">Search Results</h2>
 <div class="card" style="padding:0;overflow:hidden">
  <table class="tbl"><thead><tr><th>Kind</th><th>Name</th><th>File</th><th>Line</th></tr></thead>
   <tbody id="results-body"></tbody></table>
 </div>
</div>

<div id="empty-state" class="card">
 <div class="empty"><div class="empty-glyph">𓂀</div>Click "Scan Project" to analyze the codebase, then search for symbols</div>
</div>

<script>
(function(){
'use strict';
const kindBadge={func:'badge-info',method:'badge-info',type:'badge-warning',struct:'badge-warning',
 interface:'badge-success','const':'badge-info','var':'badge-info',field:'badge-info'};

document.getElementById('btn-horus-scan').addEventListener('click',function(){
 const btn=this;btn.disabled=true;btn.textContent='Scanning...';
 fetch('/api/horus/scan?path=.').then(r=>r.json()).then(function(g){
  btn.disabled=false;btn.textContent='𓂀 Scan Project';
  document.getElementById('empty-state').style.display='none';
  const s=g.stats||g.Stats||{};
  document.getElementById('h-files').textContent=s.files||s.Files||0;
  document.getElementById('h-pkgs').textContent=s.packages||s.Packages||0;
  document.getElementById('h-types').textContent=s.types||s.Types||0;
  document.getElementById('h-funcs').textContent=(s.functions||s.Functions||0)+(s.methods||s.Methods||0);
 }).catch(function(){btn.disabled=false;btn.textContent='Error'});
});

let searchTimer=null;
document.getElementById('horus-search').addEventListener('input',function(){
 clearTimeout(searchTimer);const q=this.value.trim();
 if(!q){document.getElementById('search-results').style.display='none';return}
 searchTimer=setTimeout(function(){
  fetch('/api/horus/query?path=.&filter='+encodeURIComponent('*'+q+'*')).then(r=>r.json()).then(function(symbols){
   document.getElementById('search-results').style.display='';
   document.getElementById('empty-state').style.display='none';
   const tb=document.getElementById('results-body');tb.textContent='';
   if(!symbols||!symbols.length){const tr=document.createElement('tr');const td=document.createElement('td');
    td.colSpan=4;td.className='empty';td.textContent='No symbols match';tr.appendChild(td);tb.appendChild(tr);return}
   symbols.slice(0,50).forEach(function(s){const tr=document.createElement('tr');
    const k=document.createElement('td');const badge=document.createElement('span');
    badge.className='badge '+(kindBadge[s.kind]||'badge-info');badge.textContent=s.kind;k.appendChild(badge);
    const n=document.createElement('td');n.style.fontWeight='600';n.textContent=(s.parent?s.parent+'.':'')+s.name;
    const f=document.createElement('td');f.className='finding-path';f.textContent=s.file;f.title=s.file;
    const l=document.createElement('td');l.style.cssText='color:#666;font-size:11px';l.textContent=s.line;
    tr.appendChild(k);tr.appendChild(n);tr.appendChild(f);tr.appendChild(l);tb.appendChild(tr)});
  }).catch(function(){});
 },300);
});
})();
</script>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, pageShell("Horus", "/horus", body))
}

// ── Vault Page ──────────────────────────────────────────────────────────

func (s *Server) handleVault(w http.ResponseWriter, r *http.Request) {
	body := `
<h1 class="page-title">🏛 Vault — Context Sandbox</h1>

<div style="display:flex;gap:12px;margin-bottom:20px;align-items:center">
 <input type="text" class="search-box" id="vault-search" placeholder="Full-text search across all sandboxed content..." style="flex:1">
 <button class="action-btn" id="btn-prune" style="flex-direction:row;padding:10px 16px;gap:6px;white-space:nowrap;font-size:12px">
  🧹 Prune Old</button>
</div>

<div class="grid grid-4" style="margin-bottom:20px">
 <div class="card"><div class="card-title">Entries</div><div class="card-value" id="v-entries" style="font-size:22px">—</div></div>
 <div class="card"><div class="card-title">Total Size</div><div class="card-value" id="v-bytes" style="font-size:22px">—</div></div>
 <div class="card"><div class="card-title">Tokens</div><div class="card-value" id="v-tokens" style="font-size:22px">—</div></div>
 <div class="card"><div class="card-title">Tags</div><div class="card-value" id="v-tags" style="font-size:22px">—</div></div>
</div>

<div id="search-results" style="display:none">
 <h2 class="page-subtitle" id="results-label">Results</h2>
 <div id="results-list"></div>
</div>

<div id="empty-state" class="card">
 <div class="empty"><div class="empty-glyph">🏛</div>Type a query to search, or use <code>sirsi vault store</code> to add content</div>
</div>

<script>
(function(){
'use strict';
const fmtSize=b=>{if(b>=1073741824)return(b/1073741824).toFixed(1)+' GB';
 if(b>=1048576)return(b/1048576).toFixed(1)+' MB';if(b>=1024)return(b/1024).toFixed(1)+' KB';return b+' B'};
const fmtNum=n=>{if(n>=1e6)return(n/1e6).toFixed(1)+'M';if(n>=1e3)return(n/1e3).toFixed(1)+'K';return n};

function loadStats(){
 fetch('/api/vault/stats').then(r=>r.json()).then(function(s){
  document.getElementById('v-entries').textContent=s.totalEntries||0;
  document.getElementById('v-bytes').textContent=fmtSize(s.totalBytes||0);
  document.getElementById('v-tokens').textContent=fmtNum(s.totalTokens||0);
  const tags=s.tagCounts||{};
  document.getElementById('v-tags').textContent=Object.keys(tags).length;
  if(s.totalEntries>0)document.getElementById('empty-state').style.display='none';
 }).catch(function(){});
}

let searchTimer=null;
document.getElementById('vault-search').addEventListener('input',function(){
 clearTimeout(searchTimer);const q=this.value.trim();
 if(!q){document.getElementById('search-results').style.display='none';
  if(document.getElementById('v-entries').textContent!=='0')document.getElementById('empty-state').style.display='none';
  return}
 searchTimer=setTimeout(function(){
  fetch('/api/vault/search?q='+encodeURIComponent(q)+'&limit=20').then(r=>r.json()).then(function(res){
   document.getElementById('search-results').style.display='';
   document.getElementById('empty-state').style.display='none';
   document.getElementById('results-label').textContent='Results ('+res.totalHits+' hits)';
   const list=document.getElementById('results-list');list.textContent='';
   if(!res.entries||!res.entries.length){
    const card=document.createElement('div');card.className='card';
    card.textContent='No results for "'+q+'"';card.style.color='#666';
    list.appendChild(card);return}
   res.entries.forEach(function(e){
    const card=document.createElement('div');card.className='card';card.style.marginBottom='8px';
    const header=document.createElement('div');header.style.cssText='display:flex;gap:12px;align-items:center;margin-bottom:8px';
    const src=document.createElement('span');src.style.cssText='font-weight:600;color:#C8A951';src.textContent=e.source;
    const tag=document.createElement('span');tag.className='badge badge-info';tag.textContent=e.tag;
    const time=document.createElement('span');time.style.cssText='color:#666;font-size:11px;margin-left:auto';time.textContent=e.createdAt;
    header.appendChild(src);header.appendChild(tag);header.appendChild(time);
    const snippet=document.createElement('pre');
    snippet.style.cssText='font-size:12px;color:#aaa;white-space:pre-wrap;word-break:break-all;margin:0;font-family:monospace;max-height:120px;overflow:hidden';
    snippet.textContent=e.snippet||e.content?.substring(0,200)||'';
    card.appendChild(header);card.appendChild(snippet);list.appendChild(card)});
  }).catch(function(){});
 },300);
});

document.getElementById('btn-prune').addEventListener('click',function(){
 const btn=this;btn.disabled=true;btn.textContent='Pruning...';
 fetch('/api/vault/prune?older_than=720h',{method:'POST'}).then(r=>r.json()).then(function(d){
  btn.disabled=false;btn.textContent='✓ Removed '+d.removed;
  setTimeout(function(){btn.textContent='🧹 Prune Old';loadStats()},2000);
 }).catch(function(){btn.disabled=false;btn.textContent='Error'});
});

loadStats();
})();
</script>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, pageShell("Vault", "/vault", body))
}

// ── Helpers ─────────────────────────────────────────────────────────────

// readSteleByType reads the Stele JSONL file and returns entries matching any of the given types.
// Returns newest first, up to 100 entries. Read-only — does not advance any consumer offset.
func (s *Server) readSteleByType(types ...string) []stele.Entry {
	path := s.stelePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	typeSet := make(map[string]bool, len(types))
	for _, t := range types {
		typeSet[t] = true
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	var entries []stele.Entry
	for i := len(lines) - 1; i >= 0 && len(entries) < 100; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		var e stele.Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		if typeSet[e.Type] {
			entries = append(entries, e)
		}
	}
	return entries
}
