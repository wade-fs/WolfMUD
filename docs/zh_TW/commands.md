# 摘要

- 參考 core/command.go
- 最重要的函式 RegisterCommandHandlers()
- 3 個重要的資料:
	- commandHandlers
	- commandNames
	- eventCommands

# commandHandlers  
	目前不知道 $ # 開頭的命令怎麼用, 大抵上是 / 非角色專用，# 管理命令, $ script 用
命令			| 說明
---------------:|:---------------
CMD HELP		| 顯示命令清單
COMMANDS		| 同上，在想怎樣預設系統 alias
QUIT			| 離開 WolfMUD
L LOOK			| 看：  顯示所在房間描述
N NORTH			| 北：  往北走
NE NORTHEAST	| 東北：往東北走
E EAST			| 東：  往東走
SE SOUTHEAST	| 東南：往東南走
S SOUTH			| 南：  往南走
SW SOUTHWEST	| 西南：往西南走
W WEST			| 西：  往西走
NW NORTHWEST	| 西北：往西北走
U UP			| 上：  往上走
D DOWN			| 下：  往下走
EXAM EXAMINE	| 檢查：檢查物品
INV INVENTORY	| 顯示庫存清單
DROP			| 丟：丟棄庫存中的物品在地上
GET				| 拿：從房間中取得物品
TAKE			| 取：從容器（房間、屍體、袋子等）中取得物品
PUT				| 放：將物品放到容器（房間、屍體、袋子等）中
READ			| 讀」：讀書(Writing)
OPEN			| 打開：打開物品(Blocker)
CLOSE			| 關上：關閉物品(Blocker)
" SAY			| 說：說話，旁邊要有人，不然會變喃喃自語，僅限 radius(1) 範圍內的人
SNEEZE			| 噴嚏(只限 radius(2) 範圍內的人
SHOUT			| 大喊(只限 radius(2) 範圍內的人
JUNK			| 丟垃圾：類似 DROP, 對應的是會消失
REMOVE			| 移除：停用設備，必須非拿非裝備非穿著
HOLD			| 拿著：不能拿著穿著、裝備的物品
WEAR			| 穿著：將衣物穿在身上
WIELD			| 裝備：將武器裝備著，可以用來攻擊，純拿著無法攻擊
VERSION			| 顯示版本
SAVE			| 儲存資料，離開 WolfMUD 會自動儲存
KILL			| 攻擊活物，會展開撕殺
ATTACK			| 同上
HIT				| 同上
TELL			| 對某人說話，對 radius(1) 範圍的人也會聽到
TALK			| 同上
WHISPER			| 耳語, 類似 Tell, 但是別人聽不到或聽不清楚
/WHO /WHOAMI	| 
/HISTORY /!		| 
#DUMP			| 管理員命令，傾印物品內部訊息
#LDUMP 			| 同上
#TELEPORT		| 管理員命令，傳送到某容器
#GOTO			| 同上
#DEBUG			| 管理員命令，切換除錯狀態: CPUPROF(ON/OFF) | MEMPROF(ON/OFF) | PANIC
#EVAL			| 戰鬥攻防資訊, TODO: 不需要管理員權限？
$POOF			| 
$ACT			| 表演，類似說話的功效，只要不是太擁擠的話，房間裡每個人都看得到訊息
$ACTION			| 同上
$RESET			| 
$CLEANUP		| 
$TRIGGER		| 
$QUIT			| 
$HEALTH			| 
$COMBAT			| 

# eventCommands
	目前內容有
- Action:  "$ACTION",
- Reset:   "$RESET",
- Cleanup: "$CLEANUP",
- Trigger: "$TRIGGER",
- Health:  "$HEALTH",
- Combat:  "$COMBAT",

# commandNames  
	包含 commandHandlers 中，非 $ 開頭的命令清單
