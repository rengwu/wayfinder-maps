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
  #panel .md{flex:1;overflow:auto;border-top:1px solid #262b36;padding-top:14px;font-size:13px;line-height:1.65;color:#cfd4de}
  #panel .md>*:first-child{margin-top:0}
  #panel .md h1,#panel .md h2,#panel .md h3,#panel .md h4{color:#e6e9ef;line-height:1.3;margin:16px 0 8px}
  #panel .md h1{font-size:16px}
  #panel .md h2{font-size:14px;border-bottom:1px solid #20242e;padding-bottom:4px}
  #panel .md h3{font-size:13px;color:#b7bdc9}
  #panel .md h4{font-size:12px;color:#9aa2b1;text-transform:uppercase;letter-spacing:.04em}
  #panel .md p{margin:9px 0}
  #panel .md ul,#panel .md ol{margin:9px 0;padding-left:20px}
  #panel .md li{margin:3px 0}
  #panel .md code{font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:12px;
    background:#20242e;padding:1px 5px;border-radius:4px;color:#d6c6a0}
  #panel .md pre{background:#0c0f15;border:1px solid #20242e;border-radius:8px;padding:12px;overflow-x:auto;margin:10px 0}
  #panel .md pre code{background:none;padding:0;color:#c8cfda;font-size:12px;line-height:1.5;white-space:pre}
  #panel .md a,#panel .md .xlink{color:#7fa8ff;text-decoration:none;cursor:pointer}
  #panel .md a:hover,#panel .md .xlink:hover{text-decoration:underline}
  #panel .md strong{color:#e6e9ef;font-weight:700}
  #panel .md em{color:#d3b98a;font-style:italic}
  #panel .md hr{border:none;border-top:1px solid #262b36;margin:14px 0}
  #panel .md blockquote{border-left:3px solid #3a4150;margin:9px 0;padding:2px 0 2px 12px;color:#9aa2b1}
  #hint{position:fixed;bottom:14px;left:16px;z-index:5;font-size:11px;color:#4b5261}
  #backbtn{position:fixed;top:16px;left:16px;z-index:7;display:none;align-items:center;
    background:rgba(14,18,26,.72);border:1px solid #262b36;border-radius:999px;color:#b7bdc9;
    padding:7px 14px;font-size:12px;cursor:pointer;backdrop-filter:blur(8px)}
  #backbtn:hover{color:#e6e9ef;border-color:#3a4356}
  /* splash + map list overlays, floating over the live starfield */
  #splash,#maplist{position:fixed;inset:0;z-index:10;display:none;
    background:radial-gradient(80% 80% at 50% 30%,rgba(8,11,18,.6),rgba(5,7,13,.92))}
  #splash{align-items:center;justify-content:center}
  .splashcard{width:min(420px,90vw);text-align:center;padding:8px}
  .logo{font-size:30px;font-weight:600;letter-spacing:.02em;color:#e6e9ef}
  .tag{color:#8b93a3;font-size:13px;margin:6px 0 26px}
  #openfolder{display:inline-block;background:#2b62d6;border:none;color:#fff;font-size:14px;
    padding:11px 22px;border-radius:10px;cursor:pointer;box-shadow:0 4px 20px rgba(43,98,214,.35)}
  #openfolder:hover{background:#356fe6}
  .recent-h{margin:30px 0 10px;font-size:11px;text-transform:uppercase;letter-spacing:.08em;color:#6b7280;text-align:left}
  #recents{display:flex;flex-direction:column;gap:8px}
  .recent{display:flex;align-items:center;justify-content:space-between;text-align:left;
    background:rgba(23,26,33,.8);border:1px solid #262b36;border-radius:10px;padding:11px 14px;cursor:pointer;color:#e6e9ef}
  .recent:hover{border-color:#3a4356;background:rgba(30,35,44,.9)}
  .recent .rname{font-size:13px}
  .recent .rmeta{font-size:11px;color:#8b93a3}
  #maplist{flex-direction:column;padding:56px 40px 40px;overflow:auto}
  .listtop{display:flex;align-items:flex-end;justify-content:space-between;
    max-width:1000px;margin:0 auto 26px;width:100%}
  .listkick{font-size:11px;text-transform:uppercase;letter-spacing:.08em;color:#6b7280}
  #projname{font-size:22px;font-weight:600;color:#e6e9ef}
  #openanother{background:none;border:1px solid #262b36;color:#b7bdc9;border-radius:999px;padding:7px 14px;font-size:12px;cursor:pointer}
  #openanother:hover{color:#e6e9ef;border-color:#3a4356}
  #cards{display:grid;grid-template-columns:repeat(auto-fill,minmax(260px,1fr));gap:14px;
    max-width:1000px;margin:0 auto;width:100%}
  .card{text-align:left;background:rgba(23,26,33,.82);border:1px solid #262b36;border-radius:14px;
    padding:18px;cursor:pointer;transition:transform .08s,border-color .08s}
  .card:hover{transform:translateY(-2px);border-color:#4f8cff}
  .card .cname{font-size:15px;font-weight:600;color:#e6e9ef;line-height:1.35;margin-bottom:10px;
    display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden}
  .card .cmeta{display:flex;gap:10px;font-size:12px;color:#8b93a3;margin-bottom:10px}
  .card .cfront{color:#ffd873}
  .card .cbar{height:5px;background:#20242e;border-radius:999px;overflow:hidden}
  .card .cbar>span{display:block;height:100%;background:linear-gradient(90deg,#6d86ad,#b9c9e0)}
</style>
</head>
<body>
<canvas id="sky"></canvas>
<button id="backbtn">&larr; maps</button>
<div id="hud"></div>
<div id="panel"><span class="x">&times;</span><h2></h2><div class="meta"></div><div class="md"></div></div>
<div id="hint">drag to pan &middot; scroll to zoom &middot; click a star</div>

<div id="splash">
  <div class="splashcard">
    <div class="logo">&#10022; wayfinder</div>
    <div class="tag">a map of the way through a foggy effort</div>
    <button id="openfolder">Open Folder&hellip;</button>
    <div class="recent-h">Recent</div>
    <div id="recents"></div>
  </div>
</div>

<div id="maplist">
  <div class="listtop">
    <div><div class="listkick">project</div><div id="projname"></div></div>
    <button id="openanother">Open another&hellip;</button>
  </div>
  <div id="cards"></div>
</div>
<script>
"use strict";
var canvas=document.getElementById("sky"), ctx=canvas.getContext("2d");
var dpr=Math.max(1,window.devicePixelRatio||1);
var cam={x:0,y:0,s:1};
var graph=null, nodes=[], edges=[], byNum={}, selected=null;
var clock=0, T0=(window.performance&&performance.now?performance.now():Date.now());
var goal={x:0,y:0,s:1}, EASE=0.28; // cam eases toward goal each frame: pan, zoom, and select all move the goal
var POSEASE=0.12, lastClock=0, lastVersion=null, polling=false; // live-reload + node position tweening
var screen="splash", currentEffort=null, lastProject=null;      // splash | maplist | map

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

// relayoutWarm places a live update without disturbing the existing map. The
// force sim has free translation/rotation modes, so re-relaxing everything slides
// the whole constellation; instead, nodes already on screen are PINNED at their
// current position (px), and only freshly-added nodes are relaxed against them as
// fixed anchors. Survivors move exactly zero; a new star flies in and settles.
function relayoutWarm(){
  var fresh=[];
  nodes.forEach(function(n){ if(n.px!=null){n.x=n.px;n.y=n.py;} else fresh.push(n); });
  if(!fresh.length)return;
  var rnd=mulberry32(97+fresh.length);
  fresh.forEach(function(n){
    var ax=null,ay=null;
    edges.forEach(function(e){
      if(e.to===n.num&&byNum[e.from]&&byNum[e.from].px!=null){ax=byNum[e.from].x;ay=byNum[e.from].y;}
      else if(e.from===n.num&&byNum[e.to]&&byNum[e.to].px!=null){ax=byNum[e.to].x;ay=byNum[e.to].y;}
    });
    var ang=rnd()*6.2831853;
    if(ax!=null){n.x=ax+Math.cos(ang)*120;n.y=ay+Math.sin(ang)*120;}
    else{var R=ringR(n.rank);n.x=Math.cos(ang)*R;n.y=Math.sin(ang)*R;}
  });
  for(var it=0;it<200;it++){
    for(var k=0;k<fresh.length;k++){
      var n=fresh[k], fx=0, fy=0, i;
      for(i=0;i<nodes.length;i++){var o=nodes[i];if(o===n)continue;var dx=n.x-o.x,dy=n.y-o.y,d2=dx*dx+dy*dy||0.01,d=Math.sqrt(d2),f=9000/d2;fx+=dx/d*f;fy+=dy/d*f;}
      for(i=0;i<edges.length;i++){var e=edges[i],o=null;if(e.from===n.num)o=byNum[e.to];else if(e.to===n.num)o=byNum[e.from];if(!o)continue;var dx=o.x-n.x,dy=o.y-n.y,d=Math.hypot(dx,dy)||0.01,f=(d-150)*0.02;fx+=dx/d*f;fy+=dy/d*f;}
      var dd=Math.hypot(n.x,n.y)||0.01,rf=(ringR(n.rank)-dd)*0.05;fx+=n.x/dd*rf;fy+=n.y/dd*rf;
      n.x+=fx;n.y+=fy;
    }
  }
}

function fitCamera(){
  if(!nodes.length){cam={x:innerWidth/2,y:innerHeight/2,s:1};return;}
  var minx=1e9,miny=1e9,maxx=-1e9,maxy=-1e9;
  nodes.forEach(function(n){minx=Math.min(minx,n.x);miny=Math.min(miny,n.y);maxx=Math.max(maxx,n.x);maxy=Math.max(maxy,n.y);});
  for(var i=0;i<fogPts.length;i++){var f=fogPts[i];minx=Math.min(minx,f.x-70);miny=Math.min(miny,f.y-70);maxx=Math.max(maxx,f.x+70);maxy=Math.max(maxy,f.y+70);}
  var pad=70; minx-=pad;miny-=pad;maxx+=pad;maxy+=pad;
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

// --- fog nebulae -----------------------------------------------------------
// Fog patches (open questions not yet sharp enough to ticket) sit as faint
// nebulae in the void beyond the constellation, each tethered to the ticket
// that will clear it. Positions are deterministic: an anchored patch sits out
// past its ticket's angle, an unanchored one is spread by the golden angle.
var fogPts=[];
function setupFog(){
  fogPts=[];
  if(!graph||!graph.fog||!graph.fog.length||!nodes.length)return;
  var cx=0,cy=0; nodes.forEach(function(n){cx+=n.x;cy+=n.y;}); cx/=nodes.length; cy/=nodes.length;
  var rim=0; nodes.forEach(function(n){rim=Math.max(rim,Math.hypot(n.x-cx,n.y-cy));}); rim+=130;
  var rnd=mulberry32(4242);
  graph.fog.forEach(function(f,i){
    var anc=f.clearsWith?byNum[f.clearsWith]:null;
    var ang=anc?Math.atan2(anc.y-cy,anc.x-cx)+(rnd()-0.5)*0.3:(i*2.399963+rnd()*0.4);
    fogPts.push({title:f.title, anchor:anc||null, x:cx+Math.cos(ang)*rim, y:cy+Math.sin(ang)*rim});
  });
}
function drawFog(){
  for(var i=0;i<fogPts.length;i++){
    var f=fogPts[i];
    if(f.anchor){
      ctx.strokeStyle="rgba(150,132,205,0.28)"; ctx.lineWidth=1; ctx.setLineDash([3,7]); ctx.lineDashOffset=-clock*4;
      ctx.beginPath(); ctx.moveTo(f.x,f.y); ctx.lineTo(f.anchor._x,f.anchor._y); ctx.stroke();
      ctx.setLineDash([]); ctx.lineDashOffset=0;
    }
    var breathe=0.82+0.18*Math.sin(clock*0.8+f.x*0.01), R=66;
    var g=ctx.createRadialGradient(f.x,f.y,0,f.x,f.y,R);
    g.addColorStop(0,"rgba(122,98,192,"+(0.24*breathe)+")");
    g.addColorStop(0.5,"rgba(92,82,170,"+(0.10*breathe)+")");
    g.addColorStop(1,"rgba(92,82,170,0)");
    ctx.fillStyle=g; ctx.beginPath(); ctx.arc(f.x,f.y,R,0,6.2831853); ctx.fill();
    ctx.fillStyle="rgba(185,168,224,0.5)"; ctx.beginPath(); ctx.arc(f.x,f.y,2.2,0,6.2831853); ctx.fill();
  }
}
function drawFogLabels(){
  if(cam.s<0.42||!fogPts.length)return;
  var fs=clamp(10*Math.pow(cam.s,0.3),7.5,11.5);
  ctx.textAlign="center"; ctx.font="italic "+fs.toFixed(1)+"px ui-sans-serif,system-ui,sans-serif";
  ctx.shadowColor="rgba(0,0,0,0.8)"; ctx.shadowBlur=4; ctx.fillStyle="rgba(184,168,220,0.8)";
  for(var i=0;i<fogPts.length;i++){var f=fogPts[i];
    var sx=f.x*cam.s+cam.x, sy=f.y*cam.s+cam.y;
    var t=f.title.length>26?f.title.slice(0,25)+"…":f.title;
    ctx.fillText(t, sx, sy+18);
  }
  ctx.shadowBlur=0;
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
  var en=(n.enter!=null)?n.enter:1;      // 0..1 grow-in for a newly-added ticket
  var fl=n.flare||0;                      // 1..0 burst on a status change
  // Frontier stars breathe: glow and ring pulse gently so the eye is drawn to
  // what is takeable now.
  var isF=n.status==="frontier";
  var beat=0.5+0.5*Math.sin(clock*2.8);
  var pulse=isF?(0.8+0.2*beat):1;
  var gr=(isF?c.gr*(0.92+0.16*beat):c.gr)*(0.55+0.45*en)*(1+fl*0.5);
  var g=ctx.createRadialGradient(x,y,0,x,y,gr);
  g.addColorStop(0,hexA(c.glow,Math.min(1,(0.85*pulse+fl*0.5)*en))); g.addColorStop(0.4,hexA(c.glow,0.22*pulse*en)); g.addColorStop(1,hexA(c.glow,0));
  ctx.fillStyle=g; ctx.beginPath(); ctx.arc(x,y,gr,0,6.2831853); ctx.fill();
  ctx.fillStyle=c.core; ctx.beginPath(); ctx.arc(x,y,c.r*(0.4+0.6*en),0,6.2831853); ctx.fill();
  // Flare: an expanding ring pulse marking a just-changed status.
  if(fl>0){ctx.strokeStyle=hexA(c.core,fl*0.7);ctx.lineWidth=1.5+2*fl;ctx.beginPath();ctx.arc(x,y,c.r+(1-fl)*40,0,6.2831853);ctx.stroke();}
  if(isF){ctx.strokeStyle=hexA("#ffd873",0.4+0.3*beat);ctx.lineWidth=1.5;ctx.beginPath();ctx.arc(x,y,c.r+6+1.5*beat,0,6.2831853);ctx.stroke();}
  // Undermined: a red cracked halo — an uneven dashed ring, its gaps travelling,
  // marking a decision resting on a premise a later ticket destroyed.
  if(n.undermined){
    ctx.strokeStyle=hexA("#e06c75",0.55+0.2*Math.sin(clock*3.5+n.num));
    ctx.lineWidth=1.6; ctx.setLineDash([5,4,2,4]); ctx.lineDashOffset=-clock*9;
    ctx.beginPath(); ctx.arc(x,y,c.r+9,0,6.2831853); ctx.stroke();
    ctx.setLineDash([]); ctx.lineDashOffset=0;
  }
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
  var now=((window.performance&&performance.now?performance.now():Date.now())-T0)/1000;
  var dt=now-lastClock; if(dt<0||dt>0.1)dt=0.016; lastClock=now; clock=now;
  // Per node: ease the render base (px,py) toward the true layout position so a
  // structural update tweens rather than jumps; add idle bob on top; decay the
  // status-change flare and ramp a new node's grow-in. Edges/labels/hit-testing
  // all read _x/_y, so nothing detaches.
  for(var i=0;i<nodes.length;i++){var n=nodes[i], ph=n.num*1.7;
    if(n.px==null){n.px=n.x;n.py=n.y;}
    n.px+=(n.x-n.px)*POSEASE; n.py+=(n.y-n.py)*POSEASE;
    n._x=n.px+Math.sin(clock*0.7+ph)*2.4; n._y=n.py+Math.cos(clock*0.55+ph)*2.4;
    if(n.flare>0)n.flare=Math.max(0,n.flare-dt/1.1);
    if(n.enter!=null&&n.enter<1)n.enter=Math.min(1,n.enter+dt/0.55);
  }
  // Camera eases toward its goal every frame — pan, zoom and select all just
  // move the goal, giving weighty motion instead of instant jumps.
  cam.x+=(goal.x-cam.x)*EASE; cam.y+=(goal.y-cam.y)*EASE; cam.s+=(goal.s-cam.s)*EASE;
  ctx.setTransform(dpr,0,0,dpr,0,0);
  ctx.fillStyle="#05070d"; ctx.fillRect(0,0,W,H);
  drawStars(W,H);
  ctx.save(); ctx.translate(cam.x,cam.y); ctx.scale(cam.s,cam.s);
  drawFog(); edges.forEach(drawEdge); nodes.forEach(drawNode);
  ctx.restore();
  drawLabels(); drawFogLabels();
  requestAnimationFrame(render);
}

// --- hud & panel -----------------------------------------------------------
function el(tag,cls,txt){var e=document.createElement(tag);if(cls)e.className=cls;if(txt!=null)e.textContent=txt;return e;}

// --- a small markdown renderer for the detail panel ------------------------
// No dependency, no build step. HTML is escaped first, then a line-based block
// pass (fences, headings, lists, hr, blockquotes, paragraphs) wraps it, and an
// inline pass handles code / links / bold / italic. Code fences are detected
// via charCode 96 because a literal backtick would close this Go raw string.
function esc(s){return s.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;");}
function mdFmt(t){
  t=t.replace(/\[([^\]]+)\]\(([^)]+)\)/g,function(m,txt,url){
    if(/^https?:\/\//.test(url))return "<a href='"+url+"' target='_blank' rel='noopener'>"+txt+"</a>";
    var tm=url.match(/(?:^|\/)0*(\d+)-[^/)]*\.md/);
    if(tm)return "<a class='xlink' data-goto='"+tm[1]+"'>"+txt+"</a>";
    return "<span class='xlink'>"+txt+"</span>";
  });
  t=t.replace(/\*\*([^*]+)\*\*/g,"<strong>$1</strong>");
  t=t.replace(/\*([^*\n]+)\*/g,"<em>$1</em>");
  return t;
}
// Protect code spans as placeholders before emphasis, so bold can WRAP an
// inline-code span, and a * inside code (Go pointer types) is never italicised.
function mdInline(s){
  var BT=String.fromCharCode(96), codes=[];
  s=s.replace(new RegExp(BT+"([^"+BT+"]+)"+BT,"g"),function(m,c){codes.push(c);return "\x00"+(codes.length-1)+"\x00";});
  s=mdFmt(s);
  return s.replace(/\x00(\d+)\x00/g,function(m,i){return "<code>"+codes[i]+"</code>";});
}
function mdToHtml(src){
  if(!src)return "<p class='muted'>(no body)</p>";
  var BT=String.fromCharCode(96), FENCE=BT+BT+BT;
  var lines=src.split("\n"), out=[], para=[], list=null, inCode=false, code=[];
  function fp(){if(para.length){out.push("<p>"+mdInline(esc(para.join(" ")))+"</p>");para=[];}}
  function fl(){if(list){out.push("</"+list+">");list=null;}}
  for(var i=0;i<lines.length;i++){
    var line=lines[i], t=line.replace(/^\s+/,"");
    if(t.indexOf(FENCE)===0){
      if(inCode){out.push("<pre><code>"+esc(code.join("\n"))+"</code></pre>");inCode=false;code=[];}
      else{fp();fl();inCode=true;code=[];}
      continue;
    }
    if(inCode){code.push(line);continue;}
    if(t===""){fp();fl();continue;}
    var h=t.match(/^(#{1,6})\s+(.*)$/);
    if(h){fp();fl();var lv=h[1].length;out.push("<h"+lv+">"+mdInline(esc(h[2]))+"</h"+lv+">");continue;}
    if(/^(---+|\*\*\*+|___+)$/.test(t)){fp();fl();out.push("<hr>");continue;}
    var li=t.match(/^([-*+]|\d+\.)\s+(.*)$/);
    if(li){fp();var ty=/\d/.test(li[1])?"ol":"ul";if(list&&list!==ty)fl();if(!list){out.push("<"+ty+">");list=ty;}out.push("<li>"+mdInline(esc(li[2]))+"</li>");continue;}
    var bq=t.match(/^>\s?(.*)$/);
    if(bq){fp();fl();out.push("<blockquote>"+mdInline(esc(bq[1]))+"</blockquote>");continue;}
    if(list)fl();
    para.push(t);
  }
  if(inCode)out.push("<pre><code>"+esc(code.join("\n"))+"</code></pre>");
  fp();fl();
  return out.join("\n");
}
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

function fillPanel(n){
  var p=document.getElementById("panel");
  var h=p.querySelector("h2"); h.innerHTML=""; h.appendChild(el("span","num",pad2(n.num))); h.appendChild(document.createTextNode(n.title));
  var meta=p.querySelector(".meta"); meta.innerHTML="";
  meta.appendChild(el("span","c "+(n.status==="out_of_scope"?"oos":n.status), n.status.replace(/_/g," ")));
  if(n.type)meta.appendChild(el("span","c blocked",n.type));
  (n.blockers||[]).forEach(function(b){
    var dep=byNum[b], ok=dep&&dep.status==="resolved";
    meta.appendChild(el("span","c "+(ok?"resolved":"blocked"),"blocked by "+pad2(b)));
  });
  if(n.undermined)meta.appendChild(el("span","c blocked","undermined"));
  p.querySelector(".md").innerHTML=mdToHtml(n.body);
}
function openPanel(n){
  selected=n;
  // Ease the camera so the star sits centred in the space left of the panel.
  var p=document.getElementById("panel"), pw=p.offsetWidth||0, vx=(innerWidth-pw)/2, vy=innerHeight/2;
  goal.x=vx-n.x*goal.s; goal.y=vy-n.y*goal.s;
  fillPanel(n); p.classList.add("open");
}
function closePanel(){selected=null;document.getElementById("panel").classList.remove("open");}
document.querySelector("#panel .x").onclick=closePanel;
// Cross-ticket links in the rendered body jump the star-map to that node.
document.querySelector("#panel .md").addEventListener("click",function(ev){
  var a=ev.target.closest?ev.target.closest("[data-goto]"):null;
  if(a){ev.preventDefault();var n=byNum[parseInt(a.getAttribute("data-goto"),10)];if(n)openPanel(n);}
});

// --- camera interaction ----------------------------------------------------
var down=false, moved=0, last={x:0,y:0};
canvas.addEventListener("mousedown",function(ev){down=true;moved=0;last={x:ev.clientX,y:ev.clientY};canvas.classList.add("drag");});
window.addEventListener("mousemove",function(ev){
  if(!down)return; var dx=ev.clientX-last.x, dy=ev.clientY-last.y;
  goal.x+=dx; goal.y+=dy; moved+=Math.abs(dx)+Math.abs(dy); last={x:ev.clientX,y:ev.clientY};
});
window.addEventListener("mouseup",function(ev){
  canvas.classList.remove("drag"); if(!down)return; down=false;
  if(moved<5)hitTest(ev.clientX,ev.clientY);
});
canvas.addEventListener("wheel",function(ev){
  ev.preventDefault();
  var f=Math.exp(-ev.deltaY*0.0012); var ns=clamp(goal.s*f,0.15,4);
  var wx=(ev.clientX-goal.x)/goal.s, wy=(ev.clientY-goal.y)/goal.s;
  goal.s=ns; goal.x=ev.clientX-wx*ns; goal.y=ev.clientY-wy*ns;
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
function sortNodes(g){return (g.nodes||[]).slice().sort(function(a,b){return a.num-b.num;});}

// applyGraph loads a graph cold: fresh node objects (no px) so layout runs from
// scratch, then fit the camera. Used when a map is first opened or switched to.
function applyGraph(g){
  graph=g; nodes=sortNodes(g); edges=g.edges||[]; byNum={}; nodes.forEach(function(n){byNum[n.num]=n;});
  layout(); nodes.forEach(function(n){n.px=n.x;n.py=n.y;n._x=n.x;n._y=n.y;}); setupFog(); fitCamera();
  goal.x=cam.x; goal.y=cam.y; goal.s=cam.s; buildHud();
}

// updateGraph folds a freshly-fetched graph into the live scene: surviving nodes
// are mutated in place (so selection and screen position persist), status changes
// fire a flare, new tickets grow in, and layout re-runs with a warm start so
// positions tween. The camera is deliberately left alone.
function updateGraph(ng){
  var incoming=sortNodes(ng), keep=[];
  incoming.forEach(function(nn){
    var old=byNum[nn.num];
    if(old){
      if(old.status!==nn.status||old.undermined!==nn.undermined)old.flare=1;
      old.status=nn.status; old.undermined=nn.undermined; old.claimedBy=nn.claimedBy;
      old.title=nn.title; old.type=nn.type; old.body=nn.body; old.rank=nn.rank; old.blockers=nn.blockers;
      keep.push(old);
    } else { nn.enter=0; nn.flare=1; keep.push(nn); }
  });
  nodes=keep; byNum={}; nodes.forEach(function(n){byNum[n.num]=n;});
  edges=ng.edges||[]; graph=ng;
  relayoutWarm();                                            // survivors pinned; only new nodes placed
  nodes.forEach(function(n){if(n.px==null){n.px=n.x;n.py=n.y;}}); // new nodes appear at their final spot
  setupFog(); buildHud();                                     // note: no fitCamera — keep the user's view
  if(selected){ if(byNum[selected.num])fillPanel(selected); else closePanel(); }
}

// One poller for the whole app: only runs while a map is open, and re-checks the
// effort mid-flight so switching maps never applies stale data.
function startPolling(){
  setInterval(function(){
    var eff=currentEffort;
    if(!eff||polling)return; polling=true;
    fetch("/api/version?effort="+encodeURIComponent(eff)).then(function(r){return r.text();}).then(function(v){
      v=v.trim();
      if(currentEffort!==eff)return;
      if(lastVersion===null){lastVersion=v; return;}
      if(v===lastVersion) return;
      lastVersion=v;
      return fetch("/api/graph?effort="+encodeURIComponent(eff)).then(function(r){return r.json();}).then(function(g){if(currentEffort===eff)updateGraph(g);});
    }).then(function(){polling=false;}).catch(function(){polling=false;});
  },1500);
}

// --- screens ---------------------------------------------------------------
function setScreen(s){
  screen=s;
  document.getElementById("splash").style.display=s==="splash"?"flex":"none";
  document.getElementById("maplist").style.display=s==="maplist"?"flex":"none";
  document.getElementById("hud").style.display=s==="map"?"block":"none";
  document.getElementById("backbtn").style.display=s==="map"?"flex":"none";
  document.getElementById("hint").style.display=s==="map"?"block":"none";
  if(s!=="map")closePanel();
}
function showSplash(){
  setScreen("splash");
  var rc=document.getElementById("recents"); rc.innerHTML="";
  fetch("/api/recents").then(function(r){return r.json();}).then(function(rs){
    if(!rs||!rs.length){rc.innerHTML="<div class='muted' style='text-align:left;font-size:12px'>No recent projects yet.</div>";return;}
    rs.forEach(function(p){
      var b=el("button","recent");
      b.appendChild(el("span","rname",p.name));
      b.appendChild(el("span","rmeta",p.maps+" map"+(p.maps===1?"":"s")));
      b.onclick=function(){openProject(p.path);};
      rc.appendChild(b);
    });
  });
}
function openProject(path){
  lastProject=path;
  fetch("/api/maps?project="+encodeURIComponent(path)).then(function(r){return r.json();}).then(function(maps){
    document.getElementById("projname").textContent=path.replace(/\/+$/,"").split("/").pop();
    var g=document.getElementById("cards"); g.innerHTML="";
    if(!maps||!maps.length){g.innerHTML="<div class='muted'>No wayfinder maps in this project's .plan/</div>";}
    (maps||[]).forEach(function(m){
      var c=el("button","card");
      c.appendChild(el("div","cname",m.name));
      var meta=el("div","cmeta");
      meta.appendChild(el("span",null,m.resolved+"/"+m.total+" resolved"));
      if(m.frontier)meta.appendChild(el("span","cfront",m.frontier+" on frontier"));
      c.appendChild(meta);
      var bar=el("div","cbar"), sp=document.createElement("span");
      sp.style.width=(m.total?Math.round(m.resolved*100/m.total):0)+"%"; bar.appendChild(sp); c.appendChild(bar);
      c.onclick=function(){loadMap(m.path);};
      g.appendChild(c);
    });
    setScreen("maplist");
  });
}
function loadMap(effort){
  currentEffort=effort; lastVersion=null; selected=null;
  fetch("/api/graph?effort="+encodeURIComponent(effort)).then(function(r){return r.json();}).then(function(g){
    applyGraph(g); setScreen("map");
  });
}

document.getElementById("openfolder").onclick=function(){
  fetch("/api/pick").then(function(r){return r.json();}).then(function(d){if(d.path)openProject(d.path);});
};
document.getElementById("openanother").onclick=showSplash;
document.getElementById("backbtn").onclick=function(){currentEffort=null; if(lastProject)openProject(lastProject); else showSplash();};

resize(); initStars(); render(); startPolling();
fetch("/api/initial").then(function(r){return r.json();}).then(function(init){
  if(init.effort)loadMap(init.effort);
  else if(init.project)openProject(init.project);
  else showSplash();
}).catch(showSplash);
</script>
</body>
</html>`
