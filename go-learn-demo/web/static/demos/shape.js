/* 任务5：Shape 接口 —— SVG 矩形/圆形 + 公式字幕 + 数值滚动 */
(function () {
  var stageEl = null;

  function render(stage) {
    stageEl = stage;
    stage.innerHTML =
      '<div class="scene">' +
        '<svg style="position:absolute;left:0;top:0;width:640px;height:220px" viewBox="0 0 640 220">' +
          '<rect id="sh-rect" x="60" y="80" width="100" height="60" rx="6" fill="rgba(99,102,241,.12)" stroke="#818cf8" stroke-width="2.5"/>' +
          '<text x="60" y="70" fill="#7dd3fc" font-size="12" font-family="Consolas">Rectangle{Width:5, Height:3}</text>' +
          '<text x="110" y="160" fill="#94a3b8" font-size="11" font-family="Consolas" text-anchor="middle">宽 5</text>' +
          '<text x="45" y="115" fill="#94a3b8" font-size="11" font-family="Consolas" text-anchor="end">高 3</text>' +
          '<circle id="sh-circle" cx="470" cy="110" r="55" fill="rgba(168,85,247,.12)" stroke="#c084fc" stroke-width="2.5"/>' +
          '<text x="470" y="70" fill="#7dd3fc" font-size="12" font-family="Consolas" text-anchor="middle">Circle{Radius:4}</text>' +
          '<text x="470" y="186" fill="#94a3b8" font-size="11" font-family="Consolas" text-anchor="middle">r = 4</text>' +
        '</svg>' +
        '<div class="badge-num" id="sh-a1" style="left:52px;top:22px;opacity:0">面积 0</div>' +
        '<div class="badge-num" id="sh-p1" style="left:138px;top:22px;opacity:0">周长 0</div>' +
        '<div class="badge-num" id="sh-a2" style="left:398px;top:22px;opacity:0">面积 0</div>' +
        '<div class="badge-num" id="sh-p2" style="left:484px;top:22px;opacity:0">周长 0</div>' +
      '</div>';
  }

  function play(api) {
    var s = stageEl.querySelector('.scene');
    var rect = s.querySelector('#sh-rect'), circle = s.querySelector('#sh-circle');
    var a1 = s.querySelector('#sh-a1'), p1 = s.querySelector('#sh-p1');
    var a2 = s.querySelector('#sh-a2'), p2 = s.querySelector('#sh-p2');

    return (async function () {
      api.subtitle('① 创建 Rectangle{宽5, 高3}，实现 Shape 接口的 Area() 与 Perimeter()');
      rect.style.filter = 'drop-shadow(0 0 8px rgba(129,140,248,.9))';
      await api.wait(750);

      api.subtitle('② 矩形：Area = 宽×高 = 15    Perimeter = 2×(5+3) = 16');
      a1.style.opacity = '1'; p1.style.opacity = '1';
      await Promise.all([api.numRoll(a1, 0, 15, 800), api.numRoll(p1, 0, 16, 800)]);
      a1.textContent = '面积 15.00'; p1.textContent = '周长 16.00';
      await api.wait(550);

      api.subtitle('③ 创建 Circle{半径4}：Area = πr² ≈ 50.27    Perimeter = 2πr ≈ 25.13');
      circle.style.filter = 'drop-shadow(0 0 8px rgba(192,132,252,.9))';
      a2.style.opacity = '1'; p2.style.opacity = '1';
      await Promise.all([api.numRoll(a2, 0, 50.27, 900, 2), api.numRoll(p2, 0, 25.13, 900, 2)]);
      await api.wait(550);

      api.subtitle('④ 通过 Shape 接口统一调用：多态 —— 同一接口，不同实现');
      await api.wait(1300);
    })();
  }

  window.MND.register(5, { render: render, play: play });
})();
