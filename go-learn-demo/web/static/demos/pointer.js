/* 任务1：指针加10 —— 变量盒子 + 内存地址 + 指针箭头，数值 5→15 */
(function () {
  var stageEl = null;

  function render(stage) {
    stageEl = stage;
    stage.innerHTML =
      '<div class="scene">' +
        '<svg style="position:absolute;left:0;top:0;width:640px;height:220px" viewBox="0 0 640 220">' +
          '<defs>' +
            '<marker id="mnd-arrow" markerWidth="10" markerHeight="10" refX="8" refY="3" orient="auto" markerUnits="strokeWidth">' +
              '<path d="M0,0 L9,3 L0,6 Z" fill="#8b5cf6"/>' +
            '</marker>' +
          '</defs>' +
          '<path id="pv-arrow" d="M 428 106 C 350 106 300 113 215 113" stroke="#8b5cf6" stroke-width="2.5" fill="none" ' +
            'marker-end="url(#mnd-arrow)" stroke-dasharray="300" stroke-dashoffset="300" stroke-linecap="round"/>' +
        '</svg>' +
        '<div class="addr-tag" style="left:60px;top:42px">&amp;n = 0xc000012345</div>' +
        '<div class="var-box" style="left:60px;top:70px;width:150px;height:86px">' +
          '<span class="var-name">n</span>' +
          '<span class="var-val" id="pv-val">5</span>' +
          '<span class="var-addr">0xc000012345</span>' +
        '</div>' +
        '<div class="func-box" style="left:430px;top:60px;width:170px;height:92px">' +
          '<span class="func-name">addTen</span>' +
          '<span class="func-sig">func (p *int)</span>' +
        '</div>' +
      '</div>';
  }

  function play(api) {
    var s = stageEl.querySelector('.scene');
    var val = s.querySelector('#pv-val');
    var arrow = s.querySelector('#pv-arrow');
    var box = s.querySelector('.var-box');
    var tick = api.el('div', 'check-tick', '✓');
    tick.style.cssText = 'left:300px;top:170px';

    return (async function () {
      api.subtitle('① 主函数声明变量 n = 5，在内存中分配地址 0xc000012345');
      await api.wait(900);

      api.subtitle('② 调用 addTen(&n)：把 n 的地址作为实参传入函数（引用传递）');
      await api.tween(arrow, 'strokeDashoffset', 300, 0, 750);
      box.classList.add('hl');
      await api.wait(450);

      api.subtitle('③ 函数内执行 *p += 10：通过指针直接修改该地址上的值');
      await api.numRoll(val, 5, 15, 900);
      await api.wait(400);

      api.subtitle('④ 返回主函数后 n = 15：传指针 = 引用传递，修改对调用方可见');
      s.appendChild(tick);
      await api.wait(1400);
    })();
  }

  window.MND.register(1, { render: render, play: play });
})();
