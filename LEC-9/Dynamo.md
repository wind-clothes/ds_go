## Dynamo 分布式的kv数据库

=============================

Dynamo：亚马逊的高可用kv数据库

我们为什么看这篇文章？
  数据库，最终一致，写任何副本
     像PNUTS一样，但总是成功
     像巴约，和解
     令人惊讶的设计。
  一个真正的系统：用于亚马逊的购物车
  比PNUTS，Spanner和c更可用
  比PNUTS，扳手和c少一致
  影响力设计; 灵感来自卡桑德拉
  2007年：在PNUTS之前，在Spanner之前

他们的痴迷
  SLA，例如延迟的百分之99.9百分之<300毫秒
  恒定故障
  “数据中心被龙卷风破坏”
  “永远可写”

大图
  [很多数据中心，Dynamo节点]
  每个项目通过密钥哈希在几个随机节点上复制

为什么复制在几个网站？为什么不在每个网站复制？
  有两个数据中心，站点故障占用节点的1/2
    所以需要注意* * * *复制在*两个*网站
  有10个数据中心，站点故障影响小部分节点
    所以只需要在几个站点的副本

在哪里放置数据 - 一致的散列
  [环，服务器的物理视图]
  节点ID =随机
  密钥ID =哈希（密钥）
  协调员：关键的继任者
    客户端将put / get发送到协调器
  接班人复制品 - “偏好列表”
  协调器将put（并获取...）转发到首选项列表中的节点

大多数远程访问的后果（因为没有保证的本地副本）
  大多数投放/获取可能涉及WAN流量 - 高延迟
    法定人数将削减尾端---见下文
  比PNUTS更容易出现网络故障
    再次因为没有本地的副本

为什么一致的哈希？
  临
    自然有些平​​衡
    分散 - 查找和加入/离开
  Con（第6.2节）
    不是很平衡（为什么不呢？），需要虚拟节点
    很难控制放置（平衡流行的按键，分布在网站上）
    加入/离开更改分区，需要数据移动

故障
  张力：临时或永久性失败？
    节点不可达 - 该怎么办？
    如果是临时的，将其他地方存储到节点可用之前
    如果永久性，需要制作所有内容的新副本
  发电机本身将所有故障视为暂时的

“永远可写”的后果
  总是可写=>没有主人！必须能够在本地写。
     想法1：草率的法定人数
  总是可写+失败=冲突的版本
     想法2：最终的一致性
        想法1避免不一致时，没有失败
     
想法＃1：草率的法定人数
  尝试获得单一主机的一致性优势，如果没有失败
    但是即使协调者失败，也允许进行进度，哪个PNUTS没有
  当没有故障时，通过单个节点发送读/写
    协调员
    导致读取在通常情况下看到写入
  但不要坚持！允许读取/写入任何副本，如果失败

临时故障处理：法定人数
  目标：不要阻止等待无法访问的节点
  目标：放置应该总是成功
  目标：得到应该有很高的看法最近的投机
  法定人数：R + W> N
    永远不要等待所有的N
    但R和W将重叠
    切断延迟分布，并容忍一些故障
  N是首选N *可达*节点的偏好列表
    每个节点对继承人进行ping，以保持粗略的上/下估计
    “马虎”的法定人数，因为节点可能不一致
  粗俗的法定意味着R / W重叠*不保证*

协调者处理put / get：
  并行发送put / get到前N个可达节点
  放：等待W回复
  得到：等待R回复
  如果失败不是太疯狂，将会看到最近的所有版本

什么时候这个法定计划*不*提供R / W交叉口？

如果put（）使数据远离环，怎么办？
  故障修复后，新数据超出N？
  该服务器记住了数据真正属于何处的“提示”
  一旦真正的家庭可达，转发
  也是 - 定期的“merkle树”同步的关键范围

想法2：最终的一致性
  接受任何副本的写入
  允许不同的副本
  允许读取看到陈旧或冲突的数据
  当故障消失时解决多个版本
    最新版本如果没有冲突的更新
    如果冲突，读者必须合并然后写
  像Bayou和Ficus  - 但在一个DB

最终一致性的不愉快后果
  可能不是唯一的“最新版本”
  读取可以产生多个冲突的版本
  应用程序必须合并并解决冲突
  没有原子操作（例如没有PNUTS测试和设置写）

多个版本如何出现？
  也许一个节点因网络问题而错过了最新的写入
  所以它有旧数据，应该被更新的put（）取代
  get（）咨询R，可能会看到更新版本以及旧版本

如何*冲突*版本出现？
  N = 3 R = 2 W = 2
  购物车，开始空“”
  偏好列表n1，n1，n3，n4
  客户端1想要添加项目X
    get（）from n1，n2，yield“”
    n1和n2失败
    put（“X”）到n3，n4
  n1，n2复活
  客户端3要添加Y
    get（）从n1，n2得到“”
    put（“Y”）到n1，n2
  客户端3想要显示购物车
    get（）从n1，n3得到两个值！
      “X”和“Y”
      不会取代另一个 -  put（）冲突

客户如何解决读取冲突？
  取决于应用程序
  购物篮：通过联合合并？
    将取消删除项目X
    比Bayou弱（删除权），但更简单
  一些应用程序可能会使用最新的挂钟时间
    如果我正在更新我的密码
    应用程序比合并更简单
  将合并结果写回Dynamo

编程API
  所有的对象都是不可变的
  -  get（k）可能返回多个版本，连同“上下文”
  -  put（k，v，context）
    创建一个新版本的k，附加上下文
  该上下文用于合并和跟踪依赖关系
  检测冲突。它由对象的VV组成。

