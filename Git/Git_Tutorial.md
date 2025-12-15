# Git 完整教程 - 新手入门指南

## 目录
1. [什么是 Git？](#什么是-git)
2. [Git 的安装和配置](#git-的安装和配置)
3. [Git 基本概念](#git-基本概念)
4. [Git 基本操作](#git-基本操作)
5. [分支管理](#分支管理)
6. [远程仓库](#远程仓库)
7. [常见场景和问题解决](#常见场景和问题解决)
8. [最佳实践](#最佳实践)
9. [常用命令速查表](#常用命令速查表)

---

## 什么是 Git？

### Git 的定义
Git 是一个**分布式版本控制系统**（Distributed Version Control System，简称 DVCS）。它可以帮助你：
- 跟踪文件的变化历史
- 多人协作开发
- 回退到任意历史版本
- 管理代码的不同版本（分支）

### 为什么需要 Git？
想象一下，你在写一个重要的文档或代码：
- 你修改了很多次，突然发现之前的版本更好，怎么办？
- 你和同事同时修改同一个文件，如何合并？
- 你想尝试一个新功能，但又不想影响现有代码，怎么办？

Git 就是为了解决这些问题而生的。

### Git vs 其他版本控制系统
- **集中式版本控制系统**（如 SVN）：所有代码存储在一个中央服务器上
- **分布式版本控制系统**（Git）：每个开发者都有完整的代码历史副本

Git 的优势：
- 可以离线工作
- 速度快
- 分支操作简单高效
- 强大的合并能力

---

## Git 的安装和配置

### Windows 系统安装

#### 方法一：官方安装包
1. 访问 [Git 官网](https://git-scm.com/download/win)
2. 下载 Windows 安装程序
3. 运行安装程序，使用默认选项即可
4. 安装完成后，打开 **Git Bash** 或 **PowerShell** 验证：
```bash
git --version
```

#### 方法二：使用包管理器
**使用 Chocolatey：**
```bash
choco install git
```

**使用 Winget：**
```bash
winget install Git.Git
```

### macOS 系统安装

#### 方法一：使用 Homebrew
```bash
brew install git
```

#### 方法二：安装 Xcode Command Line Tools
```bash
xcode-select --install
```

### Linux 系统安装

**Ubuntu/Debian：**
```bash
sudo apt update
sudo apt install git
```

**CentOS/RHEL：**
```bash
sudo yum install git
```

**Fedora：**
```bash
sudo dnf install git
```

### 首次配置 Git

安装完成后，需要配置你的身份信息：

```bash
# 设置用户名（使用你的真实姓名或 GitHub 用户名）
git config --global user.name "Your Name"

# 设置邮箱（使用你的 GitHub 邮箱）
git config --global user.email "your.email@example.com"

# 查看配置
git config --list

# 查看特定配置
git config user.name
git config user.email
```

### 其他有用的配置

```bash
# 设置默认编辑器（可选）
git config --global core.editor "code --wait"  # VS Code
git config --global core.editor "vim"          # Vim
git config --global core.editor "nano"         # Nano

# 设置默认分支名（Git 2.28+）
git config --global init.defaultBranch main

# 启用颜色输出
git config --global color.ui auto

# 设置换行符处理（Windows 推荐）
git config --global core.autocrlf true

# 设置换行符处理（Linux/macOS 推荐）
git config --global core.autocrlf input
```

---

## Git 基本概念

### 工作区、暂存区、仓库

理解这三个区域是掌握 Git 的关键：

```
┌─────────────┐      git add      ┌─────────────┐      git commit     ┌─────────────┐
│  工作区      │ ───────────────> │  暂存区      │ ───────────────> │  本地仓库    │
│ (Working    │                  │ (Staging    │                  │ (Repository)│
│  Directory) │                  │  Area)      │                  │             │
└─────────────┘                  └─────────────┘                  └─────────────┘
```

1. **工作区（Working Directory）**
   - 就是你电脑上的文件夹
   - 你在这里编辑文件

2. **暂存区（Staging Area / Index）**
   - 准备提交的文件临时存放的地方
   - 使用 `git add` 将文件添加到暂存区

3. **本地仓库（Repository）**
   - Git 存储所有版本历史的地方
   - 使用 `git commit` 将暂存区的文件提交到仓库

### 文件状态

Git 中的文件有四种状态：

1. **未跟踪（Untracked）**
   - 新创建的文件，Git 还没有开始跟踪
   - 使用 `git add` 开始跟踪

2. **已修改（Modified）**
   - 文件被修改了，但还没有添加到暂存区

3. **已暂存（Staged）**
   - 文件已修改并添加到暂存区，准备提交

4. **已提交（Committed）**
   - 文件已安全地保存在本地仓库中

### 提交（Commit）

提交是 Git 的核心概念：
- 每次提交都会创建一个**快照**（Snapshot）
- 每个提交都有一个唯一的 **SHA-1 哈希值**（如：`a1b2c3d...`）
- 提交包含：提交信息、作者、时间戳、文件变化

---

## Git 基本操作

### 1. 初始化仓库

#### 创建新仓库
```bash
# 在项目文件夹中初始化 Git 仓库
mkdir my-project
cd my-project
git init

# 这会创建一个隐藏的 .git 文件夹，存储所有 Git 数据
```

#### 克隆现有仓库
```bash
# 从远程仓库克隆（下载）项目
git clone https://github.com/username/repository.git

# 克隆到指定文件夹
git clone https://github.com/username/repository.git my-folder

# 克隆特定分支
git clone -b branch-name https://github.com/username/repository.git
```

### 2. 查看状态

```bash
# 查看工作区和暂存区的状态
git status

# 简短格式
git status -s
git status --short

# 查看更详细的信息
git status -v
```

**输出示例：**
```
On branch main
Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
  (use "git restore <file>..." to discard changes in working directory)
        modified:   README.md

Untracked files:
  (use "git add <file>..." to include in what will be committed)
        new-file.txt

Changes staged for commit:
  (use "git restore --staged <file>..." to unstage)
        modified:   config.json
```

### 3. 添加文件到暂存区

```bash
# 添加单个文件
git add filename.txt

# 添加多个文件
git add file1.txt file2.txt file3.txt

# 添加当前目录下所有文件
git add .

# 添加所有 .txt 文件
git add *.txt

# 添加所有修改过的文件（不包括新文件）
git add -u
git add --update

# 交互式添加（可以选择性地添加文件的某些部分）
git add -i
git add --interactive

# 添加所有文件（包括被删除的文件）
git add -A
git add --all
```

### 4. 提交更改

```bash
# 提交暂存区的所有更改
git commit -m "Add new feature"

# 提交信息应该清晰描述这次更改的内容
# 好的提交信息示例：
git commit -m "Fix login bug"
git commit -m "Add user authentication"
git commit -m "Update documentation"

# 打开编辑器编写详细的提交信息
git commit

# 直接提交所有已跟踪的修改（跳过 git add）
git commit -a -m "Update files"
git commit -am "Update files"

# 修改最后一次提交（如果还没有推送）
git commit --amend -m "New commit message"

# 修改最后一次提交并包含新的更改
git add forgotten-file.txt
git commit --amend --no-edit
```

### 5. 查看提交历史

```bash
# 查看提交历史
git log

# 单行显示
git log --oneline

# 图形化显示分支
git log --oneline --graph --all

# 显示最近 5 次提交
git log -5

# 显示文件变化统计
git log --stat

# 显示详细的文件变化
git log -p

# 搜索提交信息
git log --grep="bug fix"

# 按作者搜索
git log --author="John"

# 按日期范围
git log --since="2024-01-01" --until="2024-12-31"

# 查看特定文件的提交历史
git log filename.txt
```

### 6. 查看文件差异

```bash
# 查看工作区和暂存区的差异
git diff

# 查看暂存区和最后一次提交的差异
git diff --staged
git diff --cached

# 查看两次提交之间的差异
git diff commit1 commit2

# 查看特定文件的差异
git diff filename.txt

# 查看两个分支之间的差异
git diff branch1 branch2

# 统计差异
git diff --stat
```

### 7. 撤销更改

#### 撤销工作区的修改
```bash
# 撤销文件修改（危险：会丢失更改）
git restore filename.txt
# 或者（旧版本 Git）
git checkout -- filename.txt

# 撤销所有工作区的修改
git restore .
```

#### 取消暂存
```bash
# 将文件从暂存区移除（但保留工作区的修改）
git restore --staged filename.txt
# 或者
git reset HEAD filename.txt

# 取消所有暂存
git restore --staged .
```

#### 撤销提交
```bash
# 撤销最后一次提交，但保留文件更改
git reset --soft HEAD~1

# 撤销最后一次提交和暂存，但保留工作区更改
git reset --mixed HEAD~1
git reset HEAD~1

# 完全撤销最后一次提交（危险：会丢失更改）
git reset --hard HEAD~1

# 撤销到指定提交
git reset --hard commit-hash
```

### 8. 删除文件

```bash
# 从 Git 中删除文件（同时删除工作区的文件）
git rm filename.txt
git commit -m "Remove filename.txt"

# 只从 Git 中删除，保留工作区文件
git rm --cached filename.txt
git commit -m "Stop tracking filename.txt"

# 删除文件夹
git rm -r folder-name
```

### 9. 重命名/移动文件

```bash
# 重命名文件
git mv old-name.txt new-name.txt
git commit -m "Rename file"

# 移动文件
git mv file.txt folder/file.txt
git commit -m "Move file to folder"

# Git 会自动检测重命名和移动
```

---

## 分支管理

分支是 Git 最强大的功能之一，允许你在不影响主代码的情况下进行开发。

### 什么是分支？

分支就像一条时间线的分叉：
- 默认分支通常是 `main` 或 `master`
- 你可以创建新分支来开发新功能
- 完成后可以合并回主分支

```
main:     A---B---C---F---G
                \
feature:         D---E
```

### 基本分支操作

#### 查看分支
```bash
# 查看所有本地分支
git branch

# 查看所有分支（包括远程）
git branch -a

# 查看远程分支
git branch -r

# 查看分支的详细信息
git branch -v
```

#### 创建分支
```bash
# 创建新分支
git branch branch-name

# 创建并切换到新分支
git checkout -b branch-name
# 或者（Git 2.23+）
git switch -c branch-name

# 基于指定提交创建分支
git branch branch-name commit-hash

# 基于远程分支创建本地分支
git branch branch-name origin/branch-name
```

#### 切换分支
```bash
# 切换到指定分支
git checkout branch-name
# 或者（Git 2.23+）
git switch branch-name

# 切换到上一个分支
git checkout -
git switch -
```

#### 删除分支
```bash
# 删除已合并的分支
git branch -d branch-name

# 强制删除分支（即使未合并）
git branch -D branch-name

# 删除远程分支
git push origin --delete branch-name
```

### 合并分支

#### 基本合并
```bash
# 切换到目标分支（通常是 main）
git checkout main

# 合并指定分支
git merge branch-name

# 合并后删除已合并的分支
git branch -d branch-name
```

#### 合并类型

1. **Fast-forward 合并**
   - 当目标分支没有新的提交时
   - Git 只是移动指针，不创建合并提交

2. **三方合并（3-way merge）**
   - 当两个分支都有新提交时
   - Git 创建一个新的合并提交

3. **合并冲突**
   - 当两个分支修改了同一文件的同一部分时
   - 需要手动解决冲突

#### 解决合并冲突

当发生冲突时，Git 会标记冲突：

```bash
# 1. 查看冲突文件
git status

# 2. 打开冲突文件，会看到类似这样的标记：
<<<<<<< HEAD
当前分支的代码
=======
要合并分支的代码
>>>>>>> branch-name

# 3. 手动编辑文件，选择保留的代码，删除冲突标记

# 4. 标记冲突已解决
git add resolved-file.txt

# 5. 完成合并
git commit
```

**冲突解决示例：**

冲突前：
```
<<<<<<< HEAD
function calculate(a, b) {
    return a + b;
}
=======
function calculate(x, y) {
    return x * y;
}
>>>>>>> feature-branch
```

解决后（选择其中一个或合并两者）：
```
function calculate(a, b) {
    return a + b;  // 使用加法版本
}
```

#### 使用合并工具
```bash
# 配置合并工具
git config --global merge.tool vimdiff
git config --global merge.tool vscode

# 启动合并工具
git mergetool
```

### 变基（Rebase）

变基是另一种整合分支的方法，它会让提交历史更线性：

```bash
# 将当前分支变基到目标分支
git checkout feature-branch
git rebase main

# 交互式变基（可以修改、删除、合并提交）
git rebase -i HEAD~3  # 修改最近 3 次提交

# 中止变基
git rebase --abort

# 继续变基（解决冲突后）
git rebase --continue
```

**变基 vs 合并：**
- **合并**：保留完整的分支历史，创建合并提交
- **变基**：创建线性的提交历史，看起来更整洁

**注意：** 不要对已推送到远程的公共分支进行变基！

---

## 远程仓库

远程仓库是托管在服务器上的 Git 仓库（如 GitHub、GitLab、Bitbucket）。

### 基本概念

- **origin**：默认的远程仓库名称
- **fetch**：从远程下载更改，但不合并
- **pull**：从远程下载并合并更改
- **push**：将本地更改上传到远程

### 远程仓库操作

#### 查看远程仓库
```bash
# 查看所有远程仓库
git remote

# 查看详细信息
git remote -v

# 查看远程仓库的详细信息
git remote show origin
```

#### 添加远程仓库
```bash
# 添加远程仓库
git remote add origin https://github.com/username/repo.git

# 添加多个远程仓库
git remote add upstream https://github.com/original/repo.git
```

#### 修改远程仓库
```bash
# 修改远程仓库 URL
git remote set-url origin https://github.com/username/new-repo.git

# 重命名远程仓库
git remote rename origin upstream

# 删除远程仓库
git remote remove origin
```

### 获取和拉取

#### Fetch（获取）
```bash
# 从远程获取所有更新（不合并）
git fetch origin

# 获取特定分支
git fetch origin branch-name

# 获取所有远程仓库
git fetch --all
```

#### Pull（拉取）
```bash
# 从远程拉取并合并到当前分支
git pull origin main

# 拉取并变基（而不是合并）
git pull --rebase origin main

# 简写（如果已设置上游分支）
git pull
```

### 推送

```bash
# 推送到远程仓库
git push origin main

# 推送所有分支
git push --all origin

# 推送标签
git push origin --tags

# 强制推送（危险：会覆盖远程历史）
git push --force origin main
# 更安全的强制推送
git push --force-with-lease origin main

# 设置上游分支（之后可以直接用 git push）
git push -u origin main
```

### 标签（Tags）

标签用于标记重要的提交（如版本发布）：

```bash
# 创建轻量标签
git tag v1.0.0

# 创建附注标签（推荐）
git tag -a v1.0.0 -m "Release version 1.0.0"

# 在指定提交上创建标签
git tag -a v1.0.0 commit-hash -m "Version 1.0.0"

# 查看所有标签
git tag

# 查看标签信息
git show v1.0.0

# 删除标签
git tag -d v1.0.0

# 删除远程标签
git push origin --delete v1.0.0

# 推送所有标签
git push origin --tags

# 推送特定标签
git push origin v1.0.0
```

---

## 常见场景和问题解决

### 场景 1：误提交了敏感信息

```bash
# 如果还没有推送
git reset --soft HEAD~1
# 修改文件，移除敏感信息
git add .
git commit -m "Update files"

# 如果已经推送，需要使用 git filter-branch 或 BFG Repo-Cleaner
# 然后强制推送（需要团队协调）
```

### 场景 2：想撤销最后一次提交但保留更改

```bash
git reset --soft HEAD~1
# 现在更改在暂存区，可以修改后重新提交
```

### 场景 3：想撤销工作区的所有更改

```bash
# 查看会删除哪些文件
git clean -n

# 删除未跟踪的文件
git clean -f

# 删除未跟踪的文件和文件夹
git clean -fd
```

### 场景 4：提交信息写错了

```bash
# 修改最后一次提交信息
git commit --amend -m "Correct commit message"
```

### 场景 5：忘记添加文件到上次提交

```bash
git add forgotten-file.txt
git commit --amend --no-edit
```

### 场景 6：想查看某个文件的修改历史

```bash
# 查看文件的提交历史
git log filename.txt

# 查看文件的具体变化
git log -p filename.txt

# 查看谁修改了文件的哪一行
git blame filename.txt
```

### 场景 7：想临时保存工作进度

```bash
# 保存当前工作（包括未暂存的更改）
git stash

# 保存并添加描述
git stash save "Work in progress on feature X"

# 查看所有 stash
git stash list

# 恢复最近的 stash
git stash pop

# 恢复指定的 stash（不删除）
git stash apply stash@{0}

# 删除 stash
git stash drop stash@{0}

# 清空所有 stash
git stash clear
```

### 场景 8：想回到之前的某个版本

```bash
# 查看提交历史，找到目标提交的 hash
git log --oneline

# 临时查看（不修改工作区）
git checkout commit-hash

# 创建新分支基于旧提交
git checkout -b new-branch commit-hash

# 永久回退（危险）
git reset --hard commit-hash
```

### 场景 9：合并时发生冲突

```bash
# 1. 查看冲突文件
git status

# 2. 打开冲突文件，手动解决
# 3. 标记为已解决
git add resolved-file.txt

# 4. 完成合并
git commit

# 如果想取消合并
git merge --abort
```

### 场景 10：想查看两个版本之间的差异

```bash
# 查看两个提交之间的差异
git diff commit1 commit2

# 查看两个分支之间的差异
git diff branch1 branch2

# 查看特定文件的差异
git diff commit1 commit2 -- filename.txt
```

### 场景 11：想找到引入 bug 的提交

```bash
# 使用二分查找
git bisect start
git bisect bad  # 标记当前版本有问题
git bisect good commit-hash  # 标记某个旧版本没问题
# Git 会自动切换到中间版本，你测试后标记 good 或 bad
# 重复直到找到问题提交
git bisect reset  # 结束二分查找
```

### 场景 12：想重写提交历史

```bash
# 交互式变基（修改最近 3 次提交）
git rebase -i HEAD~3

# 在编辑器中：
# pick -> 保留提交
# reword -> 修改提交信息
# edit -> 修改提交内容
# squash -> 合并到上一个提交
# drop -> 删除提交
```

---

## 最佳实践

### 1. 提交信息规范

好的提交信息应该：
- 清晰描述做了什么
- 使用祈使语气（如 "Add feature" 而不是 "Added feature"）
- 第一行不超过 50 个字符
- 如果需要，添加详细描述

**示例：**
```
Add user authentication

- Implement login functionality
- Add password hashing
- Create user session management
```

### 2. 提交频率

- **频繁提交**：每完成一个小功能就提交
- **有意义的提交**：每次提交应该是一个完整的工作单元
- **不要提交**：临时文件、编译产物、敏感信息

### 3. .gitignore 文件

创建 `.gitignore` 文件来排除不需要版本控制的文件：

```gitignore
# 编译产物
*.class
*.o
*.exe
*.dll

# 依赖文件夹
node_modules/
vendor/
venv/

# IDE 文件
.vscode/
.idea/
*.swp
*.swo

# 操作系统文件
.DS_Store
Thumbs.db

# 日志文件
*.log

# 环境变量文件
.env
.env.local

# 临时文件
*.tmp
*.temp
```

### 4. 分支命名规范

- `feature/` - 新功能
- `bugfix/` 或 `fix/` - 修复 bug
- `hotfix/` - 紧急修复
- `release/` - 发布准备
- `develop` - 开发分支
- `main` 或 `master` - 主分支

**示例：**
```
feature/user-authentication
bugfix/login-error
hotfix/security-patch
release/v1.0.0
```

### 5. 工作流程

#### 功能开发流程
```bash
# 1. 从主分支创建功能分支
git checkout main
git pull origin main
git checkout -b feature/new-feature

# 2. 开发功能，频繁提交
git add .
git commit -m "Implement feature X"

# 3. 推送到远程
git push -u origin feature/new-feature

# 4. 创建 Pull Request / Merge Request

# 5. 代码审查后合并到主分支

# 6. 删除功能分支
git checkout main
git pull origin main
git branch -d feature/new-feature
```

### 6. 代码审查

- 提交 Pull Request 前先自己审查
- 保持 PR 小而专注
- 添加清晰的描述
- 回应审查意见

### 7. 保护主分支

- 不要直接在主分支上开发
- 使用 Pull Request 合并代码
- 要求代码审查
- 运行自动化测试

### 8. 定期同步

```bash
# 每天开始工作前
git checkout main
git pull origin main

# 在功能分支上定期同步主分支
git checkout feature/my-feature
git merge main
# 或
git rebase main
```

---

## 常用命令速查表

### 基础命令

| 命令 | 说明 |
|------|------|
| `git init` | 初始化仓库 |
| `git clone <url>` | 克隆远程仓库 |
| `git status` | 查看状态 |
| `git add <file>` | 添加文件到暂存区 |
| `git commit -m "msg"` | 提交更改 |
| `git log` | 查看提交历史 |
| `git diff` | 查看差异 |

### 分支命令

| 命令 | 说明 |
|------|------|
| `git branch` | 查看分支 |
| `git branch <name>` | 创建分支 |
| `git checkout <name>` | 切换分支 |
| `git checkout -b <name>` | 创建并切换分支 |
| `git merge <branch>` | 合并分支 |
| `git branch -d <name>` | 删除分支 |

### 远程命令

| 命令 | 说明 |
|------|------|
| `git remote -v` | 查看远程仓库 |
| `git fetch` | 获取远程更新 |
| `git pull` | 拉取并合并 |
| `git push` | 推送到远程 |

### 撤销命令

| 命令 | 说明 |
|------|------|
| `git restore <file>` | 撤销工作区修改 |
| `git restore --staged <file>` | 取消暂存 |
| `git reset HEAD~1` | 撤销提交 |
| `git commit --amend` | 修改最后一次提交 |

### 其他有用命令

| 命令 | 说明 |
|------|------|
| `git stash` | 临时保存工作 |
| `git tag` | 管理标签 |
| `git blame <file>` | 查看文件修改者 |
| `git bisect` | 二分查找问题 |

---

## 学习资源

### 官方资源
- [Git 官网](https://git-scm.com/)
- [Git 官方文档](https://git-scm.com/doc)
- [Pro Git 电子书](https://git-scm.com/book)（免费，强烈推荐）

### 可视化工具
- **GitHub Desktop**：图形化 Git 客户端
- **SourceTree**：免费的 Git GUI 工具
- **VS Code Git 扩展**：在编辑器中直接使用 Git

### 在线练习
- [Learn Git Branching](https://learngitbranching.js.org/)：交互式 Git 学习
- [GitHub Learning Lab](https://lab.github.com/)：GitHub 官方教程

### 常见问题
- [Git 常见问题](https://git-scm.com/docs/gitfaq)
- [GitHub Help](https://help.github.com/)

---

## 总结

Git 是一个强大的工具，掌握它需要时间和实践。建议：

1. **多练习**：在实际项目中使用 Git
2. **从小开始**：先掌握基本命令，再学习高级功能
3. **不要害怕**：Git 有撤销机制，大部分操作可以恢复
4. **阅读文档**：遇到问题先查看官方文档
5. **使用图形工具**：可视化工具可以帮助理解概念

记住：**Git 是用来帮助你的，不是来为难你的！**

祝你学习愉快！🚀

---

*最后更新：2024年*
*如有问题或建议，欢迎反馈！*

