/* 任务11：GORM 模型定义 —— 三张表卡片 + 一对多关系线 + 建表终端 */
(function () {
  var stageEl = null;

  function render(stage) {
    stageEl = stage;
    stage.innerHTML =
      '<div class="scene">' +
        '<svg style="position:absolute;left:0;top:0;width:640px;height:220px" viewBox="0 0 640 220">' +
          '<path class="gm-line" id="gm-l1" d="M 178 88 L 240 88" stroke="#8b5cf6" stroke-width="2" fill="none" stroke-linecap="round" stroke-dasharray="80" stroke-dashoffset="80"/>' +
          '<path class="gm-line" id="gm-l2" d="M 398 88 L 460 88" stroke="#8b5cf6" stroke-width="2" fill="none" stroke-linecap="round" stroke-dasharray="80" stroke-dashoffset="80"/>' +
        '</svg>' +
        '<div class="table-card" id="gm-user" style="left:25px;top:40px;width:150px;height:96px;opacity:0">' +
          '<div class="tname">User 用户</div>' +
          '<div class="tfield">ID<br>Name<br>PostCount<br>Posts</div>' +
        '</div>' +
        '<div class="table-card" id="gm-post" style="left:245px;top:40px;width:150px;height:96px;opacity:0">' +
          '<div class="tname">Post 文章</div>' +
          '<div class="tfield">ID<br>Title<br>UserID<br>Comments</div>' +
        '</div>' +
        '<div class="table-card" id="gm-comment" style="left:465px;top:40px;width:150px;height:96px;opacity:0">' +
          '<div class="tname">Comment 评论</div>' +
          '<div class="tfield">ID<br>Content<br>PostID</div>' +
        '</div>' +
        '<div class="rel-num" id="gm-n1" style="left:196px;top:62px;opacity:0">1</div>' +
        '<div class="rel-num" id="gm-n2" style="left:223px;top:100px;opacity:0">∞</div>' +
        '<div class="rel-num" id="gm-n3" style="left:418px;top:62px;opacity:0">1</div>' +
        '<div class="rel-num" id="gm-n4" style="left:443px;top:100px;opacity:0">∞</div>' +
        '<div class="term" id="gm-term" style="left:25px;top:168px;width:590px;height:36px"></div>' +
      '</div>';
  }

  function play(api) {
    var s = stageEl.querySelector('.scene');
    var cards = ['gm-user', 'gm-post', 'gm-comment'].map(function (id) { return s.querySelector('#' + id); });
    var l1 = s.querySelector('#gm-l1'), l2 = s.querySelector('#gm-l2');
    var nums = ['gm-n1', 'gm-n2', 'gm-n3', 'gm-n4'].map(function (id) { return s.querySelector('#' + id); });
    var term = s.querySelector('#gm-term');
    var tick = api.el('div', 'check-tick', '✓');
    tick.style.cssText = 'left:300px;top:130px';

    return (async function () {
      api.subtitle('① 定义三个模型：User / Post / Comment');
      for (var i = 0; i < 3; i++) {
        cards[i].style.opacity = '1';
        cards[i].classList.add('fade-in');
        await api.wait(420);
      }
      api.subtitle('② 关系：User 1—∞ Post（一对多），Post 1—∞ Comment');
      await Promise.all([
        api.tween(l1, 'strokeDashoffset', 80, 0, 600),
        api.tween(l2, 'strokeDashoffset', 80, 0, 600)
      ]);
      nums.forEach(function (n) { n.style.opacity = '1'; });
      await api.wait(500);

      api.subtitle('③ AutoMigrate 自动生成三张表的建表 SQL');
      await api.typewrite(term, '✓ AutoMigrate → users / posts / comments', 900);
      s.appendChild(tick);
      api.subtitle('④ 建表完成：关联关系由 GORM 自动处理（外键 + 预加载）');
      await api.wait(1400);
    })();
  }

  window.MND.register(11, { render: render, play: play });
})();
