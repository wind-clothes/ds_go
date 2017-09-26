## QA
>筏似乎比替代品更容易理解和简单（Paxos，
> Viewstamped复制），文章中提到的可能状态较少
>考虑。筏子为这种简单性而付出的代价是什么？

为了清晰起见，木筏放弃了一些表演; 例如：

*每个操作必须写入磁盘才能持久化; 性能
  可能需要在每个磁盘写入中批量许多操作。

*只能有效地从一个单一的AppendEntries飞行中
  每个追随者的领导者：追随者拒绝无序
  AppendEntries和发件人的nextIndex []机制需要
  一次一个。用于流水线许多AppendEntries的规定
  会更好。

*快照设计只适用于相对较小的设计
  状态，因为它将整个状态写入磁盘。如果状态是
  大（例如，如果它是一个大数据库），你需要一种方法来写
  最近改变的部分国家。

*类似地，通过发送他们来恢复最新的副本
  完整的快照会慢，不必要地，所以如果已经复制了
  有一个旧的快照。

*服务器可能无法充分利用多核，因为
  必须一次执行一次操作（按照日志顺序）。

所有这些都可以通过修改Raft来修复，但是结果会是这样
可能具有较少的价值作为教程介绍。

>在现实世界的软件中使用木筏，或者公司通常都会滚动
>他们自己的Paxos风味（或使用不同的共识协议）？

筏子是相当新的，所以没有太多的时间让人们设计
基于它的新系统。大多数国家机器复制系统我
听说是基于较旧的Multi-Paxos和Viewstamped
复制协议。

> RAFT设计师如何将系统设计成与Paxos一样高效
>同时保持学习能力？是难以学习的部分
> Paxos只是语义的结果？

有一个名为Paxos的协议，允许一组服务器同意
在一个单一的价值。而Paxos需要一些想法来理解它
比筏子简单得多。但是Paxos解决的问题要小得多
筏。以下是关于Paxos的易读文章：

  http://css.csail.mit.edu/6.824/2014/papers/paxos-simple.pdf

要建立一个真实的复制服务，复制品需要达成一致
一个不确定的值序列（客户端命令），他们需要
服务器崩溃并重新启动或丢失时有效恢复的方法
消息。人们以Paxos为起点建立了这样的系统
点; 查看Google的胖子和Paxos Made Live论文，以及
动物园管理员/ ZAB。还有一个名为Viewstamped Replication的协议;
这是一个很好的设计，类似于筏，但是关于它的文章很难
要了解

这些真实世界的协议是复杂的，（在Raft之前）没有
一篇介绍性文章，介绍他们的工作原理。筏纸，在
对比，比较容易阅读。这是一个很大的贡献。

筏式协议是否固有地更容易理解
明确。这个问题由于缺乏对其他的描述而被掩盖
真实世界的协议。另外，Raft牺牲了性能
透明度在很多方面; 这是一个教程，但不总是
在现实世界的协议中是可取的。

>作者创作筏之前Paxos存在多久？怎么样
普遍的现象是筏子在生产中的应用？

Paxos是在20世纪80年代末发明的。筏子大致开发
2012年，Paxos之后20年。

Raft非常类似于称为Viewstamped复制的协议，
最初发表于1988年。有复制的容错文件
在20世纪90年代早期建立在Viewstamped复制之上的服务器，
虽然不在生产中使用。

我知道一系列来自Paxos的真实世界系统：
胖子，扳手，Megastore和Zookeeper / ZAB。从早期或
2000年代中期的大型网站和云提供商需要容错
服务和Paxos或多或少被重新发现在那个时候
进入广泛的生产使用。

有几个真实世界的筏夫用户：Docker
（https://docs.docker.com/engine/swnarm/raft/）和etcd（https://etcd.io）。其他
据说使用Raft的系统包括CockroachDB，RethinkDB和TiKV，但是
无疑有其他的。也许你可以从别处找到别人
http://raft.github.io/

>在现实世界的应用中，Raft的性能与Paxos的相比如何？

最快的Paxos派生协议可能比Raft快得多; 有一个
看看ZAB / ZooKeeper和Paxos Made Live。但我怀疑是牺牲了
使它变得更复杂，可以发展Raft以获得非常好的表现。
Etcdv3 etcd3声称比zookeeper和许多人都取得了更好的表现
基于paxos的实现（https://www.youtube.com/watch?v=hQigKX0MxPw）。

>设计算法越来越受欢迎
>可理解性，还是筏子的特殊情况？

很少见到发表研究论文的主要贡献
清晰度

作者以这些精确的方式提出了这样一个系统
>规则 他们是从一个简单的开始，然后逐渐细化，
>添加规则来解决违反物业的情况？要么
他们是从一些总体模式开始的，这些系统是如何应用的
>工作（然后反复细化）？

