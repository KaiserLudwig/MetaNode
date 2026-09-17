#!/usr/bin/env python3
"""当 github.com 直连超时（SSL/连接超时）时，改用 GitHub API 把本地 HEAD 推到远端。

原理：本地 HEAD 与远端父提交一致时，用 Git Data API 重建一模一样的 tree/commit
（作者、提交时间、commit message 全部沿用本地），因此远端提交哈希与本地完全相同，
不会产生分叉，等直连恢复后照常 `git push` 即可。

用法：
    python3 scripts/git_push_api.py [仓库路径] [owner/repo]

Token 读取顺序：环境变量 GITHUB_TOKEN，或 ../.secrets/github_token.env 文件。
"""

import base64
import json
import os
import subprocess
import sys
import urllib.error
import urllib.request


def load_token(start_dir: str) -> str:
    token = os.environ.get("GITHUB_TOKEN")
    if token:
        return token.strip()

    probe = os.path.abspath(start_dir)
    for _ in range(5):
        candidate = os.path.join(probe, ".secrets", "github_token.env")
        if os.path.isfile(candidate):
            for line in open(candidate, encoding="utf-8"):
                if line.startswith("GITHUB_TOKEN="):
                    return line.split("=", 1)[1].strip()
        probe = os.path.dirname(probe)
    raise SystemExit("找不到 GITHUB_TOKEN（环境变量或 .secrets/github_token.env）")


def main() -> None:
    repo_dir = os.path.abspath(sys.argv[1] if len(sys.argv) > 1 else ".")
    owner_repo = sys.argv[2] if len(sys.argv) > 2 else "KaiserLudwig/MetaNode"
    token = load_token(repo_dir)
    api = "https://api.github.com"

    def git(*args: str) -> str:
        return subprocess.run(
            ["git", *args], cwd=repo_dir, capture_output=True, text=True, check=True
        ).stdout.strip()

    def call(method: str, path: str, payload=None):
        data = json.dumps(payload).encode() if payload is not None else None
        request = urllib.request.Request(
            api + path,
            data=data,
            method=method,
            headers={
                "Authorization": "token " + token,
                "Accept": "application/vnd.github+json",
                "Content-Type": "application/json",
                "User-Agent": "metanode-push",
            },
        )
        try:
            with urllib.request.urlopen(request, timeout=60) as response:
                return json.load(response)
        except urllib.error.HTTPError as error:
            raise SystemExit(f"GitHub API {method} {path} 失败: {error.code} {error.read().decode()[:300]}")

    branch = git("rev-parse", "--abbrev-ref", "HEAD")
    head = git("rev-parse", "HEAD")
    tree_sha = git("rev-parse", "HEAD^{tree}")
    parent = git("rev-parse", "HEAD~1")

    raw_commit = subprocess.run(
        ["git", "cat-file", "commit", "HEAD"], cwd=repo_dir, capture_output=True, check=True
    ).stdout
    message = raw_commit.split(b"\n\n", 1)[1].decode("utf-8")

    meta = git("log", "-1", "--format=%an%x00%ae%x00%aI%x00%cn%x00%ce%x00%cI").split("\x00")
    author_name, author_email, author_date, committer_name, committer_email, committer_date = meta

    remote_base = call("GET", f"/repos/{owner_repo}/git/commits/{parent}")
    remote_base_sha = remote_base["sha"]
    if remote_base_sha != parent:
        raise SystemExit(f"远端父提交({remote_base_sha[:12]})与本地父提交({parent[:12]})不一致，请先正常 fetch/push")

    changed = [line for line in git("diff", "--name-status", "HEAD~1", "HEAD").splitlines() if line]
    if not changed:
        print("本地 HEAD 与父提交没有差异，无需推送")
        return

    entries = []
    for line in changed:
        status, path = line.split("\t", 1)
        if status.startswith("D"):
            entries.append({"path": path, "mode": "100644", "type": "blob", "sha": None})
            continue
        blob = call(
            "POST",
            f"/repos/{owner_repo}/git/blobs",
            {
                "content": base64.b64encode(open(os.path.join(repo_dir, path), "rb").read()).decode(),
                "encoding": "base64",
            },
        )
        mode = git("ls-tree", "HEAD", path).split()[0]
        entries.append({"path": path, "mode": mode, "type": "blob", "sha": blob["sha"]})

    new_tree = call(
        "POST", f"/repos/{owner_repo}/git/trees", {"base_tree": remote_base["tree"]["sha"], "tree": entries}
    )
    if new_tree["sha"] != tree_sha:
        raise SystemExit(f"tree 不一致：本地 {tree_sha[:12]} vs 远端将创建 {new_tree['sha'][:12]}")

    new_commit = call(
        "POST",
        f"/repos/{owner_repo}/git/commits",
        {
            "message": message,
            "tree": new_tree["sha"],
            "parents": [parent],
            "author": {"name": author_name, "email": author_email, "date": author_date},
            "committer": {"name": committer_name, "email": committer_email, "date": committer_date},
        },
    )
    if new_commit["sha"] != head:
        raise SystemExit(f"提交哈希不一致：本地 {head[:12]} vs 远端将创建 {new_commit['sha'][:12]}")

    call("PATCH", f"/repos/{owner_repo}/git/refs/heads/{branch}", {"sha": head, "force": False})
    subprocess.run(["git", "update-ref", f"refs/remotes/origin/{branch}", head], cwd=repo_dir, check=True)
    print(f"已推送 {branch} -> {owner_repo}：{head[:12]} {message.splitlines()[0]}")


if __name__ == "__main__":
    main()
