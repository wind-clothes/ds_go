## Zookeeper

阅读：“ZooKeeper：互联网规模系统的等待协调”，Patrick
Hunt，Mahadev Konar Flavio P. Junqueira Benjamin Reed。2010年诉讼
USENIX年会技术会议

我们为什么看这篇文章？
  广泛应用的复制状态机服务
    在Yahoo!里面 和雅虎外面
  给予Paxos / ZAP / Raft库的建筑复制服务的案例研究
    类似的问题在实验3中出现
  API支持广泛的用例
    需要容错“主”的应用程序不需要自己滚动
    Zookeeper是通用的，他们应该能够使用zookeeper
  高性能
    不像实验室3的复制键/值服务

Apache Zookeeper
  在雅虎开发 供内部使用
    灵感来自Chubby（Google的全球锁定服务）
    由雅虎使用 YMB和履带
  开源
    作为Hadoop（MapReduce）的一部分
    它现在拥有Apache项目
  广泛应用于公司
  通常在我们阅读的论文中提到

表现：lab3
  由Raft主宰
  考虑一个3节点筏
  在返回客户之前，Raft执行
    领导者坚持日志条目
    领导同时发送消息给追随者
      每个跟随者持续日志条目
      每个追随者回应
  - > 2个磁盘写入和一次往返
    如果磁盘：2 * 10msec = 50 msg / sec
    如果SSD：2 * 2msec + 1msec = 200 msg / sec
  Zookeeper执行21,000 msg / sec
    异步调用
    允许流水线

动机：许多应用需要“主”服务
  示例：GFS
    master拥有每个块的块服务器列表
    主人决定哪个组块服务器是主要的
    等等
    注意：HDFS以有限的方式使用zookeeper
      高清机顶
  其他例子：YMB，履带式等
    YMB需要掌握分片的主题
    抓取工具需要掌握页面提取的命令
     （例如，有一点像在mapreduce的主人）
  主服务通常用于“协调”

一个解决方案：为每个应用程序开发主
  好的，如果写大师不复杂
  但是，主人常常需要：
    容错
      每个应用程序都显示如何使用筏子？
    高性能
      每个应用程序都会显示如何使“读取”操作快速？
  一些应用程序可以解决单点故障
    例如，GFS和MapReduce
    不理想

Zookeeper：通用的“主”服务
  设计挑战：
    什么API？
    新新新旗新新新旗新新新旗新新新旗新新旗2001-新新新新旗新新旗新新旗
    如何获得良好的表现？
  挑战互动
    良好的性能可能会影响API
    例如，允许流水线化的异步接口

Zookeeper API概述
  复制状态机
    几个服务器实现服务
    操作以全局顺序执行
      有一些例外，如果一致性不重要
  复制对象是：znodes
    znodes的层次结构
      以路径名命名
    znodes包含*元数据的应用程序
      配置信息
        参与应用程序的机器
        哪个机器是主要的
      时间戳
      版本号
    znode类型：
      定期
      empheral
      顺序：名称+ seqno
        如果n是新的znode，p是父znode，那么序列
        n的值不会小于任何其他值的值
        在p下创建的顺序znode。

  会议
    客户登录动物园管理员
    会话允许客户端故障切换到另一个zookeeper服务
      客户知道上次完成操作的期限和索引
      发送每个请求
        服务只有在客户看到的时候才能执行操作
    会话可以超时
      客户端必须持续刷新会话
        向服务器发送心跳（如租约）
      如果没有听到客户端的话，动物园认为客户端“死了”
      客户端可能会继续做它的事情（例如，网络分区）
        但在该会话中无法执行其他动物园管理员操作

znodes上的操作
  创建（路径，数据，标志）
  删除（路径，版本）
      如果znode.version = version，那么删除
  存在（路径，手表）
  getData（路径，观察）
  200新新新新新旗新新旗新新新旗新新旗新新旗新新旗新新旗新新旗新新旗新新旗新新旗新新
    如果znode.version = version，那么更新
  getChildren（路径，看）
  同步（）
   以上操作是*异步*
   X- 20045 X- 20045 X-4545 CEEC X-
   sync等待直到所有以前的操作都被“传播”

订购保证
  所有写入操作都是完全有序的
    如果zookeeper执行写操作，则后来其他客户端的写入就可以看到它
    例如，两个客户端创建一个znode，zookeeper以一定的顺序执行它们
  所有操作都是按客户端FIFO排序的
  影响：
    读取观察到来自同一客户端的较早写入的结果
    读取观察写入的一些前缀，也许不包括最近的写入
      - >读取可以返回陈旧的数据
    如果读取观察到某些前缀的写入，则稍后的读取也会观察该前缀