我不知道。

筏子的设计类似于现有的做法，回到
Viewstamped复制协议于1988年发表。随机领导者
至少可以追溯到20世纪70年代初。所以很久
设计和工程的传统，他们可以借鉴，如果他们想要的
至。

有关细节，也许他们有一些迭代过程
他们通过设计，实施或模拟循环，
调试问题，然后改变设计。或者开车
通过试图证明正确性而不是通过测试来进行细化。

>为什么要学习/实施筏子而不是Paxos？

我们在6.824中使用了筏子，因为有一张纸很清楚
介绍如何使用Raft构建完整的复制服务器系统。一世
知道没有令人满意的文章描述如何构建一个完整的
基于Paxos的复制服务器系统。

>文章说，Paxos实现真正的问题必须是
>从正确证明的Paxos版本显着改变，
>但是这对于筏来说不太真实，因为筏子更容易
>理解并因此更容易应用于真正的分布式共识
>应用程序。大概在筏上还需要做一些改变
>应用于实际系统。那些是什么变化？是吗
规则的放松，也许是因为简单是值得的
牺牲罕见边境案件的正确性？或者是更改
>建筑制作？我也不知道有什么变化
这篇论文正在谈论什么时候发生重大变化
>在实际应用中给Paxos，我好奇如何
>平行于筏。

在Paxos的情况下，原来的Paxos不是复制状态
机器系统; 它只执行一个协议。所以大家谁
用它作为复制的基础必须发明一种使用它的方法
同意一系列客户端命令。事实证明，如果你想要的
效率，你还需要一个领导者，所以最重要的使用Paxos
还必须添加算法来选择和使用领导者。同样的
有效的方案将滞后的副本更新到检查点
并恢复服务状态，无需担任只读操作
添加日志，更改副本中的服务器集
组，并检测重新发送的客户端请求并仅执行一次。

筏已经具有所有这些功能。也许有东西
现实世界的使用仍然需要改变关于筏，但可能不会几乎
与基于Paxos的系统一样多。

>两者之间有什么大的区别/优点/缺点？

筏子的设计非常接近许多基于Paxos的设计。有
筏子牺牲表现获得的几种方式
简单。筏不连续/重复连续操作。它
没有特别有效的方式来实现滞后的追随者
快速日期，或将筏状态保持到磁盘。筏子的快照是
对于微小的状态，但是对于具有大状态的服务而言非常缓慢。它是
拙劣的纸张不包括严肃的性能评估
这将有助于我们了解所有这些对于性能的重要性。


> 1）有没有什么好的资源来快速了解Paxos是什么？
>看起来这对于Raft系统的上下文是有用的。

这是一个介绍：

  http://css.csail.mit.edu/6.824/2014/papers/paxos-simple.pdf

这里是一个Paxos的演讲，可以与筏子相媲美：

  https://www.youtube.com/watch?v=JEpsBg0AO6o

> 2）有没有像筏子这样可以生存和继续的系统
>操作（提交新日志条目）只有少数的
>群集是活动的？如果是，如何，如果不是，它是可以证明的
>不可能有这样一个系统？

不是与筏的属性。但是你可以用不同的方式来做
假设或不同的客户端可见语义。基本的问题是
分裂脑 - 一个以上服务器的可能性
领导人。我知道有两种方法。

如果不知何故客户端和服务器可以准确了解哪些服务器正在生活
哪些是死的（而不是生活，而是由网络划分
失败），那么可以建立一个只能一个功能的系统
还活着，挑选（说）已知活着的最低编号的服务器。
但是，一台电脑很难决定是否还有一台电脑
死亡，而不是网络丢失他们之间的消息。一
做到这一点的方法是要有一个人的决定 - 人类可以检查每一个
服务器并决定哪些是活着的和死的。这可能是GFS
决定主机何时死机，并切换到备份。

另一种方法是允许分脑手术，并有办法
使服务器在分区之后调和所导致的分歧状态
被治愈 这可以用于某些类型的服务，但有
复杂的客户端可见语义（通常称为“最终”）
一致性“）。看看Bayou和Dynamo的论文
在课后分配。

>在木筏中隐含地假设的一个不变量是，
>服务器总数不变。这需要额外的要求
>这些服务器的至少一半必须存在于系统中
>功能。有没有办法摆脱这个要求呢？

这里的问题是分脑 - 可能性不止一个
服务器认为它是领导者。我知道有两种方法
的。