版本矢量
  示例树版本：
    [A：1]
           [A：1，B：2]
    VVs表示v1取代v2
    发电机节点自动丢弃[a：1]，有利于[a：1，b：2]
  例：
    [A：1]
           [A：1，B：2]
    [a2]
    客户端必须合并

VV会不会变大？
  是的，但缓慢地，因为主要是从相同的N个节点服务
  如果VV具有> 10个元素，Dynamo将删除最近最近更新的条目

删除VV条目的影响？
  将不会实现一个版本包含另一个版本，不需要时会合并：
    put @ b：[b：4]
    put @ a：[a：3，b：4]
    忘记b：4：[a：3]
    现在，如果你同步w / [b：4]，需要合并
  忘记最老的是聪明的
    因为这是最可能存在于其他分支机构中的因素
    所以如果缺少，强制合并
    忘记*最新*将清除近期差异的证据

客户端合并冲突的版本总是可能吗？
  假设我们正在保留一个柜台，x
  x开始0
  递增两次
  但是失败阻止客户看到彼此的写作
  治愈后，客户端会看到两个版本，x = 1
  什么是正确的合并结果？
  客户可以弄清楚吗？

如果两个客户端同时写入w / o失败怎么办？
  例如，两个客户端同时将diff项目添加到同一个购物车
  每个都得到修改
  他们都看到相同的初始版本
  他们都将put（）发送到同一个协调器
  协调员会创建两个具有冲突的VV的版本吗？
    我们想要这个结果，否则就被抛弃了
    论文没有说，但协调者可以通过put（）上下文来检测问题

永久服务器故障/添加？
  管理员手动修改服务器列表
  系统洗牌数据 - 这需要很长时间！

问题：
  需要一段时间才能知道添加/删除的服务器
    到所有其他服务器。这是否会造成麻烦？
  删除的服务器可能会得到put（）的替换。
  在缺少一些put（）之后，已删除的服务器可能会收到get（）。
  添加的服务器可能会错过一些协调器不知道的put（）sb / c。
  添加的服务器可能会在完全初始化之前提供get（）。
  发电机可能会做正确的事情：
    Quorum可能导致get（）看到新数据以及陈旧的数据。
    复制同步（4.7）将修复错过的get（）s。

设计本质上是低延迟吗？
  否：客户可能被迫联系遥远的协调员
  否：一些R / W节点可能遥远，协调员必须等待

设计的哪些部分可能有助于限制第99.9页的延迟？
  这是一个关于差异的问题，而不是意思
  坏消息：等待多台服务器需要* max *的延迟，而不是平均
  好消息：发电机只能等待N或者N的W或R
    切断延迟分布的尾巴
    例如，如果节点有1％的机会忙于其他的东西
    或者如果几个节点断开，网络过载，＆c

没有真正的Eval部分，只有经验

亚马逊如何使用发电机？
  购物车（合并）
  会话信息（也许最近访问和c？）（最近的TS）
  产品列表（主要是r / o，高读取吞吐量的复制）

他们声称Dynamo的主要优点是灵活的N，R，W
  通过改变他们得到什么？
  NRW
  3-2-2：默认，合理快速R / W，合理耐用
  3-3-1：快W，慢R，不是很耐用，没用吗？
  3-1-3：快R，慢W，耐用
  3-3-3：减少R的机会丢失W？
  3-1-1：没用

他们不得不解决分区/放置/负载平衡（6.2）
  旧计划：
    节点ID的随机选择意味着新节点不得不拆分旧节点的范围
    哪些需要昂贵的磁盘数据库扫描
  新计划：
    预设Q组均匀范围
    每个节点是其中的几个的协调器
    新节点占用了几个整个范围
    将每个范围存储在一个文件中，可以将整个文件xfer

具有多个版本的能力有多大？（6.3）
  那么最终的一致性是多么有用
  这是他们的一个大问题
  6.3声称0.001％的读数有不同的版本
    我相信他们的意思是冲突的版本（不是良性的多个版本）
    那是很多还是一点点？
  所以也许有0.001％的写作受益于总是可写的？
    即将在主/备份方案中阻止？
  很难猜到：
    他们暗示问题是并发作者，为此
      更好的解决方案是单主机
    但也可能他们的测量不计算的情况
      单一主人的可用性会更差

性能/吞吐量（图4，6.1）
  图4表示平均10ms读取，20 ms写入
    20 ms必须包含磁盘写入
    10 ms可能包括等待N的R / W
  图4说，第99.9 ppt是约100或200 ms
    为什么？
    “请求加载，对象大小，位置模式”
    这是否意味着有时他们不得不等待沿海海岸的msg？ 

难题：为什么图4和表2中的平均延迟如此之低？
  暗示他们很少等待WAN延迟
  但第6节说“多个数据中心”
    你会期望*大多数*协调员和大多数节点是遥远的！
    也许所有的数据中心都在西雅图附近？
    可能是因为协调员可以是偏好列表中的任何节点？
      见最后一段5
    也许N中的W-1副本靠近？

包起来
  大想法：
    最终一致性
    总是可写，尽管失败
    允许冲突的写入，客户端合并
  一些应用程序的尴尬模型（陈旧的读取，合并）
    我们很难从纸上说出来
  也许是获得高可用性+在WAN上没有阻塞的好方法
    但是PNUTS的总体规划意味着雅虎认为不是问题
  关于最终的一致性是否对存储系统有好处是没有一致的