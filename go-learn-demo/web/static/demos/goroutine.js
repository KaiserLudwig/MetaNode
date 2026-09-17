/* 任务3：奇偶协程 —— 双 Gopher 轨道并发，终端交错打字输出 */
(function () {
  var stageEl = null;

  function render(stage) {
    stageEl = stage;
    stage.innerHTML =
      '<div class="scene">' +
        '<div class="step-label" style="left:20px;top:30px">协程1 · go func()</div>' +
        '<div class="gopher" id="g1" style="left:24px;top:50px">🐹</div>' +
        '<div class="step-label" style="left:20px;top:120px">协程2 · go func()</div>' +
        '<div class="gopher" id="g2" style="left:24px;top:140px">🐹</div>' +
        '<div class="term" style="left:300px;top:20px;width:315px;height:180px">' +
          '<span id="gt-l1"></span><br><span id="gt-l2"></span>' +
        '</div>' +
      '</div>';
  }

  function play(api) {
    var s = stageEl.querySelector('.scene');
    var g1 = s.querySelector('#g1'), g2 = s.querySelector('#g2');
    var l1 = s.querySelector('#gt-l1'), l2 = s.querySelector('#gt-l2');
    var tick = api.el('div', 'check-tick', '✓');
    tick.style.cssText = 'left:150px;top:180px';
    var lines = ['奇数协程: 1 3 5 7 9', '偶数协程: 2 4 6 8 10'];
    var idx = [0, 0];

    return (async function () {
      api.subtitle('① 使用 go 关键字启动两个协程，并发执行');
      await api.wait(700);
      api.subtitle('② 两个协程同时工作（打印顺序随机，输出交错出现）');

      var bob = setInterval(function () {
        g1.style.transform = g1.style.transform === 'translateY(4px)' ? '' : 'translateY(4px)';
        g2.style.transform = g2.style.transform === 'translateY(4px)' ? '' : 'translateY(4px)';
      }, 260 / api.speed());

      while (idx[0] < lines[0].length || idx[1] < lines[1].length) {
        var li = Math.random() < 0.5 ? 0 : 1;
        if (idx[li] >= lines[li].length) li = 1 - li;
        (li === 0 ? l1 : l2).textContent += lines[li][idx[li]];
        idx[li]++;
        await api.wait(110);
      }
      clearInterval(bob);

      api.subtitle('③ 并发执行完毕：协程由 Go 运行时调度，输出顺序不固定');
      s.appendChild(tick);
      await api.wait(1400);
    })();
  }

  window.MND.register(3, { render: render, play: play });
})();