如果不知何故客户端和服务器可以准确了解哪些服务器正在生活
哪些是死的（而不是生活，而是由网络划分
失败），那么可以建立一个只能一个功能的系统
还活着，挑选（说）已知活着的最低编号的服务器。
但是，一台电脑很难决定是否还有一台电脑
死亡，而不是网络丢失他们之间的消息。一
做到这一点的方法是要有一个人的决定 - 人类可以检查每一个
服务器并决定哪些是活着的和死的。这可能是GFS
决定主机何时死机，并切换到备份。

另一种方法是允许分脑手术，并有办法
使服务器在分区之后调和所导致的分歧状态
被治愈 这可以用于某些类型的服务，但有
复杂的客户端可见语义（通常称为“最终”）
一致性“）。看看Bayou和Dynamo的论文
在课后分配。

>在Raft，正在复制的服务是
>在选举过程中不能向客户提供。
>在实践中，这是由多少问题引起的？

客户端可见的暂停似乎可能在十分之一的数量级
第二。作者期望失败（因而选举）是罕见的，
因为只有当机器或网络出现故障时才会发生。许多服务器
网络一直持续数月甚至数年，所以
这对于许多应用来说似乎不是一个巨大的问题。

>是否有其他不具备任何停机时间的系统？

有基于Paxos的复制版本没有领导者
或选举，因此在选举期间不要停顿。
相反，任何服务器都可以随时有效地担当领导者。成本
没有一个领导者是每个都需要更多的消息
协议。

>筏与VMware FT如何相关？他们的主要区别是什么？
>含义？

VM-FT可以复制任何虚拟机的guest虚拟机，从而任何服务器式
软件。筏只能复制专门为复制设计的软件
为筏。另一方面，VM-FT打破了测试的关键决定
共享磁盘的功能 - 如果共享磁盘发生故障，VM-FT将无法正常工作。筏
将是构建容错共享磁盘的好方法。

>为什么要使用VM-FT而不是Raft？

如果你有一些现有的软件，你想复制，但没有
想要努力使它与筏子一起工作。


>保护RAFT系统的最佳方法是什么？
机器接管对手，成为领导者，而不是放弃
>对存储的信息造成损害？

你是对的，筏子没有任何防御攻击
这个。它假设所有参与者都遵循协议，
只有正确的服务器集才能参与。

真正的部署必须做得比这更好。最多
直接的选择是将服务器放在防火墙后面
过滤来自Internet上随机人的数据包，并确保
防火墙内的所有计算机和人员都值得信赖。

在某些情况下，Raft必须在同一个网络上运行
潜在的攻击者 在这种情况下，一个好的计划是认证
具有一些加密方案的Raft数据包。例如，给每个
合法的Raft服务器是一个公钥/私钥对，拥有它所有的签名
它发送的数据包，给每个服务器列出公钥的列表
合法的Raft服务器，并让服务器忽略不是的数据包
由该列表上的键签署。

>该文件提到筏子在所有非拜占庭人的条件下都有作用。
>什么是拜占庭式的条件，为什么他们会使筏失败？

这篇论文的“非拜占庭条件”是指这样的情况
服务器软件和硬件可以正常工作或者停止
网络可能会丢失，延迟或重新排序消息。
该文件说，在这些条件下，筏子是安全的。例如，
大多数电源故障是非拜占庭式的，因为它们导致计算机
只需停止执行指令; 如果发生电源故障，Raft可能
停止运行，但不会向客户发送不正确的结果。

除上述以外的任何故障都称为拜占庭式
失败。例如，如果服务器上的Raft软件有错误
导致它不能正确遵循筏式协议。或者如果CPU
执行指令不正确（例如，添加100加100并获得201）。
或者如果攻击者进入其中一台服务器并修改其筏架
软件或存储状态，或者如果攻击者从/到Raft伪造RPC
服务器。如果发生此类故障，Raft可能会发送不正确的结果
给客户。

6.824的大部分是关于忍耐非拜占庭故障。正确
尽管拜占庭故障的行动更加困难; 我们会接触
这个话题在期末。

>在图1中，客户端和服务器之间的接口是什么
喜欢？

通常是RPC与服务器的接口。对于一个键/值存储服务器
例如您将在实验3中构建，它是Put（键值，值）和Get（值）RPC。
RPC由服务器中的一个键/值模块处理，该模块调用
Raft.Start（）要求Raft将客户端RPC放在日志中，并读取
applyCh来学习新提交的日志条目。

>如果客户端发送相同的请求不止一次
>有失败？

是。

>客户端会以异步方式发送多个请求吗？

这取决于软件是否支持。对于实验室系统，
没有。对于更多的工业强度体系，也许是的。

有一件我不太明白的是，当领导收到一个
>请求从客户端，然后这个服务器崩溃。另一个服务器是
>当选为新领导人。那么，为什么新的领导人可以接受
>同样的请求？

