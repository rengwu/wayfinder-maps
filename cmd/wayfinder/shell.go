package main

// shellHTML is the static canvas app. It fetches /graph.json at load and renders
// the star-map: a deterministic rank-biased force layout, a pan/zoom camera over
// a parallax starfield, glowing status-coloured stars, curved edges, an RTS HUD,
// and a click-to-open detail panel. Vanilla JS, no dependencies, no build step.
//
// The JS below must contain no backticks — it lives inside a Go raw string.
const shellHTML = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>wayfinder</title>
<style>
  *{box-sizing:border-box}
  html,body{margin:0;height:100%;overflow:hidden;background:#05070d;
    font:14px/1.5 ui-sans-serif,-apple-system,Segoe UI,Roboto,sans-serif;color:#e6e9ef}
  #sky{position:fixed;inset:0;display:block;cursor:grab}
  #sky.drag{cursor:grabbing}
  #hud{position:fixed;top:16px;left:16px;max-width:340px;z-index:5;
    background:rgba(14,18,26,.72);border:1px solid #262b36;border-radius:12px;
    padding:14px 16px;backdrop-filter:blur(8px)}
  #hud h1{font-size:15px;margin:0 0 8px;line-height:1.3}
  #hud .counts{display:flex;gap:6px;flex-wrap:wrap;margin-bottom:10px}
  #hud .c{font-size:11px;padding:2px 8px;border-radius:999px;border:1px solid #262b36;color:#8b93a3}
  .c.resolved{color:#b9c9e0;border-color:#42506a}
  .c.claimed{color:#e0a44b;border-color:#7a5a2a}
  .c.frontier{color:#ffd873;border-color:#8a6a20}
  .c.blocked{color:#c07a7a;border-color:#6a3b3b}
  .c.oos{color:#8a8496;border-color:#4a4550}
  #hud .bar{height:6px;background:#20242e;border-radius:999px;overflow:hidden;margin-bottom:10px}
  #hud .bar>span{display:block;height:100%;background:linear-gradient(90deg,#6d86ad,#b9c9e0)}
  #hud .dest{font-size:12px;color:#8b93a3;line-height:1.45;
    display:-webkit-box;-webkit-line-clamp:3;-webkit-box-orient:vertical;overflow:hidden}
  #hud .legend{display:flex;gap:12px;flex-wrap:wrap;margin-top:10px;font-size:11px;color:#8b93a3}
  #hud .legend b{font-weight:400}
  .dot{display:inline-block;width:8px;height:8px;border-radius:50%;margin-right:4px;vertical-align:middle}
  #panel{position:fixed;top:0;right:0;height:100%;width:min(420px,86vw);z-index:6;
    background:rgba(12,15,21,.94);border-left:1px solid #262b36;backdrop-filter:blur(10px);
    transform:translateX(102%);transition:transform .22s cubic-bezier(.2,.7,.2,1);
    display:flex;flex-direction:column;padding:20px 22px}
  #panel.open{transform:translateX(0)}
  #panel .x{position:absolute;top:14px;right:16px;cursor:pointer;color:#8b93a3;font-size:20px;line-height:1}
  #panel .x:hover{color:#e6e9ef}
  #panel h2{font-size:17px;margin:2px 0 10px;padding-right:24px}
  #panel h2 .num{color:#8b93a3;font-variant-numeric:tabular-nums;font-weight:700;margin-right:6px}
  #panel .meta{display:flex;gap:6px;flex-wrap:wrap;margin-bottom:14px}
  #panel pre{flex:1;overflow:auto;margin:0;white-space:pre-wrap;word-wrap:break-word;
    font:12.5px/1.6 ui-monospace,SFMono-Regular,Menlo,monospace;color:#cfd4de;
    border-top:1px solid #262b36;padding-top:14px}
  #hint{position:fixed;bottom:14px;left:16px;z-index:5;font-size:11px;color:#4b5261}
</style>
</head>
<body>
<canvas id="sky"></canvas>
<div id="hud"></div>
<div id="panel"><span class="x">&times;</span><h2></h2><div class="meta"></div><pre></pre></div>
<div id="hint">drag to pan &middot; scroll to zoom &middot; click a star</div>
<script>
"use strict";
var canvas=document.getElementById("sky"), ctx=canvas.getContext("2d");
var dpr=Math.max(1,window.devicePixelRatio||1);
var cam={x:0,y:0,s:1};
var graph=null, nodes=[], edges=[], byNum={}, selected=null;
var clock=0, T0=(window.performance&&performance.now?performance.now():Date.now()), camTarget=null;

var COL={
  resolved:{core:"#b9c9e0",glow:"#5d76ad",r:6,gr:26},
  frontier:{core:"#ffd873",glow:"#ffb020",r:9,gr:54},
  claimed:{core:"#f0c078",glow:"#e0a44b",r:8,gr:40},
  blocked:{core:"#c07a7a",glow:"#7a3b3b",r:5,gr:22},
  out_of_scope:{core:"#7d7789",glow:"#4a4550",r:5,gr:20}
};
var LABELCOL={resolved:"#9fb2cc",frontier:"#ffe6a0",claimed:"#e8c288",blocked:"#c49a9a",out_of_scope:"#8a8496"};

function col(n){return COL[n.status]||COL.blocked;}
function mod(a,n){return ((a%n)+n)%n;}
function clamp(v,a,b){return v<a?a:(v>b?b:v);}
function pad2(n){return (n<10?"0":"")+n;}
function mulberry32(a){return function(){a|=0;a=a+0x6D2B79F5|0;var t=Math.imul(a^a>>>15,1|a);t=t+Math.imul(t^t>>>7,61|t)^t;return ((t^t>>>14)>>>0)/4294967296;};}
function hexA(hex,al){var h=hex.replace("#","");var r=parseInt(h.substr(0,2),16),g=parseInt(h.substr(2,2),16),b=parseInt(h.substr(4,2),16);return "rgba("+r+","+g+","+b+","+al+")";}
// w2s maps a node's DISPLAY position (base + idle bob, set each frame) to screen.
function w2s(n){var x=(n._x!=null?n._x:n.x), y=(n._y!=null?n._y:n.y);return {x:x*cam.s+cam.x,y:y*cam.s+cam.y};}

// --- deterministic rank-biased force layout --------------------------------
function ringR(rank){return 130+rank*165;}
function layout(){
  var rnd=mulberry32(1337);
  nodes.forEach(function(n){
    var ang=rnd()*6.2831853;
    var jit=(rnd()-0.5)*70;
    var R=ringR(n.rank)+jit;
    n.x=Math.cos(ang)*R; n.y=Math.sin(ang)*R;
  });
  var REP=9000, SPRING=0.02, REST=150, RADIAL=0.05;
  for(var it=0; it<420; it++){
    for(var i=0;i<nodes.length;i++){
      var a=nodes[i];
      for(var j=i+1;j<nodes.length;j++){
        var b=nodes[j];
        var dx=a.x-b.x, dy=a.y-b.y; var d2=dx*dx+dy*dy||0.01; var d=Math.sqrt(d2);
        var f=REP/d2; var ux=dx/d, uy=dy/d;
        a.x+=ux*f; a.y+=uy*f; b.x-=ux*f; b.y-=uy*f;
      }
    }
    edges.forEach(function(e){
      var a=byNum[e.from], b=byNum[e.to]; if(!a||!b)return;
      var dx=b.x-a.x, dy=b.y-a.y; var d=Math.hypot(dx,dy)||0.01;
      var f=(d-REST)*SPRING; var ux=dx/d, uy=dy/d;
      a.x+=ux*f; a.y+=uy*f; b.x-=ux*f; b.y-=uy*f;
    });
    nodes.forEach(function(n){
      var d=Math.hypot(n.x,n.y)||0.01; var target=ringR(n.rank);
      var f=(target-d)*RADIAL; n.x+=(n.x/d)*f; n.y+=(n.y/d)*f;
    });
  }
}

function fitCamera(){
  if(!nodes.length){cam={x:innerWidth/2,y:innerHeight/2,s:1};return;}
  var minx=1e9,miny=1e9,maxx=-1e9,maxy=-1e9;
  nodes.forEach(function(n){minx=Math.min(minx,n.x);miny=Math.min(miny,n.y);maxx=Math.max(maxx,n.x);maxy=Math.max(maxy,n.y);});
  var pad=90; minx-=pad;miny-=pad;maxx+=pad;maxy+=pad;
  var spanx=maxx-minx||1, spany=maxy-miny||1;
  var s=Math.min(innerWidth/spanx, innerHeight/spany); s=clamp(s,0.15,1.4);
  cam.s=s;
  var cx=(minx+maxx)/2, cy=(miny+maxy)/2;
  cam.x=innerWidth/2-cx*s; cam.y=innerHeight/2-cy*s;
}

// --- parallax starfield ----------------------------------------------------
var starLayers=[];
function initStars(){
  starLayers=[];
  var specs=[{f:0.15,n:140,sz:0.7,a:0.45},{f:0.30,n:80,sz:1.1,a:0.65},{f:0.50,n:34,sz:1.7,a:0.9}];
  var rnd=mulberry32(9001);
  specs.forEach(function(sp){
    var arr=[];
    for(var i=0;i<sp.n;i++)arr.push({x:rnd(),y:rnd(),t:rnd()});
    starLayers.push({f:sp.f,sz:sp.sz,a:sp.a,stars:arr});
  });
}
function drawStars(W,H){
  starLayers.forEach(function(L){
    ctx.fillStyle="rgba(255,255,255,"+L.a+")";
    for(var i=0;i<L.stars.length;i++){
      var s=L.stars[i];
      var x=mod(s.x*W+cam.x*L.f, W), y=mod(s.y*H+cam.y*L.f, H);
      var tw=0.65+0.35*Math.sin(s.t*6.2831853);
      ctx.globalAlpha=L.a*tw;
      ctx.fillRect(x,y,L.sz,L.sz);
    }
  });
  ctx.globalAlpha=1;
}

// --- draw ------------------------------------------------------------------
// drawEdge draws a curved line from the blocker (from) to the dependent (to),
// with an arrowhead landing on the dependent so the blocking direction reads at
// a glance: from --> to means "from unblocks to".
function drawEdge(e){
  var a=byNum[e.from], b=byNum[e.to]; if(!a||!b)return;
  var ax=a._x, ay=a._y, bx=b._x, by=b._y;
  var mx=(ax+bx)/2, my=(ay+by)/2, dx=bx-ax, dy=by-ay, len=Math.hypot(dx,dy)||1;
  var nx=-dy/len, ny=dx/len, bow=Math.min(46,len*0.13);
  var cx=mx+nx*bow, cy=my+ny*bow;
  ctx.beginPath(); ctx.moveTo(ax,ay); ctx.quadraticCurveTo(cx,cy,bx,by);
  if(e.satisfied){ctx.strokeStyle="rgba(150,180,155,0.5)";ctx.lineWidth=1.6;ctx.setLineDash([]);}
  else{ctx.strokeStyle="rgba(120,132,152,0.22)";ctx.lineWidth=1.2;ctx.setLineDash([4,6]);}
  ctx.stroke(); ctx.setLineDash([]);
  // Flow particles: on a SATISFIED edge, motes drift blocker->dependent, reading
  // as energy having unlocked the next step. Staggered phases, dim at the ends.
  if(e.satisfied){
    for(var k=0;k<3;k++){
      var u=mod(clock*0.14 + k/3 + (e.from*0.13+e.to*0.07), 1);
      var m=1-u, fx=m*m*ax+2*m*u*cx+u*u*bx, fy=m*m*ay+2*m*u*cy+u*u*by;
      ctx.fillStyle="rgba(190,225,200,"+(0.16+0.44*Math.sin(u*Math.PI))+")";
      ctx.beginPath(); ctx.arc(fx,fy,1.7,0,6.2831853); ctx.fill();
    }
  }
  // Arrowhead at the curve MIDPOINT, pointing blocker->dependent, so direction
  // reads without crowding either vertex. For a quadratic Bezier the midpoint is
  // B(0.5)=0.25a+0.5c+0.25b and its tangent is simply b-a. The triangle is
  // centred on that point (tip half a length ahead, base half behind).
  var midx=0.25*ax+0.5*cx+0.25*bx, midy=0.25*ay+0.5*cy+0.25*by;
  var adx=bx-ax, ady=by-ay, al=Math.hypot(adx,ady)||1, ux=adx/al, uy=ady/al;
  var ah=7, aw=3.8, px=-uy, py=ux;
  var tipx=midx+ux*(ah*0.5), tipy=midy+uy*(ah*0.5);
  ctx.beginPath();
  ctx.moveTo(tipx,tipy);
  ctx.lineTo(tipx-ux*ah+px*aw, tipy-uy*ah+py*aw);
  ctx.lineTo(tipx-ux*ah-px*aw, tipy-uy*ah-py*aw);
  ctx.closePath();
  // Opaque fill: a translucent arrow over the translucent line (or over another
  // arrow near a hub) compounds alpha and reads as two stacked shapes. The
  // satisfied/unsatisfied distinction rides on hue, not transparency.
  ctx.fillStyle=e.satisfied?"#aecdb6":"#6f7889";
  ctx.fill();
}
function drawNode(n){
  var c=col(n), x=n._x, y=n._y;
  // Frontier stars breathe: glow and ring pulse gently so the eye is drawn to
  // what is takeable now.
  var isF=n.status==="frontier";
  var beat=0.5+0.5*Math.sin(clock*2.8);
  var pulse=isF?(0.8+0.2*beat):1;
  var gr=isF?c.gr*(0.92+0.16*beat):c.gr;
  var g=ctx.createRadialGradient(x,y,0,x,y,gr);
  g.addColorStop(0,hexA(c.glow,0.85*pulse)); g.addColorStop(0.4,hexA(c.glow,0.22*pulse)); g.addColorStop(1,hexA(c.glow,0));
  ctx.fillStyle=g; ctx.beginPath(); ctx.arc(x,y,gr,0,6.2831853); ctx.fill();
  ctx.fillStyle=c.core; ctx.beginPath(); ctx.arc(x,y,c.r,0,6.2831853); ctx.fill();
  if(isF){ctx.strokeStyle=hexA("#ffd873",0.4+0.3*beat);ctx.lineWidth=1.5;ctx.beginPath();ctx.arc(x,y,c.r+6+1.5*beat,0,6.2831853);ctx.stroke();}
  if(selected===n){ctx.strokeStyle="rgba(255,255,255,0.85)";ctx.lineWidth=1.5;ctx.beginPath();ctx.arc(x,y,c.r+11,0,6.2831853);ctx.stroke();}
}
function drawLabels(){
  if(cam.s<0.22)return;               // very far out: no text at all
  var numOnly=cam.s<0.42;             // far out: ticket number only
  var fs=clamp(11*Math.pow(cam.s,0.3),8,13); // subtle shrink as you zoom out
  ctx.textAlign="center"; ctx.font=fs.toFixed(1)+"px ui-sans-serif,system-ui,sans-serif";
  ctx.shadowColor="rgba(0,0,0,0.85)"; ctx.shadowBlur=4;
  nodes.forEach(function(n){
    var s=w2s(n); var c=col(n);
    ctx.fillStyle=LABELCOL[n.status]||"#c8ccd6";
    var label=pad2(n.num);
    if(!numOnly){var t=n.title.length>30?n.title.slice(0,29)+"…":n.title; label+="  "+t;}
    ctx.fillText(label, s.x, s.y+c.r*cam.s+fs+3);
  });
  ctx.shadowBlur=0;
}
function render(){
  var W=innerWidth,H=innerHeight;
  clock=((window.performance&&performance.now?performance.now():Date.now())-T0)/1000;
  // Idle bob: a tiny per-node drift, phased by ticket number. Edges, labels and
  // hit-testing all read _x/_y, so nothing detaches from a bobbing star.
  for(var i=0;i<nodes.length;i++){var n=nodes[i], ph=n.num*1.7;
    n._x=n.x+Math.sin(clock*0.7+ph)*2.4; n._y=n.y+Math.cos(clock*0.55+ph)*2.4;}
  // Camera easing toward a selected star; cancelled by any manual pan/zoom.
  if(camTarget){cam.x+=(camTarget.x-cam.x)*0.16; cam.y+=(camTarget.y-cam.y)*0.16;
    if(Math.abs(camTarget.x-cam.x)<0.4&&Math.abs(camTarget.y-cam.y)<0.4){cam.x=camTarget.x;cam.y=camTarget.y;camTarget=null;}}
  ctx.setTransform(dpr,0,0,dpr,0,0);
  ctx.fillStyle="#05070d"; ctx.fillRect(0,0,W,H);
  drawStars(W,H);
  ctx.save(); ctx.translate(cam.x,cam.y); ctx.scale(cam.s,cam.s);
  edges.forEach(drawEdge); nodes.forEach(drawNode);
  ctx.restore();
  drawLabels();
  requestAnimationFrame(render);
}

// --- hud & panel -----------------------------------------------------------
function el(tag,cls,txt){var e=document.createElement(tag);if(cls)e.className=cls;if(txt!=null)e.textContent=txt;return e;}
function buildHud(){
  var hud=document.getElementById("hud"); hud.innerHTML="";
  hud.appendChild(el("h1",null,graph.name));
  var c=graph.counts, counts=el("div","counts");
  function chip(cls,txt){counts.appendChild(el("span","c "+cls,txt));}
  chip("resolved",c.resolved+" resolved"); chip("frontier",graph.counts.total? frontierCount()+" frontier":"0 frontier");
  if(c.claimed)chip("claimed",c.claimed+" claimed");
  chip("blocked",blockedCount()+" blocked");
  if(c.outOfScope)chip("oos",c.outOfScope+" out of scope");
  hud.appendChild(counts);
  var bar=el("div","bar"); var span=document.createElement("span");
  span.style.width=(c.total? Math.round(c.resolved*100/c.total):0)+"%"; bar.appendChild(span); hud.appendChild(bar);
  if(graph.destination)hud.appendChild(el("div","dest",graph.destination));
  var lg=el("div","legend");
  [["frontier","frontier"],["claimed","claimed"],["blocked","blocked"],["resolved","resolved"]].forEach(function(p){
    var b=el("b"); var d=el("span","dot"); d.style.background=COL[p[0]].core; b.appendChild(d); b.appendChild(document.createTextNode(p[1])); lg.appendChild(b);
  });
  hud.appendChild(lg);
}
function frontierCount(){var n=0;nodes.forEach(function(x){if(x.status==="frontier")n++;});return n;}
function blockedCount(){var n=0;nodes.forEach(function(x){if(x.status==="blocked")n++;});return n;}

function openPanel(n){
  selected=n;
  var p=document.getElementById("panel");
  // Ease the camera so the star sits centred in the space left of the panel.
  var pw=p.offsetWidth||0, vx=(innerWidth-pw)/2, vy=innerHeight/2;
  camTarget={x:vx-n.x*cam.s, y:vy-n.y*cam.s};
  var h=p.querySelector("h2"); h.innerHTML=""; h.appendChild(el("span","num",pad2(n.num))); h.appendChild(document.createTextNode(n.title));
  var meta=p.querySelector(".meta"); meta.innerHTML="";
  meta.appendChild(el("span","c "+(n.status==="out_of_scope"?"oos":n.status), n.status.replace(/_/g," ")));
  if(n.type)meta.appendChild(el("span","c blocked",n.type));
  (n.blockers||[]).forEach(function(b){
    var dep=byNum[b], ok=dep&&dep.status==="resolved";
    meta.appendChild(el("span","c "+(ok?"resolved":"blocked"),"blocked by "+pad2(b)));
  });
  if(n.undermined)meta.appendChild(el("span","c blocked","undermined"));
  p.querySelector("pre").textContent=n.body||"(no body)";
  p.classList.add("open");
}
function closePanel(){selected=null;document.getElementById("panel").classList.remove("open");}
document.querySelector("#panel .x").onclick=closePanel;

// --- camera interaction ----------------------------------------------------
var down=false, moved=0, last={x:0,y:0};
canvas.addEventListener("mousedown",function(ev){down=true;moved=0;camTarget=null;last={x:ev.clientX,y:ev.clientY};canvas.classList.add("drag");});
window.addEventListener("mousemove",function(ev){
  if(!down)return; var dx=ev.clientX-last.x, dy=ev.clientY-last.y;
  cam.x+=dx; cam.y+=dy; moved+=Math.abs(dx)+Math.abs(dy); last={x:ev.clientX,y:ev.clientY};
});
window.addEventListener("mouseup",function(ev){
  canvas.classList.remove("drag"); if(!down)return; down=false;
  if(moved<5)hitTest(ev.clientX,ev.clientY);
});
canvas.addEventListener("wheel",function(ev){
  ev.preventDefault(); camTarget=null;
  var f=Math.exp(-ev.deltaY*0.0012); var ns=clamp(cam.s*f,0.15,4);
  var wx=(ev.clientX-cam.x)/cam.s, wy=(ev.clientY-cam.y)/cam.s;
  cam.s=ns; cam.x=ev.clientX-wx*ns; cam.y=ev.clientY-wy*ns;
},{passive:false});
window.addEventListener("keydown",function(ev){if(ev.key==="Escape")closePanel();});
function hitTest(mx,my){
  var best=null,bd=1e9;
  nodes.forEach(function(n){var s=w2s(n);var d=Math.hypot(s.x-mx,s.y-my);var hit=Math.max(15,col(n).r*cam.s+8);if(d<hit&&d<bd){bd=d;best=n;}});
  if(best)openPanel(best); else closePanel();
}

// --- boot ------------------------------------------------------------------
function resize(){canvas.width=innerWidth*dpr;canvas.height=innerHeight*dpr;canvas.style.width=innerWidth+"px";canvas.style.height=innerHeight+"px";}
window.addEventListener("resize",resize);
resize(); initStars();
fetch("graph.json").then(function(r){return r.json();}).then(function(g){
  graph=g; nodes=(g.nodes||[]).slice().sort(function(a,b){return a.num-b.num;});
  edges=g.edges||[]; byNum={}; nodes.forEach(function(n){byNum[n.num]=n;});
  layout(); nodes.forEach(function(n){n._x=n.x;n._y=n.y;}); fitCamera(); buildHud(); render();
}).catch(function(err){
  document.getElementById("hud").innerHTML="<h1>Couldn't load graph.json</h1><div class=dest>"+String(err)+"</div>";
  render();
});
</script>
</body>
</html>`
