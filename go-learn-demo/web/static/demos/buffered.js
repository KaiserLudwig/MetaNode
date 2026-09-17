/* 任务8：缓冲通道 —— 100 个数字快速流过，可视化 10 格缓冲槽 */
(function () {
  var stageEl = null;

  function render(stage) {
    stageEl = stage;
    var html = '<div class="scene">' +
      '<div class="step-label" style="left:22px;top:46px">生产者</div>' +
      '<div class="gopher" style="left:24px;top:72px">🐹</div>' +
      '<div class="step-label" style="left:582px;top:46px">消费者</div>' +
      '<div class="gopher" style="left:588px;top:72px">🐹</div>' +
      '<div class="pipe" style="left:72px;top:103px;width:500px"></div>' +
      '<div class="term" id="bf-term" style="left:350px;top:14px;width:270px;height:64px">缓冲槽：</div>';
    for (var i = 0; i < 10; i++) {
      html += '<div class="slot" data-slot="' + i + '" style="left:' + (80 + i * 50) + 'px;top:74px"></div>';
    }
    html += '</div>';
    stage.innerHTML = html;
  }

  function play(api) {
    var s = stageEl.querySelector('.scene');
    var term = s.querySelector('#bf-term');
    var slots = s.querySelectorAll('.slot');
    var tick = api.el('div', 'check-tick', '✓');
    tick.style.cssText = 'left:300px;top:180px';
    var total = 100;
    var count = 0;

    return (async function () {
      api.subtitle('① 创建容量 10 的缓冲通道：ch := make(chan int, 10)');
      await api.wait(800);
      api.subtitle('② 生产者快速发送 100 个整数：缓冲区未满时发送不阻塞');
      await api.wait(400);

      for (var b = 0; b < 10; b++) {
        // 一批 10 个同时流动，缓冲槽依次亮起
        var batch = [];
        for (var k = 0; k < 10; k++) {
          var ball = api.el('div', 'ball', '');
          ball.style.cssText = 'left:62px;top:96px';
          s.appendChild(ball);
          batch.push(api.tween(ball, 'left', 62, 562, 480));
        }
        await Promise.all(batch);
        batch.forEach(function (ball) { ball.remove(); });
        slots.forEach(function (slot) {
          slot.classList.add('full');
          setTimeout(function () { slot.classList.remove('full'); }, 150 / api.speed());
        });
        count += 10;
        term.textContent = '缓冲槽：10 格 · 已发送 ' + count + ' / ' + total;
        await api.wait(140);
      }
      api.subtitle('③ 消费者持续接收，共 100 个整数全部送达');
      await api.typewrite(term, '消费者共接收 100 个整数', 700);
      s.appendChild(tick);
      api.subtitle('④ 缓冲机制：生产与消费解耦，未满不阻塞，容量用尽才等待');
      await api.wait(1400);
    })();
  }

  window.MND.register(8, { render: render, play: play });
})();