如果客户端没有得到其原始请求的响应，则其RPC
将超时，并将请求发送到不同的服务器。


>我仍然对于筏子的安全守卫感到困惑。如果
>某些服务器在收到请求之后立即崩溃，然后才能提交
>它，并且另一个术语通过并且服务器重新启动后，此服务器
>与其他服务器相比，日志不是最新的，所以它
>不能是领导者。数据不能从追随者流向领导者，所以
>新领导者永远不会知道这个请求，而这个请求是
>永久丢失 是这样吗？

如果未提交日志条目，则Raft可能不会在领导变更中保留日志条目。

没关系，因为客户端不能收到回复
请求Raft没有提交请求。客户会知道（by
看到超时或领导更改）其请求未提供，和
将重新发送。

>筏子如何处理裂脑问题？那就是两组
>服务器彼此断开连接。每组会选一个
>领导人，总共有两个领导人？

最多可以有一个主动领导。

只有当一个新的领导者可以联系大多数的服务器时，才能当选
（包括自身）与RequestVote RPC。所以如果有一个分区，和
其中一个分区包含大多数服务器，一个
分区可以选一个新的领导。其他分区必须只有一个
少数民族，所以他们不能选举领导人。如果没有多数
分区，将不会有领导（直到有人修理网络
划分）。

假设在网络被分区时选出了一个新的领导者。什么
关于前任领导人 - 它将如何知道停止提交新的
项？前任领导人将无法获得多数
对其AppendEntries RPC的真实回应（如果它在少数）
分区），或者如果可以与大多数人交谈，那么多数人必须重叠
与新的领导者的多数，和重叠的服务器将告诉
老领导人有更高的期限。那会导致老的
领导转向追随者。

>我有一个类似于阅读的问题。假设我们有一个
>情况类似于图7中的文章，但我们也有一个网络
>分区，使得七节点网络的两半将具有三个
>服务器。因为两个分区都不能获得多数，所以没有人可以
>当选领导 假设第七个服务器永远不会恢复在线（或至少
>需要很长时间），是否有一种方式来改变筏子的数量
>网络中的服务器，或者至少通知sysadmin那个不
>正在取得进展？

筏子不能做任何没有多数的东西。原因是如果少数
服务器被允许做任何事情，然后两个不同分区的少数人
可以做那件事 我们不希望两个少数民族有两个不同
分区都决定减少服务器的数量，从那以后他们会
能够在不知道彼此的情况下进行操作，造成脑裂。

系统管理员可以很容易地注意到，基于Raft的服务已经不再存在了
响应。也许系统管理员可以选择其中一个少数分区，
希望是最新的一个，并允许它继续告诉它
减少服务器总数。筏不支持这个，但你可以
想象一下修改它来做到这一点。但是，系统管理员必须做一个
复杂的决定，这将是容易出错的。这将是更好的
sysadmin修复网络分区，允许Raft按照设计继续。


>在图7中，似乎是因为f永远不会当选领导者，日志
>只有它有，即从术语2的所有日志将永远失去因为
>那个工作人员必须覆盖它们。这是正确的，如果是这样的话
>不是筏子的关注？

没错 有问题的日志条目不能被提交（他们
不在大多数服务器上），因此客户端无法收到回复
要求。客户最终会注意到领导者或时间的变化
将其RPC重新发送给新的领导者。

>什么阻止领导从不覆盖或删除其任何日志？

领导者遵循图2中的规则。图2中的任何内容都不会导致a
领导从其自己的日志中删除任何内容。

>候选人如何有资格从中获得选票
>其他服务器。追随者服务器如何知道何时和谁投票？

候选人向所有服务器发送RequestVote RPC。每个服务器投票
根据算法，它接收的每个RequestVote的是或否
在图2中的RequestVote RPC下。

>我明白，这是先到先得的基础，但是
>候选人宣布候选人到其他服务器接受投票？

候选人发出RequestVote RPC; 这隐含地宣布了它
候选人资格。

> electTimeout的安全假设是多么可靠
>一个数量级大于broadcastTime，一个顺序
>幅度小于MTBF？

选举超时的一个不好的选择不会影响安全，只会影响
活跃度。

有人（程序员？）必须选择选举超时。该
程序员可以轻松选择一个太大或太小的值。如果
太小，那么追随者可能会在领导人之前重复超时
有机会发出任何AppendEntries。在这种情况下，木筏可能会花费所有
选择新领导人的时间，没有时间处理客户的要求。
如果选举超时太大，那么就不必要了
领导人失败之前，一个新的领导人当选之前的大停顿。

>你能解释一下如何从网络中恢复的方法
>分区，例如当有一个3路分区，没有一个
>分区包含大多数服务器？

