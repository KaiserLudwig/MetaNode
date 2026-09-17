/* 任务10：atomic 无锁计数器 —— 10 协程原子递增，无需加锁也安全 */
(function () {
  var stageEl = null;

  function render(stage) {
    stageEl = stage;
    var html = '<div class="scene">' +
      '<div class="big-num" id="at-count" style="left:530px;top:60px">0</div>' +
      '<div class="step-label" style="left:516px;top:110px">counter</div>' +
      '<div class="step-label" style="left:516px;top:128px">(sync/atomic 无锁)</div>';
    for (var i = 0; i < 10; i++) {
      var x = 55 + (i % 5) * 56;
      var y = i < 5 ? 34 : 156;
      html += '<div class="dot green" data-pt="' + i + '" style="left:' + x + 'px;top:' + y + 'px"></div>';
    }
    html += '<div class="formula" style="left:40px;top:196px">atomic.AddInt64(&amp;counter, 1) —— CPU 原子指令，无需加锁</div>';
    html += '</div>';
    stage.innerHTML = html;
  }

  function play(api) {
    var s = stageEl.querySelector('.scene');
    var count = s.querySelector('#at-count');
    var pts = s.querySelectorAll('.dot');
    var tick = api.el('div', 'check-tick', '✓');
    tick.style.cssText = 'left:300px;top:168px';
    var v = 0;

    return (async function () {
      api.subtitle('① 10 个协程并发执行 atomic.AddInt64(&counter, 1)');
      await api.wait(800);
      api.subtitle('② 原子操作由 CPU 保证"读-改-写"不可分割，无需加锁');
      await api.wait(400);

      for (var i = 0; i < 24; i++) {
        // 三个点同时"原子操作"
        [i % 10, (i + 3) % 10, (i + 7) % 10].forEach(function (k) {
          pts[k].classList.add('active');
        });
        await api.wait(100);
        v++;
        count.textContent = String(v);
        [i % 10, (i + 3) % 10, (i + 7) % 10].forEach(function (k) {
          pts[k].classList.remove('active');
        });
        await api.wait(60);
      }
      api.subtitle('③ 无锁但安全：不会出现数据竞争（演示为 24 次，实际 10000 次）');
      await api.numRoll(count, v, 10000, 900);
      await api.wait(400);
      api.subtitle('④ 结论：原子操作适合简单计数，比 Mutex 更轻量');
      s.appendChild(tick);
      await api.wait(1400);
    })();
  }

  window.MND.register(10, { render: render, play: play });
})();
