/* 任务2：切片指针 ×2 —— 5 个元素盒子逐个翻转翻倍 */
(function () {
  var stageEl = null;

  function render(stage) {
    stageEl = stage;
    var html = '<div class="scene">';
    for (var i = 0; i < 5; i++) {
      html += '<div class="var-box slice-box" style="left:' + (40 + i * 116) + 'px;top:70px;width:96px;height:78px">' +
        '<span class="var-name">nums[' + i + ']</span>' +
        '<span class="var-val slice-val">' + (i + 1) + '</span>' +
        '<span class="var-addr">0x..' + ['00', '08', '10', '18', '20'][i] + '</span>' +
      '</div>';
    }
    html += '<div class="formula" style="left:40px;top:182px">doubleSlice(&amp;nums) —— 通过切片指针修改底层数组</div>';
    html += '</div>';
    stage.innerHTML = html;
  }

  function play(api) {
    var s = stageEl.querySelector('.scene');
    var boxes = s.querySelectorAll('.slice-box');
    var vals = s.querySelectorAll('.slice-val');
    var tick = api.el('div', 'check-tick', '✓');
    tick.style.cssText = 'left:300px;top:168px';

    return (async function () {
      api.subtitle('① 创建切片 nums = [1 2 3 4 5]，底层是连续内存（元素地址递增）');
      await api.wait(900);
      for (var i = 0; i < 5; i++) {
        var cur = i + 1;
        api.subtitle('② 遍历第 ' + (i + 1) + ' 个：(*s)[' + i + '] *= 2  →  ' + cur + ' × 2');
        boxes[i].classList.add('hl');
        await api.numRoll(vals[i], cur, cur * 2, 500);
        boxes[i].classList.remove('hl');
        await api.wait(240);
      }
      api.subtitle('③ 遍历完成：所有元素翻倍 → nums = [2 4 6 8 10]');
      s.appendChild(tick);
      await api.wait(1400);
    })();
  }

  window.MND.register(2, { render: render, play: play });
})();