在这样一个分区中，筏子将无法选择任何领导者，因此
不会提供任何客户端请求。Raft服务器将再次尝试
再次选举领袖，永不成功。

如果有人修复网络，以使分区脱落，和a
大多数服务器可以互相交谈，然后选举会
成功了，而且Raft将开始重新提供客户端请求。

为什么重试超时成为候选人？是吗
>避免两个竞争的候选人开始一个任期的情况
同时又继续这样做，因此阻止选举？

这是对的

>在选举限制部分，如何界定“多数
>集群“，候选人需要联系？

以“多数”为主，这个意思是整体的一半以上
服务器。

>如果超过一半，那么对于纸张问题，服务器（a）可能会
>也当选

我同意。

但是，它不包含所有提交的条目（术语7）。

条款7中的条目不能被提交，因为它们不是
存在于大多数服务器上。领导者只提交日志条目
如果它从大多数收到对其AppendEntries的肯定回复
服务器。

>如果一个网络分区永远不能解决可以双方的
>分区意识到，多数节点的数量不同
并在那时选举新的领导人？

我想你建议有两个领导，每个都有自己的领导
较小的多数。这是不允许的，因为它会违反目标
复制的服务看起来像是普通的客户端
非复制服务。也就是说，它会产生一个“分裂的大脑”。

假设服务是一个键/值数据库。如果客户A和B执行
以下序列的放置/获取，一次一个，我们期望的
final get（“key1”）来产生“value2”。

客户端A：put（“key1”，“value1”）
客户B：put（“key1”，“value2”）
客户A：get（“key1”） - > ???

但是，如果有两位认为他们有多数的领导人，
那么客户A就会和一个领导和客户B谈谈对方
客户端A不会看到客户端B的放置值，最后一次获取将会产生
“VALUE1”。这是不合法的，因为它不会发生在一个
非复制单服务器设置。

>安全证明5.4.3节的步骤6对我来说有点不清楚。
>如果领导人与选民为什么有同样的最后一个任期
要使其日志至少与选民的日志一样长吗？

第2步说，选民有提交的日志条目有问题
选民是投票给leader_u的多数人的一部分。选举
限制规则（第5.4.1节）说，如果选民和候选人的
最后的日志条件是相等的，选民只会投票“是”如果
候选人的日志至少与选民的日志一样长。由于每个学期都有
只有一个领导者，那个领导人只发出一个日志序列
这个词的条目，领导者的日志更长，有的事实
与选民日志相同的最后一个术语意味着领导者的日志
包含投票人日志中该条目​​的所有条目。包括
有承诺的条目。

（f）在图7中是不同的情况，因为（f）有不同的
最后一个在日志中比新领导者。

>如果添加了新的服务器，其余的当前状态如何
>机器安全地复制到这个新服务器上？

这种机制变成了滞后的一个机制
跟随最新的领导者。当新服务器加入时，它会
拒绝从领导者收到的第一个AppendEntries，因为
新服务器的日志是空的（图2中的AppendEntries RPC下的步骤2）。
那么领导者将减少其新的服务器的nextIndex [i]直到它
是0.然后，领导将使用全新的服务器发送整个日志
AppendEntries。

>第6节中的C_old和C_new变量究竟是什么？
> 11）代表？他们是每个配置的领导者吗？

它们是旧/新配置中的一组服务器。

>有没有办法解决多个领导服务器下降
>顺序地，并且整个系统死锁？只是一个
这是不可能发生的概念，所以这样的情况是
>不可能

是的，你记住，大多数的服务器都死了，所以
选举不成功？事实上，筏子在这方面不会有任何进展
情况。如果发生这种情况，真正解决问题的唯一方法就是
获得更可靠的服务器。

>在图2中，作者说，这些规则“独立触发”。
>这是异步的一样吗？

在一个真正的实现中，你可能会在你的点上检查规则
更改相关状态的代码（例如RPC处理程序）。你也会
需要一个或多个执行诸如超时之类的独立线程
这不符合筏子状态的具体变化。

>如果是这样，不执行序列化
>锁在筏子的状态？

是。


>可以通过设置a进一步简化木筏的实施
>具体调度循环的事件？
>
>分类：
>
> func Schedule（）{
> //轮询事件而不是中断状态机
> for {
> AppendEntries？ - > do append（）//在这些函数内部发生状态变化
> AppendEntriesReply？ - > do append_reply（）
> ElectionTimeout？ - > do start_election（）
> VoteRequest？ - > do vote_response（）
> VoteRequestReply？ - > do process_vote_reply（）
> ... //你得到它
>}
>}
>
>我想像即使所有的事件都是异步触发的，那么这是可能的
>为一个事件时间表，如上所述发生。问题是是否全部
>可能的事件计划可能会发生异步触发器。

