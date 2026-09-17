/* 任务12：GORM 关联查询 —— 预加载文章评论 + 聚合统计评论最多文章 */
(function () {
  var stageEl = null;

  function render(stage) {
    stageEl = stage;
    stage.innerHTML =
      '<div class="scene">' +
        '<svg style="position:absolute;left:0;top:0;width:640px;height:220px" viewBox="0 0 640 220">' +
          '<path class="gq-line" id="gq-a1" d="M 118 100 L 160 60" stroke="#38bdf8" stroke-width="2" fill="none" stroke-linecap="round" stroke-dasharray="70" stroke-dashoffset="70"/>' +
          '<path class="gq-line" id="gq-a2" d="M 118 115 L 160 155" stroke="#38bdf8" stroke-width="2" fill="none" stroke-linecap="round" stroke-dasharray="70" stroke-dashoffset="70"/>' +
        '</svg>' +
        '<div class="table-card" style="left:15px;top:70px;width:100px;height:70px">' +
          '<div class="tname">张三</div>' +
          '<div class="tfield">用户</div>' +
        '</div>' +
        '<div class="table-card" id="gq-p1" style="left:165px;top:16px;width:235px;height:100px">' +
          '<div class="tname">《Go 并发模型入门》</div>' +
          '<div class="tfield" style="color:#64748b">Post 1 · 评论 3 条</div>' +
        '</div>' +
        '<div class="table-card" id="gq-p2" style="left:165px;top:130px;width:235px;height:70px">' +
          '<div class="tname">《GORM 使用技巧》</div>' +
          '<div class="tfield" style="color:#64748b">Post 2 · 评论 1 条</div>' +
        '</div>' +
        '<div class="comment-chip" id="gq-c1" style="left:180px;top:62px;opacity:0">写得太好了</div>' +
        '<div class="comment-chip" id="gq-c2" style="left:290px;top:62px;opacity:0">收藏了</div>' +
        '<div class="comment-chip" id="gq-c3" style="left:340px;top:88px;opacity:0">期待续篇</div>' +
        '<div class="comment-chip" id="gq-c4" style="left:180px;top:176px;opacity:0">学到了</div>' +
        '<div class="badge-num" id="gq-cnt1" style="left:355px;top:24px;opacity:0">评论 0</div>' +
        '<div class="badge-num" id="gq-cnt2" style="left:355px;top:138px;opacity:0">评论 0</div>' +
        '<div class="rel-num" id="gq-trophy" style="left:200px;top:122px;opacity:0;font-size:26px">🏆</div>' +
        '<div class="term" id="gq-term" style="left:420px;top:34px;width:200px;height:140px"></div>' +
      '</div>';
  }

  function play(api) {
    var s = stageEl.querySelector('.scene');
    var a1 = s.querySelector('#gq-a1'), a2 = s.querySelector('#gq-a2');
    var chips = ['gq-c1', 'gq-c2', 'gq-c3', 'gq-c4'].map(function (id) { return s.querySelector('#' + id); });
    var cnt1 = s.querySelector('#gq-cnt1'), cnt2 = s.querySelector('#gq-cnt2');
    var p1 = s.querySelector('#gq-p1');
    var trophy = s.querySelector('#gq-trophy');
    var term = s.querySelector('#gq-term');
    var tick = api.el('div', 'check-tick', '✓');
    tick.style.cssText = 'left:300px;top:100px';

    return (async function () {
      api.subtitle('① Preload("Posts.Comments")：一次性预加载张三的文章与评论');
      await Promise.all([api.tween(a1, 'strokeDashoffset', 70, 0, 500), api.tween(a2, 'strokeDashoffset', 70, 0, 500)]);
      for (var i = 0; i < chips.length; i++) {
        chips[i].style.opacity = '1';
        chips[i].classList.add('fade-in');
        await api.wait(300);
      }
      await api.typewrite(term, '张三 → 2 篇文章\n共 4 条评论 ✓', 800);
      await api.wait(400);

      api.subtitle('② 聚合查询：COUNT(comments.id) GROUP BY post_id 排序取第一');
      await api.wait(300);
      cnt1.style.opacity = '1'; cnt2.style.opacity = '1';
      await Promise.all([api.numRoll(cnt1, 0, 3, 700), api.numRoll(cnt2, 0, 1, 700)]);
      cnt1.textContent = '评论 3'; cnt2.textContent = '评论 1';
      p1.classList.add('hl');
      trophy.style.opacity = '1';
      await api.typewrite(term, '张三 → 2 篇文章\n共 4 条评论 ✓\n\n🏆 评论最多：《Go 并发模型入门》 3 条', 1000);
      s.appendChild(tick);
      api.subtitle('③ 结论：Preload 解决 N+1 查询，聚合函数统计评论数');
      await api.wait(1400);
    })();
  }

  window.MND.register(12, { render: render, play: play });
})();
