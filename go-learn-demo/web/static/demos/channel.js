/* 任务7：通道通信 —— 数字小球沿管道从生产者流向消费者 */
(function () {
  var stageEl = null;

  function render(stage) {
    stageEl = stage;
    stage.innerHTML =
      '<div class="scene">' +
        '<div class="step-label" style="left:30px;top:46px">生产者协程</div>' +
        '<div class="gopher" style="left:34px;top:72px">🐹</div>' +
        '<div class="step-label" style="left:556px;top:46px">消费者协程</div>' +
        '<div class="gopher" style="left:570px;top:72px">🐹</div>' +
        '<div class="pipe" style="left:95px;top:103px;width:450px"></div>' +
        '<div class="term" id="ch-term" style="left:350px;top:14px;width:270px;height:64px">接收: </div>' +
      '</div>';
  }

  function play(api) {
    var s = stageEl.querySelector('.scene');
    var term = s.querySelector('#ch-term');
    var tick = api.el('div', 'check-tick', '✓');
    tick.style.cssText = 'left:300px;top:180px';

    return (async function () {
      api.subtitle('① 创建无缓冲通道：ch := make(chan int)');
      await api.wait(800);
      api.subtitle('② 生产者协程逐个发送 1~10（无缓冲：发送后等待接收方）');
      await api.wait(500);

      for (var i = 1; i <= 10; i++) {
        var ball = api.el('div', 'ball', String(i));
        ball.style.cssText = 'left:85px;top:96px';
        s.appendChild(ball);
        api.subtitle('② 发送 ch <- ' + i + '  →  接收 <-ch 打印 ' + i);
        await api.tween(ball, 'left', 85, 555, 550);
        ball.remove();
        term.textContent += i + ' ';
        await api.wait(120);
      }
      api.subtitle('③ 发送完毕 close(ch)，消费者 range 感知关闭，通信结束');
      s.appendChild(tick);
      await api.wait(1400);
    })();
  }

  window.MND.register(7, { render: render, play: play });
})();