这似乎是合理的。大概RPC处理程序将发送消息
在AppendEntries渠道上？在我的实现我基本上
直接从RPC处理程序调用append_reply（），而没有
任何相当于Schedule（）的东西。我有一个单独的goroutine
选举超时，以及单独的goroutine来决定日志条目
已经承诺并将其发送到applyCh。

>筏如何处理加入群集的新服务器？

看看论文的第6节。

一个候选人可以在收到报名后立即宣布领导
多数投票弃权剩余票数

是的 - 多数是足够的。

>或者会产生问题，因为它可能会学习另一个服务器
>在剩下的一票中有较高的期限。

实际上可以学习一个更高一级的领导者，但是议定书
正确处理这个。任何这样的新领导本身都有
从大多数人收集投票，​​多数人会知道
新的更高的期限。低级领导只能修改
如果大多数接受其AppendEntries RPC的日志，但是它们不会
因为至少有一个多数的服务器会知道更高的服务器
新领导人的任期。


>筏子上的领导人不要在崩溃之外成为领导者吗？

是。如果领导者的网络连接中断，或丢失太多的数据包，
其他服务器将看不到其AppendEntries RPC，并将启动
选举。类似地，如果领导者或网络太慢。

>如果不是，这可能会导致负载平衡的问题，如果我们
>在同一个云端和所有的上面有多个不相交的服务
>领导者碰巧在一台服务器上聚合？

您需要在Raft之外的一个机制来管理这个。但是
机制可以轻易地要求筏子领导下台。

>选举期间如何处理日志条目？没有明确的领导，
>如何没有日志条目丢失？

在选举期间，服务器对他们的日志不做任何事情，所以
没有什么会丢失

领导人当选后，领导人可能要求追随者回滚
他们的日志使它们与领导者的日志相同。但是
这样做的算法保证永远不会丢弃提交的日志条目，所以
没有什么价值会丢失。

>追踪者的日志条目何时发送到他们的状态机？一个
>跟随者可能具有一组给定的日志条目，但它们可能会被更改
当一个新的领导人出现，有不一致的时候。

只有领导人说出一个条目才能实现。木筏永远不会
改变了关于承诺日志条目的想法。

>该文件指出，领导人永远不会回应客户，直到
>将条目应用于状态机后。这多久了
通常采取？延迟延迟是否会产生任何问题？

这取决于服务。如果它是存储其数据的数据库
磁盘，并且客户端要求写入数据，写入可能需要几个
几十毫秒。但是，这与或多或少相同的延迟
服务将施加无筏，所以大概没关系。

>如果选举比较频繁，那么服务器会继续存在
>很少经常出现微小的时间差距。那表现
打击真的被客户忽视了吗？

客户会注意到 - 他们会看到延迟，并且必须重新发送
如果最初将它们发送给失败的领导者，那么请求。但选举
只能在服务器发生故障或断开连接时发生。没有木筏
（或类似的东西），客户端根本不会得到任何服务
有失败。筏子通过继续使事情更好
尽管有一些失败，尽管确实在选举延误之后。

>在实验室阅读和工作时我不清楚的一件事是
>如果领导发送一个appendEntries RPC，如果它等待每个RPC
完成（成功或不成功），只需等待多数，
>或者没有。

领导人不应该真的等待。相反，它应该发送
AppendEntries RPC并继续。它也应该看每一个答复
它的RPC用于特定的日志条目，并计数这些回复，并标记
日志条目只有当它有大多数的回复时才提交
服务器（包括自身）。

Go中的一种方法是让领导发送每个AppendEntries
RPC在一个单独的goroutine中，使领导发送RPC
同时。这样的东西

  对于每个服务器{
    go func（）{
      发送AppendEntries RPC并等待回复
      如果reply.success == true {
        递增计数
        如果count == nservers / 2 + 1 {
          此条目已提交
        }
      }
    }（）
  }


>服务器可能有更多条目的情况
>领导者，像图7中的服务器（d）？

（d）您可能会指出，第7期可能是领导者。
还有一些其他服务器（？）是第7期的领导者，
只有（d）（？）听到新的日志条目，以及领导者之后
改变新的领导者（？）来回滚未提交的条目
第7期，但没有回头问（d）回滚。

>另外一个问题是，在一个领导的日志中，是否真的
>将被复制到其他服务器，但不是
>承诺

是。

>这很好，因为客户端不会收到任何输出
>这些未提交的命令。

是。

>这些只有当一个领导人当选时才会发生，而且它是引发的
>从旧的术语，没有被复制到所有其他
>服务器

