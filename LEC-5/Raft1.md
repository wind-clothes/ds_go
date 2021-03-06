## 一致性算法Raft（1）

这个讲座
  今天：筏选举和日志处理（实验2A，2B）
  下一个：筏持续性，客户行为，快照（实验室2C，实验3）

整体主题：使用状态机复制（SRM）的容错服务
  [客户端，副本服务器]
  示例：配置服务器，如MapReduce或GFS主控
  示例：key / value存储服务器，put（）/ get（）（lab3）
  目标：与单个非复制服务器相同的客户端可见行为
    但尽管有一些服务器故障
  战略：
    每个副本服务器以相同的顺序执行相同的命令
    所以他们执行时仍然是副本
    所以如果一个失败，其他人可以继续
    即故障时，客户端切换到另一台服务器
  GFS和VMware FT都具有这种风格

一个关键的问题：如何避免分裂脑？
  假设客户端可以联系副本A，但不能复制B
  客户端只能复制A吗？
  如果B真的崩溃了，客户端*必须*没有B，
    否则服务不能容忍故障！
  如果B已启动但网络阻止客户端联系，
    也许客户应该*不*没有它，
    因为它可能是活着的，为其他客户服务 - 冒着分裂的大脑

为什么不能允许分裂脑的例子：
  容错密钥/值数据库
  C1：put（“k1”，“v1”）
  C2：put（“k1”，“v2”）
  C1：get（“k1”） - > ???
  正确的答案是“v2”，因为这是非复制服务器将产生的
  但是如果两个服务器由于分区而独立地服务于C1和C2，
    C1的得到可以产生“v1”

问题：电脑无法区分“崩溃”vs“分区”

我们想要一个状态机复制方案：
  尽管有任何失败仍然可用
  处理分区w / o分裂脑
  如果太多故障：等待维修，然后恢复

应对分区的大见解：多数投票
  2f + 1服务器容忍f故障，例如3台服务器可以容忍1次故障
  必须获得多数（f + 1）票才能取得进展
    f服务器的失败保留了f + 1的大部分，可以继续
  为什么多数人帮助避免分裂大脑？
    最多一个分区可以拥有多数
  注意：大多数是超出所有2f + 1服务器，不仅仅是活的
  大多数人真正有用的是任何两个必须相交
    交叉路口的服务器只能以一种或另一种方式投票
    交叉点可以传达有关以前决定的信息

在1990年左右发明了两个分区容错复制方案，
  Paxos和View-Stamped复制
  在过去十年中，这项技术已经有了很多现实世界的使用
  筏纸是对现代技术的一个很好的介绍

***主题：筏概述

以Raft  -  Lab 3为例的状态机复制示例：
  [图：客户端，3个副本，k / v层，木筏层，日志]
  服务器的木筏层选举领导者
  客户端将RPC发送到k / v层
    放，获取，追加
  leader的Raft层将每个客户端命令发送到所有副本
    每个追随者都附加到其本地日志
  如果大多数人把它放在他们的日志中，进入是“承诺的” - 不会被遗忘
    多数 - >将被下一任领导人的投票要求所看到
  一旦领导者表示承诺，服务器执行进入
    k / v层应用到DB，或者获取获取结果
  然后领导回复客户端执行结果

为什么日志？
  该服务保持状态机状态，例如键/值DB
    为什么还不够？
  对命令进行编号很重要
    帮助复制品达成一个单一的执行顺序
    帮助领导者确保追随者具有相同的日志
  副本也使用日志来存储命令
    直到领导人承诺
    所以如果跟随者错过了一些，领导可以重新发送
    用于重新启动后的持久性和重播

这些服务器的日志会永远是相互准确的副本吗？
  否：一些副本可能会滞后
  否：我们会看到他们可以暂时有不同的条目
  好消息：
    他们最终会收敛
    提交机制确保服务器只执行稳定的条目
  
实验室2筏接口
  rf.Start（command）（index，term，isleader）
    实验3 k / v服务器的Put（）/ Get（）RPC处理程序调用Start（）
    在新的日志条目上启动Raft协议
    Start（）立即返回 -  RPC处理程序必须等待提交
    如果服务器在执行命令之前失去领导，可能无法成功
    isleader：如果这个服务器不是Raft的领导者，那么客户应该尝试另一个
    术语：currentTerm，以帮助呼叫者检测到领导者是否被降级
    index：日志条目，以查看该命令是否提交
  具有索引和命令的ApplyMsg
    Raft在每个“apply channel”上发送消息 
    提交日志条目。该服务然后知道执行
    命令，领导使用ApplyMsg
    知道何时/什么回复等待客户端RPC。

Raft设计有两个主要部分：
  选举新领导人
  确保相同的日志尽管失败

***主题：领导选举（实验室2A）

为什么是领导者？
  确保所有副本以相同的顺序执行相同的命令

筏子编号的领导序列
  新任领导 - >新任期
  一个任期最多只有一个领导; 可能没有领导者
  编号有助于服务器跟随最新的领导者，而不是被替代的领导者

Raft何时开始领导选举？
  其他服务器没有听到当前领导的“选举超时”
  他们增加地方当前条款，成为候选人，开始选举
  注意：这可能导致不必要的选举; 这很慢但安全
  注意：老领导人可能还活着，认为是领导者

一个领域最多只能保证一个领导？
  （图2 RequestVote RPC和服务器规则）
  领导者必须从大多数服务器获得“是”票
  每个服务器每个术语只能投一票
    对第一个服务器的投票（在图2规则内）
  一个服务器最多可以获得一个给定期限的多数票
    - >最多一个领导者即使网络分区
    - >选举可以成功，即使有些服务器失败

服务器如何知道选举成功？
  获胜者获得多数票
  其他人看到AppendEntries的心跳从胜利者
  新领导人的心跳压制任何新的选举

选举可能不成功有两个原因：
  *不到大多数服务器可达
  *同时候选人分散投票，没有获得多数票

如果选举不成功，会发生什么？
  另一次超时（无心跳），另一次选举
  较高的期限优先，老年人的候选人退出

Raft如何避免分散投票？
  每个服务器选择一个随机的选举超时
  [服务器超时到期的时间图]
  随机性破坏服务器之间的对称性
  一个会选择最低的随机延迟
  希望足够的时间在下一次超时之前选择
  其他人将看到新的领导者的AppendEntries心跳和 
    不成为候选人

如何选择选举超时？
  *至少有几个心跳间隔（网络丢失心跳）
  *随机部分足够长，让一个候选人在下次启动之前成功
  *足够短，让测试人员不舒服之前进行几次重试
    测试者需要在5秒以内完成选择

如果老领导人不知道新的领导人当选，怎么办？
  也许老领导人没有看到选举信息
  也许老领导人是在少数网络分区
  新的领导意味着大多数服务器都增加了currentTerm
    所以老领导（w / old term）不能获得AppendEntries的多数
    所以老领导不会提交或执行任何新的日志条目
    因此没有分裂的大脑
    但少数人可能会接受旧服务器的AppendEntries
      所以原木可能在旧术语结束时分歧