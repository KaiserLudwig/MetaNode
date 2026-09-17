/* 任务9：Mutex 计数器 —— 10 协程轮番抢锁，计数器跳动到 10000 */
(function () {
  var stageEl = null;

  function render(stage) {
    stageEl = stage;
    var html = '<div class="scene">' +
      '<div class="lock-big" id="mx-lock" style="left:292px;top:82px">🔒</div>' +
      '<div class="big-num" id="mx-count" style="left:530px;top:60px">0</div>' +
      '<div class="step-label" style="left:516px;top:110px">counter</div>' +
      '<div class="step-label" style="left:516px;top:128px">(sync.Mutex 保护)</div>';
    for (var i = 0; i < 10; i++) {
      var x = 55 + (i % 5) * 56;
      var y = i < 5 ? 34 : 156;
      html += '<div class="dot" data-pt="' + i + '" style="left:' + x + 'px;top:' + y + 'px"></div>';
    }
    html += '</div>';
    stage.innerHTML = html;
  }

  function play(api) {
    var s = stageEl.querySelector('.scene');
    var lock = s.querySelector('#mx-lock');
    var count = s.querySelector('#mx-count');
    var pts = s.querySelectorAll('.dot');
    var tick = api.el('div', 'check-tick', '✓');
    tick.style.cssText = 'left:300px;top:180px';
    var v = 0;

    return (async function () {
      api.subtitle('① 10 个协程同时执行 counter++（共 10000 次）');
      await api.wait(800);
      api.subtitle('② 每次递增前 mu.Lock() 抢锁：同一时刻只有一个协程能修改');
      await api.wait(400);

      for (var i = 0; i < 24; i++) {
        var p = pts[i % 10];
        p.classList.add('active');
        lock.classList.add('locked');
        await api.wait(110);
        v++;
        count.textContent = String(v);
        p.classList.remove('active');
        lock.classList.remove('locked');
        await api.wait(80);
      }
      api.subtitle('③ 锁保证互斥：counter++ 不会丢失更新（演示为 24 次，实际 10000 次）');
      await api.numRoll(count, v, 10000, 900);
      await api.wait(400);
      api.subtitle('④ 结论：sync.Mutex 让并发数据安全，最终 counter = 10000');
      s.appendChild(tick);
      await api.wait(1400);
    })();
  }

  window.MND.register(9, { render: render, play: play });
})();