领导也可以在其日志中提供未提交的条目，因为它
还没有听到大多数的AppendEntries RPC回复。

>我对这个问题感到困惑：它
>似乎承诺的条目是111445566，因为大多数
>服务器有这些条目。服务器（a）有这些条目，但只有它们
>和（b）将投票，因为（c）和（d）有更新的条目和
根据我的理解，不会投票。但不应该
>它技术上被允许赢，因为它有所有的承诺
>条目？

我认为服务器（a）可以获得多数，因为（a），（b），（e）和（f）
将投票。

>如果超过一半的服务器死机会发生什么？（我也问了这个
>广场）

服务不能取得任何进展; 它会继续选择一个
领导一遍又一遍。如果/足够的服务器恢复生活
持续的筏状态完好无损，他们将能够选举领导和
继续。

>筏确保确保一致性最终，但不是很多
>日志之间的差异可能导致的实际示例
>有用的数据被删除？

将被删除的唯一日志条目对应于客户端请求
服务从来没有回复给客户端（因为请求从不
提交）。这样的客户端将重新发送请求到另一个
服务器，直到他们找到当前的筏子领导者，（如果没有
更多的失败）会将请求提交到日志中，执行它们
回复客户端 所以最终没有什么会丢失的。

>当向另一个服务器发送requestVote消息时，如果发送
>消息具有比接收服务器的术语更高的术语
>接收服务器增加其条件以匹配发送的taht
>讯息？

是。您可以在图2的“服务器规则”顶部附近看到这一点。


>为什么他们选择首先索引日志？如果有问题的话
>日志被零索引了吗？

这只是一个方便。实际上我怀疑作者的实施
在日志数组中有一个条目0，其中Term = 0。这样就可以了
第一个AppendEntries RPC包含0作为PrevLogIndex，并且是有效的
索引到日志中。我认为修改协议很容易
具有零索引日志，并更改使用0的各个位置
意思是“在第一个日志条目之前”改为使用-1。

>另一种选择是不允许单个服务器成为太多的领导者
>条款？因此，这将阻止某些服务器成为领导者太久
>万一有恶意。

我担心，如果服务器是恶意的，它可以做到同样多的伤害
喜欢在短时间内。

>在复制状态机模型中，状态机从中读取
>登录每个复制的服务器。但是，如果发生某些问题
>发送命令到日志的共识模块，不可能
>导致不同的状态？为什么我们不能拥有状态机
>不同的服务器而不是从一个日志读取，从而保证
>所有服务器的状态都是一致的？一个论点是错误
>容差，在这种情况下，我们可以保留2或3个替换备份日志
原来因原因而丢失原件 看完之后
>更多的论文，似乎这点是通过使用
>领导者追随者模型，我们可以保证一致性
>服务器。但是，这似乎更复杂一些
>容易出错的方式。

筏基本上做的是你的建议，但事实证明是公平的
复杂。筏子保留日志的一个主要副本（在领导者中），还有
日志的一些备份副本（在追踪者中）。它使用领导者
如果领导者还活着，数据的副本可以回答客户请求
如果发生错误，请切换到备份。我们不能期待的副本
的日志总是相同的，因为一些服务器可能会
被破坏或网络连接断开。大部分的复杂性
筏与维护正确性有关，尽管可能会有日志
不完全一样。

>在给定期限内选择领导者时，为什么不分立投票
>随机，或给服务器整数，并选择服务器
>最高的id（类似排名的想法，但只有在分裂的情况下
>投票）？

你可以根据这些想法进行选举。作者
尝试了至少一种这样的方式（排名），但请注意它是
更复杂。如果选举非常频繁，而且经常
分裂，可能值得设计一个更复杂但更快的方案。

>不要选择领导者一个任期是不是浪费？

是; 这意味着有时选举需要几百毫秒
长于严格要求（额外的选举超时）。但
选举只有在失败之后才会发生，如果这样做不会很频繁
该系统是完成任何工作。所以选举可能的事实
有时需要更长的时间可能不是很重要
整体系统性能的因素。

>甚至在文中承认选举有可能
>无限期地发生（尽管不太可能由于随机选举
>超时）。

这只能在服务器反复挑选随机选举时发生
对于多个服务器而言，超时是相同的。这可以
发生，但它就像滚动骰子多次，得到相同
数字可能，但概率迅速下降
你滚多次

>我可能会理解一些错误，但是文章说RAFT
>如果没有全部提交，则阻止服务器被选举
>条目。然而，从服务器“a”可能的练习看来似乎是
>当然没有所有提交的条目。我失踪了什么

我认为服务器“a”具有可能已经提交的所有条目。

