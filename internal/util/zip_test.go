package util

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func init() {
	var cstZone = time.FixedZone("UTC", 8*3600) // 东八
	time.Local = cstZone
}

func TestZipFile(t *testing.T) {

	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
	time.Local = time.FixedZone("UTC", 2*3600)
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))

	dst := "/Users/zgia/Desktop/xinyue.zip"
	src := "/Users/zgia/Desktop/欣乐文档"

	if err := ZipFile(src, dst); err != nil {
		panic(fmt.Sprintf("Cannot zip file , %s", err.Error()))
	}

	t.Run("", func(t *testing.T) {
		assert.Equal(t, 1, 1)
	})
}

func TestStringsIndex(t *testing.T) {
	str := txt()

	content := make([]string, 1)

	word := "克莱恩"
	length := len([]rune(word))

	for {
		fmt.Println(utf8.RuneCountInString(str))

		idx, _ := runeIndex(str, word)
		fmt.Printf("index : %d\n", idx)
		start := 0
		if idx > 5 {
			start = idx - 5
		}

		if idx != -1 {
			content = append(content, substr(str, start, length+10))

			str = substr(str, idx+length, len([]rune(str))-idx)
		} else {
			break
		}
	}

	for index, value := range content {
		fmt.Printf("inde : %v , value : %v\n", index, value)
	}
}

func runeIndex(s, substr string) (int, error) {
	byteIndex := strings.Index(s, substr)
	if byteIndex < 0 {
		return byteIndex, nil
	}
	reader := strings.NewReader(s)
	count := 0
	for byteIndex > 0 {
		_, bytes, err := reader.ReadRune()
		if err != nil {
			return 0, err
		}
		byteIndex = byteIndex - bytes
		count += 1
	}
	return count, nil
}

func substr(input string, start int, length int) string {
	asRunes := []rune(input)

	if start >= len(asRunes) {
		return ""
	}

	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}

	return string(asRunes[start : start+length])
}

