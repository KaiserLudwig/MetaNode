/* 任务4：并发任务调度器 —— 并行进度条 + 耗时徽章 + 串行/并发对比条 */
(function () {
  var stageEl = null;

  function render(stage) {
    stageEl = stage;
    var jobs = [
      ['任务A 模拟下载', 300], ['任务B 模拟解析', 150],
      ['任务C 模拟渲染', 450], ['任务D 模拟上传', 200]
    ];
    var html = '<div class="scene">';
    for (var i = 0; i < 4; i++) {
      var y = 16 + i * 38;
      html += '<div class="step-label" style="left:20px;top:' + (y + 1) + 'px">' + jobs[i][0] + '</div>' +
        '<div class="tbar" style="left:150px;top:' + (y + 3) + 'px;width:250px">' +
          '<div class="tbar-fill" data-i="' + i + '"></div>' +
        '</div>' +
        '<div class="badge-num" data-ms="' + jobs[i][1] + '" style="left:420px;top:' + (y - 4) + 'px;opacity:0">' + jobs[i][1] + 'ms</div>';
    }
    html +=
      '<div class="step-label" style="left:20px;top:182px">串行 1102ms</div>' +
      '<div class="tbar" style="left:150px;top:182px;width:250px"><div class="tbar-fill" id="seq-bar" style="background:linear-gradient(90deg,#64748b,#94a3b8)"></div></div>' +
      '<div class="step-label" style="left:20px;top:203px">并发 451ms</div>' +
      '<div class="tbar" style="left:150px;top:203px;width:250px"><div class="tbar-fill" id="par-bar"></div></div>' +
      '</div>';
    stage.innerHTML = html;
  }

  function play(api) {
    var s = stageEl.querySelector('.scene');
    var fills = s.querySelectorAll('.tbar-fill[data-i]');
    var badges = s.querySelectorAll('.badge-num');
    var durs = [600, 300, 900, 400]; // 视觉时长（真实耗时 300/150/450/200ms 的 2x 慢放）

    return (async function () {
      api.subtitle('① 每个任务启动一个协程，4 个任务并发执行');
      await api.wait(600);
      api.subtitle('② 并发调度：任务各自独立推进，互不等待');
      var tasks = [];
      for (var i = 0; i < 4; i++) {
        (function (i) {
          tasks.push(api.tween(fills[i], 'width', 0, 100, durs[i], { unit: '%', digits: 1 })
            .then(function () {
              fills[i].classList.add('done');
              badges[i].style.opacity = '1';
              badges[i].classList.add('fade-in');
              return api.wait(120);
            }));
        })(i);
      }
      await Promise.all(tasks);
      api.subtitle('③ 总耗时 = 最慢任务（450ms）；而串行执行需要 1102ms');
      await api.tween(s.querySelector('#seq-bar'), 'width', 0, 100, 700, { unit: '%', digits: 1 });
      await api.tween(s.querySelector('#par-bar'), 'width', 0, 41, 700, { unit: '%', digits: 1 });
      await api.wait(900);
      api.subtitle('④ 结论：并发调度大幅缩短总耗时（451ms vs 1102ms）');
      await api.wait(1200);
    })();
  }

  window.MND.register(4, { render: render, play: play });
})();
