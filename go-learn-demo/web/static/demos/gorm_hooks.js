/* 任务13：GORM 钩子 —— AfterCreate 自动加统计，AfterDelete 自动改状态 */
(function () {
  var stageEl = null;

  function render(stage) {
    stageEl = stage;
    stage.innerHTML =
      '<div class="scene">' +
        '<div class="table-card" style="left:15px;top:55px;width:150px;height:100px">' +
          '<div class="tname">张三</div>' +
          '<div class="tfield">用户</div>' +
          '<div class="tfield" style="color:#7dd3fc;font-size:24px;font-weight:700;line-height:1.2" id="gh-count">2</div>' +
          '<div class="tfield" style="color:#64748b">PostCount</div>' +
        '</div>' +
        '<div class="table-card" id="gh-post" style="left:210px;top:25px;width:405px;height:160px;opacity:0">' +
          '<div class="tname">《GORM 钩子实战》</div>' +
          '<div class="tfield" style="color:#64748b">新文章 · AfterCreate 触发</div>' +
          '<div class="status-pill has" id="gh-status" style="left:285px;top:70px">有评论</div>' +
        '</div>' +
        '<div class="comment-chip" id="gh-cc1" style="left:230px;top:118px;opacity:0">钩子真有用</div>' +
        '<div class="comment-chip" id="gh-cc2" style="left:340px;top:118px;opacity:0">学到了</div>' +
        '<div class="term" id="gh-term" style="left:60px;top:178px;width:520px;height:32px"></div>' +
      '</div>';
  }

  function play(api) {
    var s = stageEl.querySelector('.scene');
    var post = s.querySelector('#gh-post');
    var count = s.querySelector('#gh-count');
    var status = s.querySelector('#gh-status');
    var cc1 = s.querySelector('#gh-cc1'), cc2 = s.querySelector('#gh-cc2');
    var term = s.querySelector('#gh-term');
    var tick = api.el('div', 'check-tick', '✓');
    tick.style.cssText = 'left:300px;top:100px';

    return (async function () {
      api.subtitle('① 张三发布新文章《GORM 钩子实战》 → AfterCreate 钩子触发');
      post.style.opacity = '1';
      post.classList.add('fade-in');
      await api.numRoll(count, 2, 3, 800);
      await api.typewrite(term, 'AfterCreate: UPDATE users SET post_count = post_count + 1 ✓', 1000);
      await api.wait(500);

      api.subtitle('② 文章收到 2 条评论');
      cc1.style.opacity = '1'; cc2.style.opacity = '1';
      cc1.classList.add('fade-in'); cc2.classList.add('fade-in');
      await api.wait(700);

      api.subtitle('③ 删除 2 条评论 → AfterDelete 钩子检查剩余评论数');
      cc1.style.transition = 'opacity .4s';
      cc1.style.opacity = '0';
      await api.wait(450);
      cc2.style.opacity = '0';
      await api.wait(450);

      api.subtitle('④ 评论数为 0 → 钩子自动把文章状态更新为「无评论」');
      status.classList.remove('has');
      status.classList.add('none');
      status.textContent = '无评论';
      await api.typewrite(term, 'AfterDelete: 评论数=0 → UPDATE posts SET comment_status = 无评论 ✓', 1100);
      s.appendChild(tick);
      api.subtitle('⑤ 结论：钩子在事务内执行，自动维护统计与状态');
      await api.wait(1400);
    })();
  }

  window.MND.register(13, { render: render, play: play });
})();