你在想“c”和“d”的“6”条目
超出“一个”日志的结尾？这些都不能被承诺。为了
对于那些把这些“6”条目交给他们的领导人来说，
那个领导人不得不从大多数人那里得到积极的答复
（包括自己），这是7台服务器中的4台。但那些“6”条目
只有三台服务器，显然领导没有得到
多数，所以没有承诺。

>如果领导人的选举是必要的（现在的领导者）会发生什么
>失败），没有追踪者的日志完全承诺，
>选举中会发生什么？

如果大多数服务器是活着的，可以互相交谈，他们会的
选举领袖 当选的领导人将至少有一个日志
作为任何服务器的大多数（这是选举）
限制在5.4.1节）。

当选的领导人可能在其日志末尾没有条目
也许是因为它是唯一接收这些服务器的服务器
前任领导人的作品。它会将这些条目发送到
追随者，条目将一旦领先，就会承诺
设法从当前条款提交条目（见第5.4.2节）。


>如果追随者包含日志条目，它是否一直是真的
>当前领导者没有，那么这些日志条目将最终
>删除？我的理解是，领导人将试图强迫
>跟随者的日志符合领导者的日志，
>因此不同意的条目将被删除。

是的 - 一旦新的领导人提交任何新的条目，它会告诉
追随者与冲突的日志条目删除它们并替换它们
与领导的日志条目。

>但是，这些条目被删除会发生什么？他们不是
对客户端来说重要，系统不会通知客户端
有些行动没有发生吗？

跟随者将删除日志条目，而不通知客户端。
这是可以的原因是条目不能被提交，
所以任何客户端发送RPC的领导都无法答复
RPC，所以客户端不能想到RPC已经成功了。
客户可能会超时并将RPC重新发送给新的领导者。


>在图7中，机器如何获得额外的未提交日志？

示例：server（c）的最后一个日志条目未提交，因为它没有
存在于大多数服务器上。这可能是因为服务器
（c）是第6任领导人，失去领导（可能坠毁）
在发送自己的日志索引11的命令之后，然后发送给它
任何其他服务器。

>可能没有一个节点满足选举
>限制 即它们有所有提交的日志？我猜这是真的
>因为3个属性保证追随者只能接受一个
>新日志，当其所有以前的条目是正确的和最新的
>进入只会在至少有一个“好”
追随者？我不确定。

我相信至少有一个服务器总是满足“至少as”
最新的“规则，在5.4.1结尾处描述
服务器可能具有相同的最后一个术语和相同的长度日志，但是
他们中的任何一个被允许成为领导者。

>此外，当网络分区发生时，如果客户端发送命令
>不同的网络，不会将这些发送到“较小的”组
失去了 我想在这种情况下，客户端可以再次发送，但我不是
>肯定

是的，只有具有大多数服务器的分区才能提交
执行客户端操作。少数分区中的服务器
将无法提交客户端操作，所以他们不会回复客户端
要求。客户端将继续重新发送请求，直到可以
联系多数筏分区。

>即使作者的解释，我仍然不能完全理解
>图7.在这种情况下，服务器（c）和服务器（c）为什么会这样
>额外的未提交的条目？

服务器（c）的最后一个日志条目未提交，因为它不存在
大多数服务器。这可能是因为server（c）是
第6期的领导，发送后失去领导（可能坠毁）
本身就是一个日志索引11的命令，但是在发送给任何其他命令之前
服务器。

服务器（d）可能是第7期的领导者，并发送了两个
命令（日志索引11和12），但在发送之前崩溃
命令到任何其他服务器。

>关于“多数”的另一个问题：
>
>例如，有8台服务器。服务器A是当选的领导者
>其余都是追随者。如果发生了坏事，还有4个追随者
>坠毁，他们永远不会发回信息给领导。有
>只剩下3个追随者。我认为即使是服务器A得到所有的响应
>从剩下的追随者中，数量还是少于
>“多数”。以这种方式，领导者永远不会得到更多的承诺
>条目。在这种情况下，筏子能做什么？

至少大多数（5个服务器）必须存活并进行通信
命令Raft进步。在这种情况下，4台实时服务器
将一再企图选举一名领导人，并且失败（因为没有
多数）。如果4个崩溃的服务器中的一个重新启动，那么会有一个
多数，筏筏将选举领导者并继续。

>第1至5节看起来非常直观。我们怎么知道
> 5.4.3完全证明属性的论据 好像合乎逻辑
>但我不知道这个论点是否涵盖了所有可能的情况。

5.4.3不是一个完整的证明。这里有一些地方去寻找东西
更完整：

http://ramcloud.stanford.edu/~ongaro/thesis.pdf
http://verdi.uwplse.org/raft-proof.pdf