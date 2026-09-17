/* MND - MetaNode Demo 动画框架（零依赖） */
window.MND = (function () {
  'use strict';

  var reg = {};            // taskId -> {render, play}
  var speed = 1;           // 1 | 2 | 4
  var playing = false;
  var canvasEl = null, subEl = null, badgeEl = null;

  /* ---------- DOM helpers ---------- */
  function el(tag, cls, text) {
    var e = document.createElement(tag);
    if (cls) e.className = cls;
    if (text !== undefined && text !== null) e.textContent = text;
    return e;
  }

  /* ---------- 动画原语（均受速度缩放） ---------- */
  function wait(ms) {
    return new Promise(function (res) { setTimeout(res, ms / speed); });
  }

  // 补间任意数值 CSS 属性（camelCase），easeOutCubic
  function tween(elem, prop, from, to, dur, opts) {
    opts = opts || {};
    var unit = opts.unit || '';
    var digits = opts.digits != null ? opts.digits : 0;
    return new Promise(function (res) {
      var start = performance.now();
      (function tick(now) {
        var p = Math.min(1, (now - start) / (dur / speed));
        var v = from + (to - from) * (1 - Math.pow(1 - p, 3));
        elem.style[prop] = v.toFixed(digits) + unit;
        if (p < 1) requestAnimationFrame(tick); else res();
      })(performance.now());
    });
  }

  // 数字滚动（digits: 小数位数）
  function numRoll(elem, from, to, dur, digits) {
    digits = digits || 0;
    return new Promise(function (res) {
      var start = performance.now();
      (function tick(now) {
        var p = Math.min(1, (now - start) / (dur / speed));
        elem.textContent = (from + (to - from) * (1 - Math.pow(1 - p, 3))).toFixed(digits);
        if (p < 1) requestAnimationFrame(tick); else res();
      })(performance.now());
    });
  }

  // 打字机
  function typewrite(elem, text, dur) {
    return new Promise(function (res) {
      elem.textContent = '';
      var step = dur / Math.max(1, text.length) / speed;
      var i = 0;
      (function next() {
        if (i >= text.length) { res(); return; }
        elem.textContent += text[i++];
        setTimeout(next, step);
      })();
    });
  }

  /* ---------- 舞台控件 ---------- */
  function subtitle(text) {
    if (!subEl) return;
    subEl.textContent = '· ' + text;
    subEl.classList.remove('sub-flash');
    void subEl.offsetWidth;
    subEl.classList.add('sub-flash');
  }

  function setBadge(text, state) {
    if (!badgeEl) return;
    badgeEl.textContent = text;
    badgeEl.className = 'stage-badge' + (state ? ' ' + state : '');
  }

  function clearCanvas() {
    if (canvasEl) canvasEl.innerHTML = '';
  }

  /* ---------- 注册与启动 ---------- */
  function register(taskId, demo) {
    reg[taskId] = demo;
  }

  function boot(taskId) {
    var demo = reg[taskId];
    if (!demo) return;

    canvasEl = document.getElementById('stage-canvas');
    subEl = document.getElementById('stage-subtitle');
    badgeEl = document.getElementById('stage-badge');
    var btnRun = document.getElementById('btn-run');
    var btnReplay = document.getElementById('btn-replay');
    var btnSpeed = document.getElementById('btn-speed');
    var realOut = document.getElementById('real-output');

    // 页面加载：渲染静态示意图
    demo.render(canvasEl);

    function playAll() {
      if (playing) return;
      playing = true;
      btnRun.disabled = true;
      btnReplay.disabled = true;
      setBadge('运行中', 'running');
      clearCanvas();
      demo.render(canvasEl); // 重新画基础场景，动画在其上推进

      var api = {
        el: el,
        wait: wait,
        tween: tween,
        numRoll: numRoll,
        typewrite: typewrite,
        subtitle: subtitle,
        setBadge: setBadge,
        speed: function () { return speed; }
      };

      Promise.resolve()
        .then(function () { return demo.play(api); })
        .then(function () {
          setBadge('完成', 'done');
          btnRun.disabled = false;
          btnReplay.disabled = false;
          btnRun.textContent = '▶ 再次运行';
          playing = false;
          // 高亮下方真实运行结果，与动画对照
          if (realOut) {
            realOut.classList.remove('flash');
            void realOut.offsetWidth;
            realOut.classList.add('flash');
            realOut.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
          }
        })
        .catch(function (e) {
          setBadge('出错', 'idle');
          subtitle('动画出错: ' + (e && e.message ? e.message : e));
          btnRun.disabled = false;
          btnReplay.disabled = false;
          playing = false;
        });
    }

    btnRun.addEventListener('click', playAll);
    btnReplay.addEventListener('click', playAll);
    btnSpeed.addEventListener('click', function () {
      speed = speed === 1 ? 2 : speed === 2 ? 4 : 1;
      btnSpeed.textContent = '速度 ' + speed + 'x';
    });
  }

  return {
    register: register,
    boot: boot
  };
})();