func txt() string {
	return "　　妹，咱能不能不要哪壶不开提哪壶……克莱恩暗自吐槽，只觉脑袋又开始一抽一抽地痛。　　原主遗忘的知识不算多，也绝对不少，后天就要面试了，哪有时间补得上来……　　而且还卷入了诡异恐怖的事件，怎么可能有心思去“复习”……　　敷衍了妹妹几句，克莱恩开始装模作样读书，梅丽莎搬了椅子，坐在旁边，借着煤气灯的光芒做起了作业。　　气氛宁静安乐，快十一点时，兄妹互道晚安，各自上床。　　……　　咚！　　咚咚！　　一阵敲门声响起，克莱恩从梦中醒来。　　他看了眼窗外的晨曦，脑袋略显迷糊地翻身坐起：　　“谁啊？”　　这都几点了？梅丽莎怎么没叫醒我？　　“我，邓恩·史密斯。”门外有沉稳的男声回答。　　邓恩·史密斯？不认识……克莱恩摇头下床，走向门边。　　他拉开房门，看见了昨天那位有着灰色眼眸的警官。　　“出什么事了吗？”克莱恩警惕问道。　　灰眸警官表情严肃地回答：　　“我们找到了一个马车夫，他证实你在27日，也就是韦尔奇先生和娜娅女士死亡的当天，去过韦尔奇先生的住所，而且还是韦尔奇先生帮你付的车钱。”　　克莱恩怔了一下，丝毫没有谎言被揭穿的惊恐和心虚。　　因为他根本不是在撒谎，反倒感觉灰眸警官邓恩·史密斯提供的证据不出自身的预料。　　6月27日那天，原主果然还是去了韦尔奇的住所，回来的当天夜里就自杀身亡，和韦尔奇、娜娅一模一样！　　克莱恩张了张嘴，泛起一抹苦笑道：　　“这不是足够有力的证据，不能直接证明我和韦尔奇、娜娅的死亡有关，老实说，我也很想知道事情的经过，弄清楚我两位可怜朋友的遭遇，但是，但是，我真的记不得了，我几乎完全遗忘了27号那天做过的事情，说出来你可能不相信，我全靠我自己的笔记才勉强猜到我27号也许去过韦尔奇的住所。”　　“心理素质不错。”灰眸警官邓恩·史密斯不见愤怒也不见微笑地点了点头。　　“你应该能听得出我的诚恳。”克莱恩直视着对方的双眼。　　我说的都是真话，当然，只是真话的其中一部分！　　邓恩·史密斯没有立刻回应，视线扫了房间一圈才慢悠悠道：　　“韦尔奇先生丢失了一把左轮手枪，我想我应该能在这里找到他，对吧，克莱恩先生？”　　果然……克莱恩总算弄清楚了左轮手枪的来历，脑海念头如闪电跳跃般转动，瞬间做出了决断。　　他半举起双手，一步步退后，让开了道路，然后用下巴指向高低床道：　　“在床板背面。”　　他没具体说是下面那张，因为正常人都不会把东西藏在上层床板的背面，那会让访客一目了然地看到。　　灰眸警官邓恩没有往前，抽了下嘴角道：　　“没什么想要补充的吗？”　　克莱恩毫不犹豫地回答：　　“有！”　　“前晚半夜醒来，我发现自己趴在书桌上，旁边是左轮手枪，墙脚有子弹，看起来像是经历了一场自杀，只是也许没经验，没用过手枪，或者最后关头害怕了，总之，子弹没达到预想的效果，我的脑袋还完好，我活到了现在。”　　“而从那时候开始，我遗忘了一些记忆，包括27日到韦尔奇住所做过什么，看到了什么，我没有撒谎，我真的不记得了。”　　为了洗清嫌疑，为了解决缠上自己的诡异事件，克莱恩几乎说出了全部的事情，除开穿越和“聚会”。　　另外，他在措辞上有所修饰，让每句话都能经得起考验，比如没说子弹未击中脑袋，只提未达到预想的效果，事后头部依旧完好。　　在旁人耳中，这两者几乎表达一样的意思，但实际上截然不同。　　灰眸警官邓恩安静听完，沉缓开口：　　“这很符合我推测的发展，也符合之前类似事件的隐藏逻辑，当然，我不知道你是怎么活下来的。”　　“你相信就好，我也不知道我怎么活下来的。”克莱恩稍微松了口气。　　“但是……”邓恩抛出了一个转折词，“我相信没用，现在的你有很高的嫌疑，你必须通过‘专家’的确认，确认你真的遗忘了遭遇，或者真的没直接导致韦尔奇先生和娜娅女士的死亡。”　　他咳嗽一声，表情变得严肃：　　“克莱恩先生，请你配合调查，和我们回一趟警局，这大概需要两到三天，如果你确实没有问题的话。”　　“专家到了？”克莱恩愣愣反问。　　不是说过两天吗？　　“她比我们预料得都早。”邓恩侧过身体，示意克莱恩出门。　　“我留张纸条。”克莱恩请求道。　　班森还在出差，梅丽莎上学去了，只能留言告诉他们自己涉及韦尔奇的一件事情，让他们不要担心。　　邓恩不甚在意地点头：　　“可以。”　　克莱恩回到书桌旁，一边找出纸张书写，一边开始思考接下来的事情。　　老实说，他非常不希望见那位专家，毕竟自身还藏着一个更大的秘密。　　在有七大教会的地方，在疑似“前辈”的罗塞尔大帝被刺杀的前提下，“穿越”这种事情多半是要进裁判所，上仲裁庭的！　　但是，没武器，没格斗技巧，没超凡之力的自己哪里是职业警官的对手，更何况，门外昏暗里还站着几位邓恩的下属。　　他们拔枪一个齐射，自己就算交代了！　　“呼，走一步算一步。”克莱恩留下纸条，拿上钥匙，跟着邓恩出了房间。　　昏暗的走廊里，四位黑衣白格的警察分列两边，非常戒备。　　啪，啪，啪，克莱恩跟在邓恩身边，踩着木制的楼梯，一阶一阶往下，时而能听到吱吱呀呀的声音。　　公寓门外停着一辆四轮单马的马车，它厢体侧面绘刻有“双剑交叉、簇拥王冠”的警察系统标志，周围和之前每个清晨一样热热闹闹，拥挤嘈杂。　　“上去吧。”邓恩示意克莱恩先。　　克莱恩刚要迈步，突然有个卖牡蛎的小贩抓住一位顾客，指责对方是小偷。　　双方扭打起来，惊到了马匹，周围顿时变得混乱。　　机会！　　克莱恩来不及多想，猛地弯腰前冲，抢入了人群里。　　或推搡，或闪避，他疯狂奔逃，向着街道另外一头。　　现在的情况下，为了不“见”专家，只能去城外码头，坐船顺塔索克河而下，逃到首都贝克兰德去，那里人口众多，便于隐藏。　　当然，也能扒蒸汽列车，往东去最近的恩马特港口，走海路到普利兹，然后才前往贝克兰德。　　不多时，克莱恩跑到了街口，拐入了铁十字街，那里停着几辆可雇佣的马车。　　“去城外码头。”克莱恩手一撑，跳上了其中一辆。　　他想的很清楚，要先故意误导追赶的警察，等到马车驶出一段距离，自己就直接跳下去！　　“好的。”车夫扯起了缰绳。　　哒哒哒，马车驶离了铁十字街。　　正当克莱恩准备跳车时，他忽然发现马车拐向了另外一条路，并非通往城外的道路！　　“你要去哪里？”克莱恩愣了一下，脱口问道。　　“去韦尔奇的住所……”马车夫语气不见起伏地回答。　　什么？克莱恩惊愕之中，马车夫转过身体，露出深邃冷漠的灰色眼眸，俨然便是邓恩·史密斯警官！　　“你！”克莱恩惊恐莫名，突感天旋地转，整个人猛然坐了起来。　　坐了起来？克莱恩疑惑地左看右看，发现窗外红月正盛，房间铺满“轻纱”。　　他伸手摸了下额头，湿润而冰凉，尽是冷汗，背后也是同样的感觉。　　“做了个噩梦……”克莱恩缓缓吐了口气，“还好，还好……”　　他觉得自己梦里还挺清醒的，还能冷静思考，颇为奇怪。　　稍微缓和后，克莱恩拿起怀表看了一眼，发现才半夜两点多，于是悄声下床，打算去公用盥洗室洗个脸，顺便解决下憋胀的小腹问题。　　扭开房门，他来到昏暗的过道上，就着微弱难辨的月光，脚步很轻地靠近公用盥洗室。　　突然，他看到走廊尽头的窗户前站了一道人影。　　那人影穿着比长袍短、比正装长的黑色类风衣服饰。　　那人影半融入于黑暗里，沐浴着清冷的绯红月华。　　那人影缓缓转过了身体，眼眸深邃、灰暗、冷漠。　　邓恩·史密斯！"
}