示例就绪znode：
  发生故障
  主要将一串写入文件发送到Zookeeper
    W1 ... Wn C（准备好）
  最终写更新准备好znode
    - >所有以前的写入都可见
  最终写入会导致手表在备份时关闭
    备份问题R（就绪）R1 ... Rn
    然而，它会观察所有写入，因为zookeeper会延迟读取直到
      节点已经看到观察到的所有txn
  假设在R1 .. Rn期间发生故障，说在返回Rj到客户端之后
    主要删除就绪文件 - >手表关闭
    监视警报发送给客户端
    客户知道它必须发出新的R（就绪）R1 ... Rn
  尼斯物业：高性能
    管道写入和读取
    可以从* any * zookeeper节点读取

示例用法1：缓慢锁定
  获取锁定
   重试：
     r = create（“app / lock”，“”，empheral）
     如果r：
       返回
     其他：
       getData（“app / lock”，watch = True）

    watch_event：
       goto重试
      
  释放锁：（自愿或会话超时）
    删除（“应用程序/锁定”）

示例用法2：“ticket”锁
  获取锁定
     n = create（“app / lock / request-”，“”，empheral | sequential）
   重试：
     requests = getChildren（l，false）
     如果n是请求中最低的znode：
       返回
     p =“request-％d”％n  -  1
     如果存在（p，watch = True）
       goto重试

    watch_event：
       goto重试
    
  问：可以在锁定之前看守，这是客户的轮到
  答：是的
     lock / request-10 < - 当前锁定夹
     lock / request-11 < - 下一个
     lock / request-12 < - 我的请求

     如果与请求-11相关联的客户在获得锁定之前就会死机
     手表甚至会发火，但现在还不行。

使用锁
  不直接：失败可能导致您的锁被撤销
    客户端1获取锁
      开始做它的东西
      网络分区
      zookeeper声明客户端1（但不是）
    客户端2获取锁
  在某些情况下，锁是性能优化
    例如，客户端1锁定爬网某些URL
    客户现在会做2，但是没关系
  对于其他情况，锁是一个积木
    例如，应用程序使用它来构建事务
    交易是全无
    我们将在Frangipani论文中看到一个例子

Zookeeper简化了构建应用程序，但不是端到端的解决方案
  大量的难题留给应用程序处理
  考虑在GFS中使用Zookeeper
    即，用Zookeeper替换主人
  应用程序/ GFS仍然需要GFS的所有其他部分
    块的主/备份计划
    块上的版本号
    用于处理主故障切换的协议
    等等
  使用Zookeeper，至少master是容错的
    而且，不会遇到分脑问题
    即使它已经复制了服务器

实施概述
  类似于实验3（见上一讲）
  两层：
    动物园管理员服务（K / V服务）
    X-
  Start（）在底层插入ops
  一段时间后，ops从每个副本的底层弹出
    这些操作按照它们弹出的顺序进行
    在实验室3的应用渠道
    在ZAB中的abdeliver（）upcall

挑战：复制客户端请求
  脚本
    主要接收客户端请求，失败
    客户端重新发送客户端请求到新的主要
  实验3： 
    表检测重复
    限制：每个客户一个优秀的
    问题问题：无法管理客户端请求
  动物园管理员：
    一些操作是幂等的时期
    一些操作很容易造成幂等
      测试版本和随后-DO-OP
    有些操作客户端负责检测重复
     考虑锁定示例。
       创建一个包含其会话ID的文件名
         “应用程序/锁定/请求的sessionID-的SeqNo”
         动物管理员在切换到新的主要时通知客户端
     客户端运行getChildren（）
       如果有新的请求，全部设置
       如果没有，重新发布创建

挑战：阅读操作
  许多操作都是读操作
    它们不会修改复制状态
  他们必须经过ZAB / Raft吗？
  任何副本都可以执行读操作吗？
  如果阅读操作通过Raft，性能会很慢

问题：如果只有主人执行，读取可能会返回过期的数据
  主要可能不知道它不是主要的
    网络分区使另一个节点成为主节点
    该分区可能已经处理了写入操作
  如果旧的主服务器进行读取操作，则不会看到这些写操作
   =>读取返回陈旧的数据

Zookeeper解决方案：不承诺不过时的数据（默认情况下）
  读取被允许返回陈旧的数据
    读取可以由任何副本执行
    读取吞吐量随着服务器数量的增加而增加
    读取返回它看到的最后一个zxid
     所以新服务器在服务阅读之前可以赶上zxid
     避免从过去的阅读
  只有sync-read（）保证数据不会过时

同步优化：避免ZAB层进行同步读取
  必须确保读取观察最后一次提交txn
  领导者同步在它和副本之间的队列中
    如果在队列提前执行，则领导者必须是领导者
    否则，发出空事务
  在同样的精神中阅读筏纸中的优化
    见最后一节第8节筏纸

表现（见表1）
  阅读价格便宜
    问：为什么服务器增加读数？
  写得贵
    问：为什么随着服务器数量的增加而减慢？
  快速故障恢复（图8）
    即使在失败发生的时候，体面的体面

参考文献：
  ZAB：http://dl.acm.org/citation.cfm?id=2056409
  https://zookeeper.apache.org/
  https://cs.brown.edu/~mph/Herlihy91/p124-herlihy.pdf（等待免费，普遍物件等）