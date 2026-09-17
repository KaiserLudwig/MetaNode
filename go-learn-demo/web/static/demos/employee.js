/* 任务6：Employee 组合 —— Person 字段"提升"飞入 Employee，终端打印信息 */
(function () {
  var stageEl = null;

  function render(stage) {
    stageEl = stage;
    stage.innerHTML =
      '<div class="scene">' +
        '<div class="var-box" style="left:40px;top:50px;width:180px;height:112px">' +
          '<span class="var-name" style="font-size:12px">Person（组合）</span>' +
          '<span class="var-addr">Name · Age</span>' +
        '</div>' +
        '<div class="var-box" style="left:420px;top:50px;width:190px;height:112px">' +
          '<span class="var-name" style="font-size:12px">Employee（员工）</span>' +
          '<span class="var-addr">EmployeeID: E1001</span>' +
        '</div>' +
        '<div class="field-block" id="f-name" style="left:55px;top:84px;width:72px;height:34px">Name: 张三</div>' +
        '<div class="field-block" id="f-age" style="left:136px;top:84px;width:72px;height:34px">Age: 28</div>' +
        '<div class="drop-zone" style="left:435px;top:126px;width:72px;height:26px">Name</div>' +
        '<div class="drop-zone" style="left:516px;top:126px;width:72px;height:26px">Age</div>' +
        '<div class="term" style="left:60px;top:180px;width:520px;height:30px" id="emp-term"></div>' +
      '</div>';
  }

  function play(api) {
    var s = stageEl.querySelector('.scene');
    var name = s.querySelector('#f-name'), age = s.querySelector('#f-age');
    var empBox = s.querySelectorAll('.var-box')[1];
    var term = s.querySelector('#emp-term');
    var tick = api.el('div', 'check-tick', '✓');
    tick.style.cssText = 'left:300px;top:140px';

    return (async function () {
      api.subtitle('① Employee 匿名组合 Person：字段和方法被"提升"');
      await api.wait(850);
      api.subtitle('② 提升后，Employee 可直接访问 Name / Age（emp.Name）');
      name.classList.add('hl'); age.classList.add('hl');
      await api.wait(600);

      // 字段块从 Person 盒飞入 Employee 盒（场景坐标系）
      api.subtitle('③ Name 与 Age 字段提升到 Employee —— 组合实现"继承"效果');
      var moves = [
        api.tween(name, 'left', 55, 435, 700),
        api.tween(name, 'top', 84, 126, 700),
        api.tween(age, 'left', 136, 516, 700),
        api.tween(age, 'top', 84, 126, 700)
      ];
      await Promise.all(moves);
      name.classList.remove('hl'); age.classList.remove('hl');
      empBox.classList.add('hl');
      await api.wait(350);

      api.subtitle('④ 调用 PrintInfo() 输出员工信息');
      await api.typewrite(term, '员工信息: ID=E1001 姓名=张三 年龄=28', 900);
      await api.wait(500);

      api.subtitle('⑤ 结论：Go 用组合代替继承，方法接收者定义行为');
      s.appendChild(tick);
      await api.wait(1300);
    })();
  }

  window.MND.register(6, { render: render, play: play });
})();
