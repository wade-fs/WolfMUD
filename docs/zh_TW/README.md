# 說明
- 純粹做為研究用，
- 先做好翻譯工作，因為英文太差

# 想法
- 賦與管理帳號的方式建議修改為，
	- 玩家登入，
	- 提出申請
	- 賦與權限
	- 必要時玩家登出
	- 玩家登入
- redirect to web by use [sshwifty][1]
- I18N
- 第一個角色自動為 admin
	- 管理者可以讀寫檔案
- Script by using [Mere][2]

# clone to github
- git clone ttps://code.wolfmud.org/WolfMUD
- cd WolfMUD; git co -b dev origin/dev
- git remote add github https://github.com/wade-fs/WolfMUD
- git fetch github
- git co Docs

[1]: https://github.com/nirui/sshwifty
[2]: https://www.wolfmud.org/annex/mere/ice.html
