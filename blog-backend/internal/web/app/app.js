/* 博客前端 SPA：hash 路由 + fetch 调用后端 API */
(function () {
  'use strict';

  var API = '/api/v1';
  var TOKEN_KEY = 'blog_token';
  var USER_KEY = 'blog_user';
  var token = localStorage.getItem(TOKEN_KEY) || '';
  var user = JSON.parse(localStorage.getItem(USER_KEY) || 'null');

  var app = document.getElementById('app');
  var toastBox = document.getElementById('toast');
  var navUser = document.getElementById('nav-user');

  /* ---------- 工具 ---------- */
  function esc(s) {
    return String(s == null ? '' : s).replace(/[&<>"']/g, function (ch) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[ch];
    });
  }
  function fmtTime(s) {
    if (!s) return '';
    var d = new Date(s);
    var p = function (n) { return n < 10 ? '0' + n : String(n); };
    return d.getFullYear() + '-' + p(d.getMonth() + 1) + '-' + p(d.getDate()) + ' ' + p(d.getHours()) + ':' + p(d.getMinutes());
  }
  function toast(msg, type) {
    var el = document.createElement('div');
    el.className = 'toast-item ' + (type || '');
    el.textContent = msg;
    toastBox.appendChild(el);
    setTimeout(function () { el.remove(); }, 2600);
  }
  function setUser(u) {
    user = u;
    if (u) {
      localStorage.setItem(USER_KEY, JSON.stringify(u));
    } else {
      localStorage.removeItem(USER_KEY);
      localStorage.removeItem(TOKEN_KEY);
    }
    updateNav();
  }
  function updateNav() {
    document.getElementById('nav-new').classList.toggle('hidden', !user);
    if (user) {
      navUser.innerHTML =
        '<span class="uname">' + esc(user.username) + '</span>' +
        '<button class="btn btn-ghost" id="btn-logout">退出</button>';
      document.getElementById('btn-logout').addEventListener('click', function () {
        token = '';
        setUser(null);
        toast('已退出登录', 'ok');
        location.hash = '#/';
      });
    } else {
      navUser.innerHTML = '<a class="btn btn-ghost" href="#/login">登录 / 注册</a>';
    }
  }

  /* ---------- API 封装 ---------- */
  function api(path, opts) {
    opts = opts || {};
    var headers = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = 'Bearer ' + token;
    return fetch(API + path, {
      method: opts.method || 'GET',
      headers: headers,
      body: opts.body ? JSON.stringify(opts.body) : undefined
    }).then(function (res) {
      return res.json().catch(function () { return { code: res.status, message: 'HTTP ' + res.status }; });
    }).then(function (json) {
      if (json.code !== 0) {
        var err = new Error(json.message || '请求失败');
        err.code = json.code;
        throw err;
      }
      return json.data;
    });
  }

  /* ---------- 渲染入口（hash 路由） ---------- */
  function render() {
    var hash = location.hash || '#/';
    var path = hash.slice(1).split('?')[0];
    var parts = path.split('/').filter(Boolean);
    if (parts[0] === 'post' && parts[1]) renderPost(parts[1]);
    else if (parts[0] === 'new') renderNew();
    else if (parts[0] === 'edit' && parts[1]) renderEdit(parts[1]);
    else if (parts[0] === 'login') renderLogin();
    else renderList();
    window.scrollTo(0, 0);
  }
  window.addEventListener('hashchange', render);

  /* ---------- 文章列表 ---------- */
  function renderList() {
    app.innerHTML = '<div class="loading">加载中...</div>';
    api('/posts?page=1&size=12').then(function (data) {
      var html = '<h2 class="page-title">📚 全部文章 <span class="muted">(共 ' + data.total + ' 篇)</span></h2>';
      if (!data.items.length) {
        html += '<div class="error-box">还没有文章，' + (user ? '<a href="#/new">点击写第一篇 →</a>' : '登录后即可发布') + '</div>';
      } else {
        html += '<div class="post-grid">';
        data.items.forEach(function (p) {
          html += '<div class="post-card" data-id="' + p.id + '">' +
            '<h3>' + esc(p.title) + '</h3>' +
            '<div class="meta"><span class="author">' + esc(p.author.username) + '</span>' +
            '<span>💬 ' + p.comment_count + '</span><span>' + fmtTime(p.created_at) + '</span></div>' +
            '</div>';
        });
        html += '</div>';
      }
      app.innerHTML = html;
      app.querySelectorAll('.post-card').forEach(function (card) {
        card.addEventListener('click', function () { location.hash = '#/post/' + card.dataset.id; });
      });
    }).catch(function (e) {
      app.innerHTML = '<div class="error-box">加载失败：' + esc(e.message) + '</div>';
    });
  }

  /* ---------- 文章详情 ---------- */
  function renderPost(id) {
    app.innerHTML = '<div class="loading">加载中...</div>';
    Promise.all([api('/posts/' + id), api('/posts/' + id + '/comments')]).then(function (rs) {
      var p = rs[0], comments = rs[1];
      var isMine = user && p.author.id === user.id;
      var html = '<div class="post-detail">' +
        '<h1>' + esc(p.title) + '</h1>' +
        '<div class="meta">✍ <span class="author">' + esc(p.author.username) + '</span>' +
        ' · 💬 ' + p.comment_count + ' · 🕐 ' + fmtTime(p.created_at) +
        (p.updated_at && p.updated_at !== p.created_at ? ' · 更新于 ' + fmtTime(p.updated_at) : '') +
        '</div>' +
        '<div class="post-content">' + esc(p.content) + '</div>';
      if (isMine) {
        html += '<div class="post-actions">' +
          '<a class="btn" href="#/edit/' + p.id + '">✏ 编辑</a>' +
          '<button class="btn btn-danger" id="btn-del">🗑 删除</button></div>';
      }
      html += '</div>';

      // 评论区
      html += '<div class="comment-section"><h3>💬 评论 (' + comments.total + ')</h3>';
      if (!comments.items.length) html += '<div class="muted" style="margin-bottom:12px">暂无评论，来抢沙发~</div>';
      comments.items.forEach(function (cm) {
        html += '<div class="comment-item"><div class="head">' +
          '<span class="author">' + esc(cm.author.username) + '</span>' +
          '<span class="muted">' + fmtTime(cm.created_at) + '</span></div>' +
          '<div class="content">' + esc(cm.content) + '</div></div>';
      });
      if (user) {
        html += '<div class="comment-form">' +
          '<textarea id="cmt-content" placeholder="写下你的评论..."></textarea>' +
          '<button class="btn" id="btn-cmt" style="margin-top:10px">发表评论</button></div>';
      } else {
        html += '<div class="muted" style="margin-top:14px"><a href="#/login">登录</a> 后即可发表评论</div>';
      }
      html += '</div>';

      app.innerHTML = html;

      var delBtn = document.getElementById('btn-del');
      if (delBtn) delBtn.addEventListener('click', function () {
        if (!confirm('确定删除这篇文章吗？其下的评论也会一并删除。')) return;
        api('/posts/' + id, { method: 'DELETE' }).then(function () {
          toast('文章已删除', 'ok');
          location.hash = '#/';
        }).catch(function (e) { toast(e.message, 'err'); });
      });
      var cmtBtn = document.getElementById('btn-cmt');
      if (cmtBtn) cmtBtn.addEventListener('click', function () {
        var content = document.getElementById('cmt-content').value.trim();
        if (!content) return toast('评论内容不能为空', 'err');
        api('/posts/' + id + '/comments', { method: 'POST', body: { content: content } }).then(function () {
          toast('评论成功', 'ok');
          renderPost(id);
        }).catch(function (e) { toast(e.message, 'err'); });
      });
    }).catch(function (e) {
      app.innerHTML = '<div class="error-box">' + esc(e.message) + '</div>' +
        '<p style="margin-top:14px"><a class="btn btn-ghost" href="#/">← 返回列表</a></p>';
    });
  }

  /* ---------- 写文章 / 编辑 ---------- */
  function renderNew() {
    if (!user) { location.hash = '#/login'; return; }
    renderForm({ mode: 'create' });
  }
  function renderEdit(id) {
    if (!user) { location.hash = '#/login'; return; }
    app.innerHTML = '<div class="loading">加载中...</div>';
    api('/posts/' + id).then(function (p) {
      if (p.author.id !== user.id) {
        app.innerHTML = '<div class="error-box">只能编辑自己的文章</div>';
        return;
      }
      renderForm({ mode: 'edit', post: p });
    }).catch(function (e) {
      app.innerHTML = '<div class="error-box">' + esc(e.message) + '</div>';
    });
  }
  function renderForm(opt) {
    var p = opt.post || { title: '', content: '' };
    app.innerHTML =
      '<h2 class="page-title">' + (opt.mode === 'edit' ? '✏ 编辑文章' : '✍ 写文章') + '</h2>' +
      '<div class="post-detail">' +
      '<div class="form-row"><label>标题</label>' +
      '<input type="text" id="f-title" maxlength="128" value="' + esc(p.title) + '" placeholder="文章标题"></div>' +
      '<div class="form-row"><label>内容</label>' +
      '<textarea id="f-content" style="min-height:200px" placeholder="正文内容...">' + esc(p.content) + '</textarea></div>' +
      '<button class="btn" id="f-submit">' + (opt.mode === 'edit' ? '保存修改' : '发布文章') + '</button>' +
      ' <a class="btn btn-ghost" href="#/">取消</a>' +
      '</div>';
    document.getElementById('f-submit').addEventListener('click', function () {
      var title = document.getElementById('f-title').value.trim();
      var content = document.getElementById('f-content').value.trim();
      if (!title || !content) return toast('标题和内容都不能为空', 'err');
      var req = opt.mode === 'edit'
        ? api('/posts/' + p.id, { method: 'PUT', body: { title: title, content: content } })
        : api('/posts', { method: 'POST', body: { title: title, content: content } });
      req.then(function (d) {
        toast(opt.mode === 'edit' ? '保存成功' : '发布成功！', 'ok');
        location.hash = '#/post/' + (opt.mode === 'edit' ? p.id : d.id);
      }).catch(function (e) { toast(e.message, 'err'); });
    });
  }

  /* ---------- 登录 / 注册 ---------- */
  function renderLogin() {
    if (user) { location.hash = '#/'; return; }
    var isRegister = location.hash.indexOf('/register') === 0;
    app.innerHTML =
      '<div class="auth-card">' +
      '<div class="auth-tabs">' +
      '<a href="#/login" class="' + (isRegister ? '' : 'active') + '">登录</a>' +
      '<a href="#/register" class="' + (isRegister ? 'active' : '') + '">注册</a></div>' +
      (isRegister
        ? '<div class="form-row"><label>用户名（3-32 字符）</label><input type="text" id="a-username" maxlength="32"></div>' +
          '<div class="form-row"><label>邮箱</label><input type="email" id="a-email"></div>' +
          '<div class="form-row"><label>密码（至少 8 位）</label><input type="password" id="a-password"></div>' +
          '<button class="btn btn-block" id="a-submit">注 册</button>' +
          '<div class="hint">注册后自动登录，密码经 bcrypt 加密存储</div>'
        : '<div class="form-row"><label>用户名</label><input type="text" id="a-username"></div>' +
          '<div class="form-row"><label>密码</label><input type="password" id="a-password"></div>' +
          '<button class="btn btn-block" id="a-submit">登 录</button>' +
          '<div class="hint">演示账号：demo / demo123456</div>') +
      '</div>';

    var btn = document.getElementById('a-submit');
    var doSubmit = function () {
      var username = document.getElementById('a-username').value.trim();
      var password = document.getElementById('a-password').value;
      if (!username || !password) return toast('请填写用户名和密码', 'err');
      if (isRegister) {
        var email = document.getElementById('a-email').value.trim();
        if (!email) return toast('请填写邮箱', 'err');
        api('/auth/register', { method: 'POST', body: { username: username, password: password, email: email } })
          .then(function () {
            toast('注册成功，正在登录...', 'ok');
            return api('/auth/login', { method: 'POST', body: { username: username, password: password } });
          })
          .then(afterLogin).catch(function (e) { toast(e.message, 'err'); });
      } else {
        api('/auth/login', { method: 'POST', body: { username: username, password: password } })
          .then(afterLogin).catch(function (e) { toast(e.message, 'err'); });
      }
    };
    btn.addEventListener('click', doSubmit);
    document.querySelectorAll('.auth-card input').forEach(function (inp) {
      inp.addEventListener('keydown', function (e) { if (e.key === 'Enter') doSubmit(); });
    });
  }
  function afterLogin(data) {
    token = data.token;
    localStorage.setItem(TOKEN_KEY, token);
    setUser(data.user);
    toast('欢迎回来，' + data.user.username + '！', 'ok');
    location.hash = '#/';
  }

  /* ---------- 启动 ---------- */
  updateNav();
  render();
})();
